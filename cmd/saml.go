package main

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/samlsso"
	"github.com/abhinavxd/libredesk/internal/ssrf"
	"github.com/abhinavxd/libredesk/internal/stringutil"
	"github.com/abhinavxd/libredesk/internal/user/models"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

const (
	samlRelayKeyPrefix = "saml_relay:"
	samlRelayTTL       = 10 * time.Minute

	samlErrLoginFailed     = "saml_login_failed"
	samlErrNoAccount       = "saml_no_account"
	samlErrAccountDisabled = "saml_account_disabled"
)

// samlRelay is stored in Redis under the RelayState, which the IdP echoes back to the ACS.
// Redis is used instead of the session because the ACS is a cross-site POST that does not carry Lax cookies.
type samlRelay struct {
	RequestID string `json:"request_id"`
	Next      string `json:"next"`
}

// ssoSettings is the admin view of passwordless and SAML sign-in settings.
type ssoSettings struct {
	MagicLinkEnabled bool            `json:"magic_link_enabled"`
	SAML             samlAdminConfig `json:"saml"`
}

type samlAdminConfig struct {
	Enabled           bool   `json:"enabled"`
	Name              string `json:"name"`
	IDPMetadataURL    string `json:"idp_metadata_url"`
	IDPMetadataXML    string `json:"idp_metadata_xml"`
	AllowIDPInitiated bool   `json:"allow_idp_initiated"`
	EmailAttribute    string `json:"email_attribute"`
	EntityID          string `json:"entity_id"`
	ACSURL            string `json:"acs_url"`
	MetadataURL       string `json:"metadata_url"`
	SPCertificate     string `json:"sp_certificate"`
}

// initSAML builds the SAML manager from stored settings. A bad config is logged, not fatal.
func initSAML(app *App, dialControl ssrf.Control) *samlsso.Manager {
	transport := ssrf.NewTransport(dialControl, 3*time.Second)
	transport.TLSHandshakeTimeout = 5 * time.Second
	m := samlsso.New(&http.Client{Timeout: 15 * time.Second, Transport: transport}, initLogger("saml"))
	app.saml = m
	if err := reloadSAML(app); err != nil {
		app.lo.Error("error initializing saml sso", "error", err)
	}
	return m
}

// reloadSAML reloads the SAML config from settings.
func reloadSAML(app *App) error {
	cfg, err := getSAMLConfig(app)
	if err != nil {
		return err
	}
	rootURL, err := app.setting.GetAppRootURL()
	if err != nil {
		return err
	}
	return app.saml.Load(cfg, rootURL)
}

func getSAMLConfig(app *App) (samlsso.Config, error) {
	var cfg samlsso.Config
	b, err := app.setting.GetByPrefix("saml.")
	if err != nil {
		return cfg, err
	}
	if len(b) == 0 || string(b) == "null" {
		return cfg, nil
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// safeNext keeps post-login redirects on this site.
func safeNext(next string) string {
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.HasPrefix(next, "/\\") {
		return ""
	}
	return next
}

// handleSAMLLogin starts SP-initiated SAML sign-in.
func handleSAMLLogin(r *fastglue.Request) error {
	var (
		app  = r.Context.(*App)
		next = safeNext(string(r.RequestCtx.QueryArgs().Peek("next")))
	)
	relayState, err := stringutil.RandomAlphanumeric(32)
	if err != nil {
		app.lo.Error("error generating saml relay state", "error", err)
		return redirectLoginError(r, samlErrLoginFailed, next)
	}
	loginURL, requestID, err := app.saml.LoginURL(r.RequestCtx, relayState)
	if err != nil {
		app.lo.Error("error building saml login url", "error", err)
		return redirectLoginError(r, samlErrLoginFailed, next)
	}
	b, _ := json.Marshal(samlRelay{RequestID: requestID, Next: next})
	if err := app.redis.Set(context.Background(), samlRelayKeyPrefix+relayState, b, samlRelayTTL).Err(); err != nil {
		app.lo.Error("error storing saml relay state", "error", err)
		return redirectLoginError(r, samlErrLoginFailed, next)
	}
	return r.Redirect(loginURL, fasthttp.StatusFound, nil, "")
}

// handleSAMLACS receives the IdP's POSTed assertion and signs the agent in.
func handleSAMLACS(r *fastglue.Request) error {
	var (
		app          = r.Context.(*App)
		samlResponse = string(r.RequestCtx.PostArgs().Peek("SAMLResponse"))
		relayState   = string(r.RequestCtx.PostArgs().Peek("RelayState"))
		relay        samlRelay
		requestIDs   []string
	)
	if samlResponse == "" {
		return redirectLoginError(r, samlErrLoginFailed, "")
	}

	// A known RelayState means we started this login; otherwise it is IdP-initiated and only allowed when configured.
	if relayState != "" && len(relayState) <= 64 {
		if val, err := app.redis.GetDel(context.Background(), samlRelayKeyPrefix+relayState).Result(); err == nil {
			if err := json.Unmarshal([]byte(val), &relay); err == nil && relay.RequestID != "" {
				requestIDs = []string{relay.RequestID}
			}
		}
	}

	email, err := app.saml.ParseResponse(r.RequestCtx, samlResponse, requestIDs)
	if err != nil {
		app.lo.Error("saml response rejected", "error", err, "idp_initiated", len(requestIDs) == 0)
		return redirectLoginError(r, samlErrLoginFailed, relay.Next)
	}

	user, err := app.user.GetAgent(0, email)
	if err != nil {
		var e envelope.Error
		if errors.As(err, &e) && e.ErrorType == envelope.NotFoundError {
			app.lo.Warn("no agent account matching saml email", "email", email)
			return redirectLoginError(r, samlErrNoAccount, relay.Next)
		}
		return redirectLoginError(r, samlErrLoginFailed, relay.Next)
	}
	if user.Type != models.UserTypeAgent {
		return redirectLoginError(r, samlErrNoAccount, relay.Next)
	}
	if !user.Enabled {
		return redirectLoginError(r, samlErrAccountDisabled, relay.Next)
	}
	if err := startAgentSession(r, user, "saml"); err != nil {
		return redirectLoginError(r, samlErrLoginFailed, relay.Next)
	}

	// Leave the cross-site POST with a same-site navigation so the new Lax session cookie is sent.
	dest := html.EscapeString(stringOr(relay.Next, "/"))
	r.RequestCtx.SetContentType("text/html; charset=utf-8")
	r.RequestCtx.Response.Header.Set("Cache-Control", "no-store")
	r.RequestCtx.SetBodyString(`<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="refresh" content="0;url=` +
		dest + `"><title>Signing in</title></head><body><a href="` + dest + `">Continue</a></body></html>`)
	return nil
}

func stringOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// handleSAMLMetadata serves the SP metadata document for the IdP.
func handleSAMLMetadata(r *fastglue.Request) error {
	app := r.Context.(*App)
	b, err := app.saml.Metadata(r.RequestCtx)
	if err != nil {
		if errors.Is(err, samlsso.ErrNotConfigured) {
			return r.SendErrorEnvelope(fasthttp.StatusNotFound, app.i18n.T("admin.sso.samlNotConfigured"), nil, envelope.NotFoundError)
		}
		app.lo.Error("error generating saml metadata", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, app.i18n.T("globals.messages.somethingWentWrong"), nil, envelope.GeneralError)
	}
	r.RequestCtx.SetContentType("application/samlmetadata+xml")
	r.RequestCtx.SetBody(b)
	return nil
}

// handleGetSSOSettings returns magic link and SAML settings for the admin page.
func handleGetSSOSettings(r *fastglue.Request) error {
	app := r.Context.(*App)
	cfg, err := getSAMLConfig(app)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	rootURL, _ := app.setting.GetAppRootURL()
	return r.SendEnvelope(ssoSettings{
		MagicLinkEnabled: magicLinkEnabled(app),
		SAML: samlAdminConfig{
			Enabled:           cfg.Enabled,
			Name:              cfg.Name,
			IDPMetadataURL:    cfg.IDPMetadataURL,
			IDPMetadataXML:    cfg.IDPMetadataXML,
			AllowIDPInitiated: cfg.AllowIDPInitiated,
			EmailAttribute:    cfg.EmailAttribute,
			EntityID:          samlsso.MetadataURL(rootURL),
			ACSURL:            samlsso.ACSURL(rootURL),
			MetadataURL:       samlsso.MetadataURL(rootURL),
			SPCertificate:     cfg.SPCertificate,
		},
	})
}

// handleUpdateSSOSettings validates and saves magic link and SAML settings.
func handleUpdateSSOSettings(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req ssoSettings
	)
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("errors.parsingRequest"), nil, envelope.InputError)
	}
	current, err := getSAMLConfig(app)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	rootURL, err := app.setting.GetAppRootURL()
	if err != nil {
		return sendErrorEnvelope(r, err)
	}

	next := samlsso.Config{
		Enabled:           req.SAML.Enabled,
		Name:              strings.TrimSpace(req.SAML.Name),
		IDPMetadataURL:    strings.TrimSpace(req.SAML.IDPMetadataURL),
		IDPMetadataXML:    strings.TrimSpace(req.SAML.IDPMetadataXML),
		AllowIDPInitiated: req.SAML.AllowIDPInitiated,
		EmailAttribute:    strings.TrimSpace(req.SAML.EmailAttribute),
		SPCertificate:     current.SPCertificate,
		SPPrivateKey:      current.SPPrivateKey,
	}
	if next.Name == "" {
		next.Name = "SAML SSO"
	}
	if next.SPCertificate == "" || next.SPPrivateKey == "" {
		cert, key, err := samlsso.GenerateKeyPair("libredesk-saml-sp")
		if err != nil {
			app.lo.Error("error generating saml sp key pair", "error", err)
			return sendErrorEnvelope(r, envelope.NewError(envelope.GeneralError, app.i18n.T("globals.messages.somethingWentWrong"), nil))
		}
		next.SPCertificate, next.SPPrivateKey = cert, key
	}

	// Build the SP before saving so a bad metadata URL or XML is reported instead of stored.
	if next.Enabled {
		if rootURL == "" {
			return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("admin.sso.rootURLRequired"), nil, envelope.InputError)
		}
		probe := samlsso.New(app.saml.HTTPClient(), app.lo)
		if err := probe.Load(next, rootURL); err != nil {
			return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.Ts("admin.sso.samlInvalid", "error", err.Error()), nil, envelope.InputError)
		}
	}

	if err := app.setting.Update(map[string]any{
		magicLinkEnabledKey:        req.MagicLinkEnabled,
		"saml.enabled":             next.Enabled,
		"saml.name":                next.Name,
		"saml.idp_metadata_url":    next.IDPMetadataURL,
		"saml.idp_metadata_xml":    next.IDPMetadataXML,
		"saml.allow_idp_initiated": next.AllowIDPInitiated,
		"saml.email_attribute":     next.EmailAttribute,
		"saml.sp_certificate":      next.SPCertificate,
		"saml.sp_private_key":      next.SPPrivateKey,
	}); err != nil {
		return sendErrorEnvelope(r, err)
	}
	if err := reloadSAML(app); err != nil {
		app.lo.Error("error reloading saml after update", "error", err)
	}
	return handleGetSSOSettings(r)
}
