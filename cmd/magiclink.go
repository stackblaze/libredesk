package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	amodels "github.com/abhinavxd/libredesk/internal/auth/models"
	"github.com/abhinavxd/libredesk/internal/envelope"
	notifier "github.com/abhinavxd/libredesk/internal/notification"
	"github.com/abhinavxd/libredesk/internal/stringutil"
	tmpl "github.com/abhinavxd/libredesk/internal/template"
	"github.com/abhinavxd/libredesk/internal/user/models"
	realip "github.com/ferluci/fast-realip"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

const (
	magicLinkTTL          = 15 * time.Minute
	magicLinkKeyPrefix    = "magic_link:"
	magicLinkEnabledKey   = "auth.magic_link_enabled"
	magicLinkTokenLength  = 48
	magicLinkEmailSubject = "Your sign-in link"
)

type magicLinkRequest struct {
	Email string `json:"email"`
}

type magicLinkVerifyRequest struct {
	Token string `json:"token"`
}

// magicLinkEnabled reports whether passwordless email sign-in is switched on.
func magicLinkEnabled(app *App) bool {
	b, err := app.setting.Get(magicLinkEnabledKey)
	if err != nil {
		return false
	}
	var enabled bool
	_ = json.Unmarshal(b, &enabled)
	return enabled
}

// magicLinkKey hashes the token so Redis never holds a usable sign-in token.
func magicLinkKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return magicLinkKeyPrefix + hex.EncodeToString(sum[:])
}

// handleRequestMagicLink emails a single-use sign-in link to an agent.
// It always answers with success so the endpoint cannot be used to discover accounts.
func handleRequestMagicLink(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req magicLinkRequest
	)
	if !magicLinkEnabled(app) {
		return r.SendErrorEnvelope(fasthttp.StatusForbidden, app.i18n.T("auth.magicLinkDisabled"), nil, envelope.PermissionError)
	}
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("errors.parsingRequest"), nil, envelope.InputError)
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.Ts("globals.messages.empty", "name", "`email`"), nil, envelope.InputError)
	}

	user, err := app.user.GetAgent(0, email)
	if err != nil || !user.Enabled || user.Type != models.UserTypeAgent {
		app.lo.Info("magic link requested for unknown or disabled agent", "email", email)
		return r.SendEnvelope(true)
	}

	token, err := stringutil.RandomAlphanumeric(magicLinkTokenLength)
	if err != nil {
		app.lo.Error("error generating magic link token", "error", err)
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
	}
	if err := app.redis.Set(context.Background(), magicLinkKey(token), strconv.Itoa(user.ID), magicLinkTTL).Err(); err != nil {
		app.lo.Error("error storing magic link token", "error", err)
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
	}

	content, err := app.tmpl.RenderInMemoryTemplate(tmpl.TmplMagicLink, map[string]any{
		"Token":         token,
		"ExpiryMinutes": int(magicLinkTTL.Minutes()),
	})
	if err != nil {
		app.lo.Error("error rendering magic link template", "error", err)
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("auth.magicLinkSendError"), nil))
	}
	if err := app.notifier.Send(notifier.Message{
		RecipientEmails: []string{user.Email.String},
		Subject:         magicLinkEmailSubject,
		Content:         content,
		Provider:        notifier.ProviderEmail,
	}); err != nil {
		app.lo.Error("error sending magic link email", "error", err)
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("auth.magicLinkSendError"), nil))
	}
	return r.SendEnvelope(true)
}

// handleVerifyMagicLink consumes a sign-in token. Agents with TOTP get a pending token instead of a session.
func handleVerifyMagicLink(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req magicLinkVerifyRequest
	)
	if !magicLinkEnabled(app) {
		return r.SendErrorEnvelope(fasthttp.StatusForbidden, app.i18n.T("auth.magicLinkDisabled"), nil, envelope.PermissionError)
	}
	if err := r.Decode(&req, "json"); err != nil || strings.TrimSpace(req.Token) == "" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("globals.messages.badRequest"), nil, envelope.InputError)
	}

	// GETDEL makes the link single-use even under concurrent clicks.
	val, err := app.redis.GetDel(context.Background(), magicLinkKey(strings.TrimSpace(req.Token))).Result()
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, app.i18n.T("auth.magicLinkInvalid"), nil, envelope.PermissionError)
	}
	userID, err := strconv.Atoi(val)
	if err != nil {
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
	}
	user, err := app.user.GetAgent(userID, "")
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, app.i18n.T("auth.magicLinkInvalid"), nil, envelope.PermissionError)
	}
	if !user.Enabled || user.Type != models.UserTypeAgent {
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("user.accountDisabled"), nil))
	}

	totpOn, err := app.user.TOTPEnabled(user.ID)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	if totpOn {
		pending, err := issueTOTPPending(app, user.ID)
		if err != nil {
			app.lo.Error("error creating totp pending token", "error", err)
			return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
		}
		return r.SendEnvelope(map[string]any{"requires_totp": true, "pending_token": pending})
	}

	if err := startAgentSession(r, user, "magic_link"); err != nil {
		return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
	}
	return r.SendEnvelope(user)
}

// startAgentSession creates the session and CSRF cookie and records the login.
func startAgentSession(r *fastglue.Request, user models.User, method string) error {
	app := r.Context.(*App)
	if err := app.auth.SaveSession(amodels.User{
		ID:        user.ID,
		Email:     user.Email.String,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, r); err != nil {
		app.lo.Error("error saving session", "method", method, "user_id", user.ID, "error", err)
		return err
	}
	if err := app.auth.SetCSRFCookie(r); err != nil {
		app.lo.Error("error setting csrf cookie", "method", method, "error", err)
		return err
	}
	if err := app.user.UpdateLastLoginAt(user.ID); err != nil {
		app.lo.Error("error updating last login at", "method", method, "user_id", user.ID, "error", err)
	}
	app.user.InvalidateAgentCache(user.ID)
	if err := app.activityLog.Login(user.ID, user.Email.String, realip.FromRequest(r.RequestCtx)); err != nil {
		app.lo.Error("error creating login activity log", "error", err)
	}
	app.lo.Info("login successful", "method", method, "user_id", user.ID)
	return nil
}
