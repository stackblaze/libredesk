package main

import (
	"strings"
	"time"

	amodels "github.com/abhinavxd/libredesk/internal/auth/models"
	cmodels "github.com/abhinavxd/libredesk/internal/conversation/models"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/zerodha/fastglue"
)

type aiCompletionReq struct {
	PromptKey string `json:"prompt_key"`
	Content   string `json:"content"`
}

type aiDraftReplyReq struct {
	ConversationUUID string `json:"conversation_uuid"`
	Instructions     string `json:"instructions"`
}

type providerUpdateReq struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

// handleAICompletion handles AI completion requests
func handleAICompletion(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req = aiCompletionReq{}
	)

	if err := r.Decode(&req, "json"); err != nil {
		return sendErrorEnvelope(r, envelope.NewError(envelope.InputError, app.i18n.T("errors.parsingRequest"), nil))
	}

	resp, err := app.ai.Completion(req.PromptKey, req.Content)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}

// handleAIDraftReply drafts a customer reply from conversation context and agent instructions.
func handleAIDraftReply(r *fastglue.Request) error {
	var (
		app   = r.Context.(*App)
		req   = aiDraftReplyReq{}
		auser = r.RequestCtx.UserValue("user").(amodels.User)
	)

	if err := r.Decode(&req, "json"); err != nil {
		return sendErrorEnvelope(r, envelope.NewError(envelope.InputError, app.i18n.T("errors.parsingRequest"), nil))
	}

	req.ConversationUUID = strings.TrimSpace(req.ConversationUUID)
	req.Instructions = strings.TrimSpace(req.Instructions)
	if req.ConversationUUID == "" || req.Instructions == "" {
		return sendErrorEnvelope(r, envelope.NewError(envelope.InputError, app.i18n.T("ai.draftReplyInstructionsRequired"), nil))
	}

	user, err := app.user.GetAgentCachedOrLoad(auser.ID)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}

	conversation, err := enforceConversationAccess(app, req.ConversationUUID, user)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}

	private := false
	messages, err := app.conversation.GetAllConversationMessages(req.ConversationUUID, &private, []string{cmodels.MessageIncoming, cmodels.MessageOutgoing})
	if err != nil {
		return sendErrorEnvelope(r, err)
	}

	transcript := string(app.conversation.BuildTranscript(*conversation, messages, time.Now()))
	resp, err := app.ai.DraftReply(transcript, req.Instructions)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}

// handleGetAIPrompts returns AI prompts
func handleGetAIPrompts(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
	)
	resp, err := app.ai.GetPrompts()
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}

// handleGetAIProvider returns AI provider settings for admin UI.
func handleGetAIProvider(r *fastglue.Request) error {
	app := r.Context.(*App)
	resp, err := app.ai.GetProviderSettings()
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}

// handleUpdateAIProvider updates the AI provider
func handleUpdateAIProvider(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req providerUpdateReq
	)
	if err := r.Decode(&req, "json"); err != nil {
		return sendErrorEnvelope(r, envelope.NewError(envelope.InputError, app.i18n.T("errors.parsingRequest"), nil))
	}
	if err := app.ai.UpdateProvider(req.Provider, req.APIKey, req.Model); err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope("Provider updated successfully")
}
