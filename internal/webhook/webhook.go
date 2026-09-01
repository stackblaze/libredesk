// Package webhook handles the management of webhooks and webhook deliveries.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/abhinavxd/libredesk/internal/crypto"
	"github.com/abhinavxd/libredesk/internal/dbutil"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/ssrf"
	"github.com/abhinavxd/libredesk/internal/stringutil"
	"github.com/abhinavxd/libredesk/internal/version"
	"github.com/abhinavxd/libredesk/internal/webhook/models"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-i18n"
	"github.com/lib/pq"
	"github.com/zerodha/logf"
)

var (
	//go:embed queries.sql
	efs embed.FS
)

// Manager handles webhook-related operations.
type Manager struct {
	q             queries
	lo            *logf.Logger
	i18n          *i18n.I18n
	db            *sqlx.DB
	deliveryQueue chan DeliveryTask
	httpClient    *http.Client
	workers       int
	closed        bool
	closedMu      sync.RWMutex
	wg            sync.WaitGroup
	encryptionKey string
	rootURL       func() string
	threadMu      sync.Map
}

// Opts contains options for initializing the Manager.
type Opts struct {
	DB            *sqlx.DB
	Lo            *logf.Logger
	I18n          *i18n.I18n
	Workers       int
	QueueSize     int
	Timeout       time.Duration
	EncryptionKey string
	DialControl   ssrf.Control
	RootURL       func() string
}

// DeliveryTask represents a webhook delivery task
type DeliveryTask struct {
	Event   models.WebhookEvent
	Payload any
	// Zero means fan out to all subscribers of Event.
	WebhookID int
}

// queries contains prepared SQL queries.
type queries struct {
	GetWebhooksCompact  *sqlx.Stmt `query:"get-webhooks-compact"`
	GetAllWebhooks      *sqlx.Stmt `query:"get-all-webhooks"`
	GetWebhook          *sqlx.Stmt `query:"get-webhook"`
	GetWebhookSecret    *sqlx.Stmt `query:"get-webhook-secret"`
	GetActiveWebhooks   *sqlx.Stmt `query:"get-active-webhooks"`
	GetWebhooksByEvent  *sqlx.Stmt `query:"get-webhooks-by-event"`
	GetDiscordThread    *sqlx.Stmt `query:"get-discord-thread"`
	UpsertDiscordThread *sqlx.Stmt `query:"upsert-discord-thread"`
	InsertWebhook       *sqlx.Stmt `query:"insert-webhook"`
	UpdateWebhook       *sqlx.Stmt `query:"update-webhook"`
	DeleteWebhook       *sqlx.Stmt `query:"delete-webhook"`
	ToggleWebhook       *sqlx.Stmt `query:"toggle-webhook"`
}

// New creates and returns a new instance of the Manager.
func New(opts Opts) (*Manager, error) {
	var q queries

	if _, err := opts.DB.Exec(`
		CREATE TABLE IF NOT EXISTS webhook_discord_threads (
			webhook_id INT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
			conversation_uuid TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (webhook_id, conversation_uuid)
		);
	`); err != nil {
		return nil, fmt.Errorf("creating webhook_discord_threads: %w", err)
	}

	if err := dbutil.ScanSQLFile("queries.sql", &q, opts.DB, efs); err != nil {
		return nil, err
	}

	transport := ssrf.NewTransport(opts.DialControl, 3*time.Second)
	transport.TLSHandshakeTimeout = 3 * time.Second
	transport.ResponseHeaderTimeout = 3 * time.Second

	return &Manager{
		q:             q,
		lo:            opts.Lo,
		i18n:          opts.I18n,
		db:            opts.DB,
		deliveryQueue: make(chan DeliveryTask, opts.QueueSize),
		httpClient: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
		workers:       opts.Workers,
		encryptionKey: opts.EncryptionKey,
		rootURL:       opts.RootURL,
	}, nil
}

// GetAllCompact retrieves all webhooks with only id and name.
func (m *Manager) GetAllCompact() ([]models.WebhookCompact, error) {
	var webhooks = make([]models.WebhookCompact, 0)
	if err := m.q.GetWebhooksCompact.Select(&webhooks); err != nil {
		m.lo.Error("error fetching webhooks", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return webhooks, nil
}

// GetAll retrieves all webhooks.
func (m *Manager) GetAll() ([]models.Webhook, error) {
	var webhooks = make([]models.Webhook, 0)
	if err := m.q.GetAllWebhooks.Select(&webhooks); err != nil {
		m.lo.Error("error fetching webhooks", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	// Decrypt secrets
	m.decryptWebhooks(webhooks)

	return webhooks, nil
}

// Get retrieves a webhook by ID.
func (m *Manager) Get(id int) (models.Webhook, error) {
	var webhook models.Webhook
	if err := m.q.GetWebhook.Get(&webhook, id); err != nil {
		if err == sql.ErrNoRows {
			return webhook, envelope.NewError(envelope.NotFoundError, m.i18n.T("globals.messages.notFound"), nil)
		}
		m.lo.Error("error fetching webhook", "error", err)
		return webhook, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	// Decrypt secret
	if err := m.decryptWebhook(&webhook); err != nil {
		return webhook, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	return webhook, nil
}

// Create creates a new webhook.
func (m *Manager) Create(webhook models.Webhook) (models.Webhook, error) {
	var result models.Webhook

	// Encrypt secret before storing
	encryptedSecret, err := m.encryptSecret(webhook.Secret)
	if err != nil {
		return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	delivery, err := normalizeDelivery(webhook.URL, webhook.Delivery)
	if err != nil {
		return models.Webhook{}, envelope.NewError(envelope.InputError, m.i18n.T("admin.webhook.invalidDiscordURL"), nil)
	}

	if err := m.q.InsertWebhook.Get(&result, webhook.Name, webhook.URL, pq.Array(webhook.Events), encryptedSecret, webhook.IsActive, delivery, normalizeIDs(webhook.InboxIDs), normalizeIDs(webhook.TeamIDs), normalizeIDs(webhook.UserIDs)); err != nil {
		if dbutil.IsUniqueViolationError(err) {
			return models.Webhook{}, envelope.NewError(envelope.ConflictError, m.i18n.T("globals.messages.errorAlreadyExists"), nil)
		}
		m.lo.Error("error inserting webhook", "error", err)
		return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	// Decrypt secret before returning (ignore errors as non-critical)
	if err := m.decryptWebhook(&result); err != nil {
		m.lo.Error("error decrypting webhook secret after creation", "webhook_id", result.ID, "error", err)
	}

	return result, nil
}

// Update updates a webhook by ID.
func (m *Manager) Update(id int, webhook models.Webhook) (models.Webhook, error) {
	var result models.Webhook

	// Preserve the existing encrypted secret.
	encryptedSecret := webhook.Secret
	if strings.Contains(webhook.Secret, stringutil.PasswordDummy) {
		var existingSecret string
		if err := m.q.GetWebhookSecret.Get(&existingSecret, id); err != nil {
			m.lo.Error("error fetching existing webhook secret", "id", id, "error", err)
			return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
		encryptedSecret = existingSecret
	} else if !crypto.IsEncrypted(webhook.Secret) {
		// Encrypt new secret before storing
		var err error
		encryptedSecret, err = m.encryptSecret(webhook.Secret)
		if err != nil {
			return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
		}
	}

	delivery, err := normalizeDelivery(webhook.URL, webhook.Delivery)
	if err != nil {
		return models.Webhook{}, envelope.NewError(envelope.InputError, m.i18n.T("admin.webhook.invalidDiscordURL"), nil)
	}

	if err := m.q.UpdateWebhook.Get(&result, id, webhook.Name, webhook.URL, pq.Array(webhook.Events), encryptedSecret, webhook.IsActive, delivery, normalizeIDs(webhook.InboxIDs), normalizeIDs(webhook.TeamIDs), normalizeIDs(webhook.UserIDs)); err != nil {
		m.lo.Error("error updating webhook", "error", err)
		return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}

	// Decrypt secret before returning (ignore errors as non-critical)
	if err := m.decryptWebhook(&result); err != nil {
		m.lo.Error("error decrypting webhook secret after update", "webhook_id", result.ID, "error", err)
	}

	return result, nil
}

// Delete deletes a webhook by ID.
func (m *Manager) Delete(id int) error {
	if _, err := m.q.DeleteWebhook.Exec(id); err != nil {
		m.lo.Error("error deleting webhook", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return nil
}

// Toggle toggles the active status of a webhook by ID.
func (m *Manager) Toggle(id int) (models.Webhook, error) {
	var result models.Webhook
	if err := m.q.ToggleWebhook.Get(&result, id); err != nil {
		m.lo.Error("error toggling webhook", "error", err)
		return models.Webhook{}, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return result, nil
}

// SendTestWebhook sends a test webhook to the specified webhook ID.
func (m *Manager) SendTestWebhook(id int) error {
	webhook, err := m.Get(id)
	if err != nil {
		return envelope.NewError(envelope.NotFoundError, m.i18n.T("globals.messages.notFound"), nil)
	}

	m.deliverSingleWebhook(webhook, DeliveryTask{
		Event: models.EventWebhookTest,
		Payload: map[string]any{
			"id":   webhook.ID,
			"name": webhook.Name,
		},
	})

	return nil
}

// TriggerEvent triggers webhooks for a specific event with the provided data.
func (m *Manager) TriggerEvent(event models.WebhookEvent, data any) {
	m.closedMu.RLock()
	defer m.closedMu.RUnlock()
	if m.closed {
		return
	}

	select {
	case m.deliveryQueue <- DeliveryTask{
		Event:   event,
		Payload: data,
	}:
	default:
		m.lo.Warn("webhook delivery queue is full, dropping webhook delivery", "event", event, "queue_size", len(m.deliveryQueue))
	}
}

// TriggerWebhook enqueues a delivery of the given event to one specific webhook.
func (m *Manager) TriggerWebhook(webhookID int, event models.WebhookEvent, data any) {
	// A non-positive ID would be treated as a fan out to every subscriber of the event.
	if webhookID <= 0 {
		m.lo.Warn("dropping targeted webhook delivery, webhook ID is not positive", "webhook_id", webhookID, "event", event)
		return
	}

	m.closedMu.RLock()
	defer m.closedMu.RUnlock()
	if m.closed {
		return
	}

	select {
	case m.deliveryQueue <- DeliveryTask{
		Event:     event,
		Payload:   data,
		WebhookID: webhookID,
	}:
	default:
		m.lo.Warn("webhook delivery queue is full, dropping webhook delivery", "event", event, "webhook_id", webhookID, "queue_size", len(m.deliveryQueue))
	}
}

// Run starts the webhook delivery worker pool.
func (m *Manager) Run(ctx context.Context) {
	for i := 0; i < m.workers; i++ {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.worker(ctx)
		}()
	}
}

// Close signals the manager to stop processing and waits for all workers to finish.
func (m *Manager) Close() {
	m.closedMu.Lock()
	defer m.closedMu.Unlock()
	if m.closed {
		return
	}
	m.closed = true
	close(m.deliveryQueue)
	m.wg.Wait()
}

// worker processes webhook delivery tasks from the queue.
func (m *Manager) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-m.deliveryQueue:
			if !ok {
				return
			}
			m.deliverWebhook(task)
		}
	}
}

// deliverWebhook delivers webhooks for an event by making HTTP requests.
func (m *Manager) deliverWebhook(task DeliveryTask) {
	if task.WebhookID > 0 {
		webhook, err := m.Get(task.WebhookID)
		if err != nil {
			m.lo.Error("error fetching webhook for delivery", "webhook_id", task.WebhookID, "event", task.Event, "error", err)
			return
		}
		if !webhook.IsActive {
			m.lo.Debug("skipping delivery, webhook is inactive", "webhook_id", webhook.ID, "event", task.Event)
			return
		}
		m.deliverSingleWebhook(webhook, task)
		return
	}

	webhooks, err := m.getWebhooksByEvent(string(task.Event))
	if err != nil {
		m.lo.Error("error fetching webhooks for event", "event", task.Event, "error", err)
		return
	}

	for _, webhook := range webhooks {
		if !matchesWebhookFilters(webhook, task) {
			m.lo.Debug("skipping webhook, event does not match inbox/team/user filters",
				"webhook_id", webhook.ID, "event", task.Event)
			continue
		}
		m.deliverSingleWebhook(webhook, task)
	}
}

// deliverSingleWebhook delivers a webhook to a single endpoint.
func (m *Manager) deliverSingleWebhook(webhook models.Webhook, task DeliveryTask) {
	if usesDiscordPayload(webhook) {
		m.deliverDiscordWebhook(webhook, task)
		return
	}

	payloadBytes, err := json.Marshal(map[string]any{
		"event":     task.Event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"payload":   task.Payload,
	})
	if err != nil {
		m.lo.Error("error marshaling webhook payload", "webhook_id", webhook.ID, "event", task.Event, "error", err)
		return
	}
	m.postWebhook(webhook, task, webhook.URL, payloadBytes)
}

func (m *Manager) deliverDiscordWebhook(webhook models.Webhook, task DeliveryTask) {
	convUUID := conversationUUIDFromTask(task)
	threadID := ""
	threadName := ""
	wait := false

	if task.Event != models.EventWebhookTest && convUUID != "" {
		mu := m.threadLock(webhook.ID, convUUID)
		mu.Lock()
		defer mu.Unlock()

		threadID = m.discordThreadID(webhook.ID, convUUID)
		if threadID == "" {
			threadName = discordThreadName(task)
			wait = true
		}
	}

	payloadBytes, err := buildDiscordPayload(task, m.appRootURL(), threadName)
	if err != nil {
		m.lo.Error("error marshaling webhook payload", "webhook_id", webhook.ID, "event", task.Event, "error", err)
		return
	}

	execURL := discordExecuteURL(webhook.URL, threadID, wait)
	status, body, ok := m.postWebhook(webhook, task, execURL, payloadBytes)
	if !ok && wait && threadName != "" && status >= 400 {
		// Text channels reject thread_name; post into the parent channel instead.
		payloadBytes, err = buildDiscordPayload(task, m.appRootURL(), "")
		if err != nil {
			return
		}
		_, body, ok = m.postWebhook(webhook, task, webhook.URL, payloadBytes)
	}
	if ok && wait && threadID == "" {
		if id := discordThreadIDFromResponse(body); id != "" {
			m.saveDiscordThread(webhook.ID, convUUID, id)
		}
	}
}

func (m *Manager) threadLock(webhookID int, convUUID string) *sync.Mutex {
	key := fmt.Sprintf("%d:%s", webhookID, convUUID)
	actual, _ := m.threadMu.LoadOrStore(key, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (m *Manager) discordThreadID(webhookID int, convUUID string) string {
	var id string
	if err := m.q.GetDiscordThread.Get(&id, webhookID, convUUID); err != nil {
		if err != sql.ErrNoRows {
			m.lo.Error("error fetching discord thread", "webhook_id", webhookID, "conversation_uuid", convUUID, "error", err)
		}
		return ""
	}
	return id
}

func (m *Manager) saveDiscordThread(webhookID int, convUUID, threadID string) {
	if webhookID <= 0 || convUUID == "" || threadID == "" {
		return
	}
	if _, err := m.q.UpsertDiscordThread.Exec(webhookID, convUUID, threadID); err != nil {
		m.lo.Error("error saving discord thread", "webhook_id", webhookID, "conversation_uuid", convUUID, "error", err)
	}
}

func (m *Manager) postWebhook(webhook models.Webhook, task DeliveryTask, destURL string, payloadBytes []byte) (int, []byte, bool) {
	req, err := http.NewRequest("POST", destURL, bytes.NewReader(payloadBytes))
	if err != nil {
		m.lo.Error("error creating webhook request", "webhook_id", webhook.ID, "url", destURL, "event", task.Event, "error", err)
		return 0, nil, false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Libredesk-Webhook/"+version.Version)

	if webhook.Secret != "" && !usesDiscordPayload(webhook) {
		signature := m.generateSignature(payloadBytes, webhook.Secret)
		req.Header.Set("X-Libredesk-Signature", signature)
	}

	m.lo.Debug("delivering webhook",
		"webhook_id", webhook.ID,
		"url", destURL,
		"event", task.Event,
		"payload", string(payloadBytes),
		"headers", req.Header,
	)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.lo.Error("webhook delivery failed - HTTP request error",
			"webhook_id", webhook.ID,
			"url", destURL,
			"event", task.Event,
			"error", err)
		return 0, nil, false
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		m.lo.Error("error reading webhook response", "webhook_id", webhook.ID, "error", err)
		responseBody = []byte(fmt.Sprintf("Error reading response: %v", err))
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if success {
		m.lo.Info("webhook delivered successfully",
			"webhook_id", webhook.ID,
			"event", task.Event,
			"url", destURL,
			"status_code", resp.StatusCode)
	} else {
		m.lo.Error("webhook delivery failed",
			"webhook_id", webhook.ID,
			"event", task.Event,
			"url", destURL,
			"status_code", resp.StatusCode,
			"response", string(responseBody))
	}
	return resp.StatusCode, responseBody, success
}

func (m *Manager) appRootURL() string {
	if m.rootURL == nil {
		return ""
	}
	return m.rootURL()
}

// generateSignature generates HMAC-SHA256 signature for webhook payload.
func (m *Manager) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func normalizeIDs(ids pq.Int64Array) pq.Int64Array {
	if len(ids) == 0 {
		return pq.Int64Array{}
	}
	out := make(pq.Int64Array, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// getWebhooksByEvent retrieves active webhooks that are subscribed to a specific event.
func (m *Manager) getWebhooksByEvent(event string) ([]models.Webhook, error) {
	var webhooks = make([]models.Webhook, 0)
	if err := m.q.GetWebhooksByEvent.Select(&webhooks, event); err != nil {
		return nil, err
	}

	// Decrypt secrets
	m.decryptWebhooks(webhooks)

	return webhooks, nil
}
