package main

import (
	"encoding/json"
	"strings"

	amodels "github.com/abhinavxd/libredesk/internal/auth/models"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

const onboardingStatusKey = "onboarding.status"

// Onboarding status values. An empty status means the wizard has not been finished or dismissed.
const (
	onboardingStatusNone      = ""
	onboardingStatusDismissed = "dismissed"
	onboardingStatusCompleted = "completed"
)

// onboardingStep reports whether one setup step is done, derived from live data rather than wizard clicks.
type onboardingStep struct {
	Key      string `json:"key"`
	Done     bool   `json:"done"`
	Required bool   `json:"required"`
}

type onboardingState struct {
	Status string           `json:"status"`
	Steps  []onboardingStep `json:"steps"`
	// ShowWizard is true when an admin should be sent to the wizard after login.
	ShowWizard bool `json:"show_wizard"`
}

type onboardingUpdateReq struct {
	Status string `json:"status"`
}

func getOnboardingStatus(app *App) string {
	b, err := app.setting.Get(onboardingStatusKey)
	if err != nil {
		return onboardingStatusNone
	}
	var s string
	_ = json.Unmarshal(b, &s)
	return s
}

// buildOnboardingState checks each setup step against the current configuration.
func buildOnboardingState(app *App, currentUserID int) (onboardingState, error) {
	general, err := app.setting.GetAll()
	if err != nil {
		return onboardingState{}, err
	}
	inboxes, err := app.inbox.GetAll()
	if err != nil {
		return onboardingState{}, err
	}
	agents, err := app.user.GetAgents()
	if err != nil {
		return onboardingState{}, err
	}
	teams, err := app.team.GetAll()
	if err != nil {
		return onboardingState{}, err
	}

	rootURL := strings.ToLower(general.RootURL)
	workspaceDone := general.SiteName != "" && !strings.EqualFold(general.SiteName, "libredesk") &&
		rootURL != "" && !strings.Contains(rootURL, "localhost") && !strings.Contains(rootURL, "127.0.0.1")
	emailDone := general.EmailNotification.Enabled && general.EmailNotification.Host != ""

	// Teammates means an agent other than the admin running the wizard. The built-in
	// System user is already excluded from this list.
	teammates := 0
	for _, a := range agents {
		if a.Type == "agent" && a.ID != currentUserID {
			teammates++
		}
	}

	signInDone := magicLinkEnabled(app) || (app.saml != nil && app.saml.Enabled())
	if providers, err := app.oidc.GetAll(); err == nil {
		for _, p := range providers {
			if p.Enabled {
				signInDone = true
			}
		}
	}

	steps := []onboardingStep{
		{Key: "workspace", Done: workspaceDone, Required: true},
		{Key: "email", Done: emailDone, Required: true},
		{Key: "inbox", Done: len(inboxes) > 0, Required: true},
		{Key: "teammates", Done: teammates > 0, Required: false},
		{Key: "teams", Done: len(teams) > 0, Required: false},
		{Key: "sign_in", Done: signInDone, Required: false},
	}

	status := getOnboardingStatus(app)
	requiredDone := true
	for _, s := range steps {
		if s.Required && !s.Done {
			requiredDone = false
		}
	}
	return onboardingState{
		Status:     status,
		Steps:      steps,
		ShowWizard: status == onboardingStatusNone && !requiredDone,
	}, nil
}

// handleGetOnboarding returns setup progress for the onboarding wizard.
func handleGetOnboarding(r *fastglue.Request) error {
	app := r.Context.(*App)
	user, _ := r.RequestCtx.UserValue("user").(amodels.User)
	state, err := buildOnboardingState(app, user.ID)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(state)
}

// handleUpdateOnboarding records that the wizard was completed, dismissed, or reopened.
func handleUpdateOnboarding(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
		req onboardingUpdateReq
	)
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("errors.parsingRequest"), nil, envelope.InputError)
	}
	switch req.Status {
	case onboardingStatusNone, onboardingStatusDismissed, onboardingStatusCompleted:
	default:
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("globals.messages.badRequest"), nil, envelope.InputError)
	}
	if err := app.setting.Update(map[string]any{onboardingStatusKey: req.Status}); err != nil {
		return sendErrorEnvelope(r, err)
	}
	return handleGetOnboarding(r)
}
