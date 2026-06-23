// Package ai manages AI prompts and integrates with LLM providers.
package ai

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/abhinavxd/libredesk/internal/ai/models"
	"github.com/abhinavxd/libredesk/internal/crypto"
	"github.com/abhinavxd/libredesk/internal/dbutil"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-i18n"
	"github.com/zerodha/logf"
)

var (
	//go:embed queries.sql
	efs embed.FS

	ErrInvalidAPIKey = errors.New("invalid API Key")
	ErrApiKeyNotSet  = errors.New("api Key not set")

	DefaultOpenAIModel = "gpt-4o-mini"
)

type Manager struct {
	q             queries
	lo            *logf.Logger
	i18n          *i18n.I18n
	encryptionKey string
}

// Opts contains options for initializing the Manager.
type Opts struct {
	DB            *sqlx.DB
	I18n          *i18n.I18n
	Lo            *logf.Logger
	EncryptionKey string
}

// queries contains prepared SQL queries.
type queries struct {
	GetDefaultProvider          *sqlx.Stmt `query:"get-default-provider"`
	GetPrompt                   *sqlx.Stmt `query:"get-prompt"`
	GetPrompts                  *sqlx.Stmt `query:"get-prompts"`
	SetOpenAIKey                *sqlx.Stmt `query:"set-openai-key"`
	UpdateDefaultProviderConfig *sqlx.Stmt `query:"update-default-provider-config"`
}

// New creates and returns a new instance of the Manager.
func New(opts Opts) (*Manager, error) {
	var q queries
	if err := dbutil.ScanSQLFile("queries.sql", &q, opts.DB, efs); err != nil {
		return nil, err
	}
	return &Manager{
		q:             q,
		lo:            opts.Lo,
		i18n:          opts.I18n,
		encryptionKey: opts.EncryptionKey,
	}, nil
}

// Completion sends a prompt to the default provider and returns the response.
func (m *Manager) Completion(k string, prompt string) (string, error) {
	systemPrompt, err := m.getPrompt(k)
	if err != nil {
		return "", err
	}

	client, err := m.getDefaultProviderClient()
	if err != nil {
		m.lo.Error("error getting provider client", "error", err)
		return "", envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	payload := PromptPayload{
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
	}

	response, err := client.SendPrompt(payload)
	if err != nil {
		if errors.Is(err, ErrInvalidAPIKey) {
			m.lo.Error("error invalid API key", "error", err)
			return "", envelope.NewError(envelope.InputError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
		if errors.Is(err, ErrApiKeyNotSet) {
			m.lo.Error("error API key not set", "error", err)
			return "", envelope.NewError(envelope.InputError, m.i18n.Ts("ai.apiKeyNotSet", "provider", "OpenAI"), nil)
		}
		m.lo.Error("error sending prompt to provider", "error", err)
		return "", envelope.NewError(envelope.GeneralError, err.Error(), nil)
	}

	return response, nil
}

const draftReplySystemPrompt = `You help customer support agents draft replies to customers.

Rules:
- Use ONLY information from the conversation transcript and the agent's instructions.
- Do NOT invent policies, refunds, timelines, order details, or promises not supported by the transcript or instructions.
- If the instructions ask for something the transcript does not support, stay factual and avoid confirming unsupported claims.
- Write a clear, professional, customer-facing reply in plain language.
- Return only the reply body text with no subject line, preamble, or meta-commentary.`

// DraftReply generates a customer reply grounded in the conversation transcript and agent instructions.
func (m *Manager) DraftReply(conversationContext, instructions string) (string, error) {
	instructions = strings.TrimSpace(instructions)
	if instructions == "" {
		return "", envelope.NewError(envelope.InputError, m.i18n.T("ai.draftReplyInstructionsRequired"), nil)
	}

	client, err := m.getDefaultProviderClient()
	if err != nil {
		m.lo.Error("error getting provider client", "error", err)
		return "", envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	userPrompt := fmt.Sprintf(`## Conversation transcript
%s

## Agent instructions (what to communicate to the customer)
%s

Draft the customer-facing reply.`, strings.TrimSpace(conversationContext), instructions)

	payload := PromptPayload{
		SystemPrompt: draftReplySystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    1500,
		Temperature:  0.3,
	}

	response, err := client.SendPrompt(payload)
	if err != nil {
		if errors.Is(err, ErrInvalidAPIKey) {
			m.lo.Error("error invalid API key", "error", err)
			return "", envelope.NewError(envelope.InputError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
		if errors.Is(err, ErrApiKeyNotSet) {
			m.lo.Error("error API key not set", "error", err)
			return "", envelope.NewError(envelope.InputError, m.i18n.Ts("ai.apiKeyNotSet", "provider", "OpenAI"), nil)
		}
		m.lo.Error("error sending draft reply prompt to provider", "error", err)
		return "", envelope.NewError(envelope.GeneralError, err.Error(), nil)
	}

	return response, nil
}

// GetPrompts returns a list of prompts from the database.
func (m *Manager) GetPrompts() ([]models.Prompt, error) {
	var prompts = make([]models.Prompt, 0)
	if err := m.q.GetPrompts.Select(&prompts); err != nil {
		m.lo.Error("error fetching prompts", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return prompts, nil
}

// UpdateProvider updates a provider.
func (m *Manager) UpdateProvider(provider, apiKey, model string) error {
	switch ProviderType(provider) {
	case ProviderOpenAI:
		return m.updateOpenAIProvider(apiKey, model)
	default:
		m.lo.Error("unsupported provider type", "provider", provider)
		return envelope.NewError(envelope.GeneralError, m.i18n.T("validation.invalidProvider"), nil)
	}
}

// GetProviderSettings returns the default AI provider settings for admin UI.
func (m *Manager) GetProviderSettings() (models.ProviderSettings, error) {
	var settings models.ProviderSettings

	p, config, err := m.getDefaultOpenAIConfig()
	if err != nil {
		return settings, err
	}

	settings.Provider = p.Provider
	settings.Model = config.model
	if settings.Model == "" {
		settings.Model = DefaultOpenAIModel
	}

	if config.encryptedAPIKey == "" {
		return settings, nil
	}

	decryptedKey, err := crypto.Decrypt(config.encryptedAPIKey, m.encryptionKey)
	if err != nil {
		m.lo.Error("error decrypting API key", "error", err)
		return settings, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	settings.APIKeySet = decryptedKey != ""
	settings.APIKeyHint = maskAPIKey(decryptedKey)
	return settings, nil
}

type openAIStoredConfig struct {
	encryptedAPIKey string
	model           string
}

func (m *Manager) getDefaultOpenAIConfig() (models.Provider, openAIStoredConfig, error) {
	var (
		p      models.Provider
		config openAIStoredConfig
	)

	if err := m.q.GetDefaultProvider.Get(&p); err != nil {
		m.lo.Error("error fetching provider details", "error", err)
		return p, config, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	parsed := struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	}{}
	if err := json.Unmarshal([]byte(p.Config), &parsed); err != nil {
		m.lo.Error("error parsing provider config", "error", err)
		return p, config, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	config.encryptedAPIKey = parsed.APIKey
	config.model = strings.TrimSpace(parsed.Model)
	return p, config, nil
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return strings.Repeat("•", 8)
	}
	return strings.Repeat("•", 8) + key[len(key)-4:]
}

func isMaskedAPIKey(value string) bool {
	return strings.Contains(value, "•")
}

func (m *Manager) updateOpenAIProvider(apiKey, model string) error {
	_, config, err := m.getDefaultOpenAIConfig()
	if err != nil {
		return err
	}

	encryptedKey := config.encryptedAPIKey
	apiKey = strings.TrimSpace(apiKey)
	if apiKey != "" && !isMaskedAPIKey(apiKey) {
		encryptedKey, err = crypto.Encrypt(apiKey, m.encryptionKey)
		if err != nil {
			m.lo.Error("error encrypting API key", "error", err)
			return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
	}

	model = strings.TrimSpace(model)
	if model == "" {
		model = config.model
	}
	if model == "" {
		model = DefaultOpenAIModel
	}

	stored, err := json.Marshal(map[string]string{
		"api_key": encryptedKey,
		"model":   model,
	})
	if err != nil {
		m.lo.Error("error marshalling provider config", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	if _, err := m.q.UpdateDefaultProviderConfig.Exec(string(stored)); err != nil {
		m.lo.Error("error updating provider config", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return nil
}

// setOpenAIAPIKey sets the OpenAI API key in the database.
func (m *Manager) setOpenAIAPIKey(apiKey string) error {
	return m.updateOpenAIProvider(apiKey, "")
}

// getPrompt returns a prompt from the database.
func (m *Manager) getPrompt(k string) (string, error) {
	var p models.Prompt
	if err := m.q.GetPrompt.Get(&p, k); err != nil {
		if err == sql.ErrNoRows {
			m.lo.Error("error prompt not found", "key", k)
			return "", envelope.NewError(envelope.InputError, m.i18n.T("validation.notFoundTemplate"), nil)
		}
		m.lo.Error("error fetching prompt", "error", err)
		return "", envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return p.Content, nil
}

// getDefaultProviderClient returns a ProviderClient for the default provider.
func (m *Manager) getDefaultProviderClient() (ProviderClient, error) {
	var p models.Provider

	if err := m.q.GetDefaultProvider.Get(&p); err != nil {
		m.lo.Error("error fetching provider details", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	switch ProviderType(p.Provider) {
	case ProviderOpenAI:
		config := struct {
			APIKey string `json:"api_key"`
			Model  string `json:"model"`
		}{}
		if err := json.Unmarshal([]byte(p.Config), &config); err != nil {
			m.lo.Error("error parsing provider config", "error", err)
			return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
		// Decrypt API key.
		decryptedKey, err := crypto.Decrypt(config.APIKey, m.encryptionKey)
		if err != nil {
			m.lo.Error("error decrypting API key", "error", err)
			return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
		model := strings.TrimSpace(config.Model)
		if model == "" {
			model = DefaultOpenAIModel
		}
		return NewOpenAIClient(decryptedKey, model, m.lo), nil
	default:
		m.lo.Error("unsupported provider type", "provider", p.Provider)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("validation.invalidProvider"), nil)
	}
}
