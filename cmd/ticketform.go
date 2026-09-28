package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/stringutil"
	tfmodels "github.com/abhinavxd/libredesk/internal/ticketform/models"
	umodels "github.com/abhinavxd/libredesk/internal/user/models"
	"github.com/valyala/fasthttp"
	"github.com/volatiletech/null/v9"
	"github.com/zerodha/fastglue"
)

var ticketFieldKey = regexp.MustCompile(`^[a-z][a-z0-9_]{0,40}$`)

type ticketFormRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InboxID     int             `json:"inbox_id"`
	Enabled     bool            `json:"enabled"`
	Fields      json.RawMessage `json:"fields"`
}

func handleGetTicketForms(r *fastglue.Request) error {
	app := r.Context.(*App)
	forms, err := app.ticketForm.GetAll()
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(forms)
}

func handleCreateTicketForm(r *fastglue.Request) error {
	app := r.Context.(*App)
	req, fields, err := decodeTicketForm(r, app)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	form, err := app.ticketForm.Create(req.Name, req.Description, req.InboxID, req.Enabled, fields)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(form)
}

func handleUpdateTicketForm(r *fastglue.Request) error {
	app := r.Context.(*App)
	id, err := parsePathID(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("errors.parsingRequest"), nil, envelope.InputError)
	}
	req, fields, err := decodeTicketForm(r, app)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	form, err := app.ticketForm.Update(id, req.Name, req.Description, req.InboxID, req.Enabled, fields)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(form)
}

func handleDeleteTicketForm(r *fastglue.Request) error {
	app := r.Context.(*App)
	id, err := parsePathID(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, app.i18n.T("errors.parsingRequest"), nil, envelope.InputError)
	}
	if err := app.ticketForm.Delete(id); err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(true)
}

func handlePublicTicketForm(r *fastglue.Request) error {
	app := r.Context.(*App)
	form, fields, ok := loadPublicTicketForm(r)
	if !ok {
		return nil
	}
	return renderTicketForm(r, app, form, fields, string(r.RequestCtx.QueryArgs().Peek("sent")) == "1", "")
}

func handlePublicTicketFormSubmit(r *fastglue.Request) error {
	app := r.Context.(*App)
	form, fields, ok := loadPublicTicketForm(r)
	if !ok {
		return nil
	}
	email := strings.ToLower(strings.TrimSpace(string(r.RequestCtx.FormValue("email"))))
	name := strings.TrimSpace(string(r.RequestCtx.FormValue("name")))
	subject := strings.TrimSpace(string(r.RequestCtx.FormValue("subject")))
	message := strings.TrimSpace(string(r.RequestCtx.FormValue("message")))
	if !stringutil.ValidEmail(email) || subject == "" || message == "" || utf8.RuneCountInString(subject) > 200 || utf8.RuneCountInString(message) > 20000 {
		return renderTicketForm(r, app, form, fields, false, app.i18n.T("ticketForm.invalid"))
	}
	inbox, err := app.inbox.GetDBRecord(form.InboxID)
	if err != nil || !inbox.Enabled {
		return renderTicketForm(r, app, form, fields, false, app.i18n.T("ticketForm.unavailable"))
	}
	attrs := map[string]any{}
	for _, field := range fields {
		value := strings.TrimSpace(string(r.RequestCtx.FormValue("field_" + field.Key)))
		if field.Required && value == "" {
			return renderTicketForm(r, app, form, fields, false, app.i18n.T("ticketForm.invalid"))
		}
		if utf8.RuneCountInString(value) > 2000 {
			value = string([]rune(value)[:2000])
		}
		if value != "" {
			attrs[field.Key] = value
		}
	}
	contact := umodels.User{
		Email:            null.StringFrom(email),
		FirstName:        name,
		CustomAttributes: json.RawMessage(`{}`),
	}
	if err := app.user.ResolveContact(&contact, umodels.ContactReuse); err != nil {
		app.lo.Error("error resolving ticket form contact", "error", err)
		return renderTicketForm(r, app, form, fields, false, app.i18n.T("ticketForm.unavailable"))
	}
	_, uuid, err := app.conversation.CreateConversation(contact.ID, form.InboxID, "", time.Now(), subject, true, nil, attrs, 0, 0)
	if err != nil {
		app.lo.Error("error creating ticket form conversation", "error", err)
		return renderTicketForm(r, app, form, fields, false, app.i18n.T("ticketForm.unavailable"))
	}
	if _, err := app.conversation.CreateContactMessage(nil, contact.ID, uuid, plainToHTML(message), "html", true, ""); err != nil {
		app.lo.Error("error creating ticket form message", "error", err)
	}
	r.RequestCtx.Redirect("/forms/"+strconv.Itoa(form.ID)+"?sent=1", fasthttp.StatusSeeOther)
	return nil
}

func loadPublicTicketForm(r *fastglue.Request) (tfmodels.TicketForm, []tfmodels.Field, bool) {
	app := r.Context.(*App)
	id, err := parsePathID(r)
	if err != nil {
		_ = app.tmpl.RenderWebPage(r.RequestCtx, "error", map[string]any{
			"Data": map[string]any{"ErrorMessage": app.i18n.T("globals.messages.pageNotFound")},
		})
		return tfmodels.TicketForm{}, nil, false
	}
	form, err := app.ticketForm.Get(id)
	if err != nil || !form.Enabled {
		_ = app.tmpl.RenderWebPage(r.RequestCtx, "error", map[string]any{
			"Data": map[string]any{"ErrorMessage": app.i18n.T("globals.messages.pageNotFound")},
		})
		return tfmodels.TicketForm{}, nil, false
	}
	var fields []tfmodels.Field
	_ = json.Unmarshal(form.Fields, &fields)
	return form, fields, true
}

func renderTicketForm(r *fastglue.Request, app *App, form tfmodels.TicketForm, fields []tfmodels.Field, sent bool, errMsg string) error {
	return app.tmpl.RenderWebPage(r.RequestCtx, "ticket-form", map[string]any{
		"Data": map[string]any{
			"Title":  form.Name,
			"Form":   form,
			"Fields": fields,
			"Sent":   sent,
			"Error":  errMsg,
		},
	})
}

func decodeTicketForm(r *fastglue.Request, app *App) (ticketFormRequest, json.RawMessage, error) {
	var req ticketFormRequest
	if err := r.Decode(&req, "json"); err != nil {
		return req, nil, envelope.NewError(envelope.InputError, app.i18n.T("errors.parsingRequest"), nil)
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || utf8.RuneCountInString(req.Name) > 120 || utf8.RuneCountInString(req.Description) > 500 || req.InboxID <= 0 {
		return req, nil, envelope.NewError(envelope.InputError, app.i18n.T("ticketForm.invalid"), nil)
	}
	inbox, err := app.inbox.GetDBRecord(req.InboxID)
	if err != nil || inbox.ID == 0 {
		return req, nil, envelope.NewError(envelope.InputError, app.i18n.T("ticketForm.invalid"), nil)
	}
	fields, raw, err := normalizeTicketFields(req.Fields)
	if err != nil || len(fields) > 20 {
		return req, nil, envelope.NewError(envelope.InputError, app.i18n.T("ticketForm.invalid"), nil)
	}
	return req, raw, nil
}

func normalizeTicketFields(raw json.RawMessage) ([]tfmodels.Field, json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		raw = []byte("[]")
	}
	var fields []tfmodels.Field
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, nil, err
	}
	seen := map[string]bool{}
	out := make([]tfmodels.Field, 0, len(fields))
	for _, field := range fields {
		field.Key = strings.TrimSpace(field.Key)
		field.Label = strings.TrimSpace(field.Label)
		field.Type = strings.TrimSpace(field.Type)
		if field.Label == "" || !ticketFieldKey.MatchString(field.Key) || seen[field.Key] {
			return nil, nil, envelope.NewError(envelope.InputError, "invalid field", nil)
		}
		if field.Type != "textarea" {
			field.Type = "text"
		}
		seen[field.Key] = true
		out = append(out, field)
	}
	encoded, err := json.Marshal(out)
	return out, encoded, err
}

func parsePathID(r *fastglue.Request) (int, error) {
	return strconv.Atoi(r.RequestCtx.UserValue("id").(string))
}
