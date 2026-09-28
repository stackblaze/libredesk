// Package ticketform stores public request forms.
package ticketform

import (
	"database/sql"
	"embed"
	"encoding/json"

	"github.com/abhinavxd/libredesk/internal/dbutil"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/abhinavxd/libredesk/internal/ticketform/models"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-i18n"
	"github.com/zerodha/logf"
)

var (
	//go:embed queries.sql
	efs embed.FS
)

// Manager handles ticket form records.
type Manager struct {
	q    queries
	lo   *logf.Logger
	i18n *i18n.I18n
}

// Opts contains options for initializing the Manager.
type Opts struct {
	DB   *sqlx.DB
	Lo   *logf.Logger
	I18n *i18n.I18n
}

type queries struct {
	GetAll *sqlx.Stmt `query:"get-all-ticket-forms"`
	Get    *sqlx.Stmt `query:"get-ticket-form"`
	Insert *sqlx.Stmt `query:"insert-ticket-form"`
	Update *sqlx.Stmt `query:"update-ticket-form"`
	Delete *sqlx.Stmt `query:"delete-ticket-form"`
}

// New creates and returns a new instance of the Manager.
func New(opts Opts) (*Manager, error) {
	var q queries
	if err := dbutil.ScanSQLFile("queries.sql", &q, opts.DB, efs); err != nil {
		return nil, err
	}
	return &Manager{q: q, lo: opts.Lo, i18n: opts.I18n}, nil
}

// GetAll returns every ticket form.
func (m *Manager) GetAll() ([]models.TicketForm, error) {
	out := make([]models.TicketForm, 0)
	if err := m.q.GetAll.Select(&out); err != nil {
		m.lo.Error("error fetching ticket forms", "error", err)
		return nil, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return out, nil
}

// Get returns one ticket form.
func (m *Manager) Get(id int) (models.TicketForm, error) {
	var form models.TicketForm
	if err := m.q.Get.Get(&form, id); err != nil {
		if err == sql.ErrNoRows {
			return form, envelope.NewError(envelope.NotFoundError, m.i18n.T("globals.messages.pageNotFound"), nil)
		}
		m.lo.Error("error fetching ticket form", "error", err)
		return form, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return form, nil
}

// Create inserts a ticket form.
func (m *Manager) Create(name, description string, inboxID int, enabled bool, fields json.RawMessage) (models.TicketForm, error) {
	var form models.TicketForm
	if err := m.q.Insert.Get(&form, name, description, inboxID, enabled, []byte(fields)); err != nil {
		m.lo.Error("error inserting ticket form", "error", err)
		return form, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return form, nil
}

// Update changes a ticket form.
func (m *Manager) Update(id int, name, description string, inboxID int, enabled bool, fields json.RawMessage) (models.TicketForm, error) {
	var form models.TicketForm
	if err := m.q.Update.Get(&form, id, name, description, inboxID, enabled, []byte(fields)); err != nil {
		if err == sql.ErrNoRows {
			return form, envelope.NewError(envelope.NotFoundError, m.i18n.T("globals.messages.pageNotFound"), nil)
		}
		m.lo.Error("error updating ticket form", "error", err)
		return form, envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return form, nil
}

// Delete removes a ticket form.
func (m *Manager) Delete(id int) error {
	if _, err := m.q.Delete.Exec(id); err != nil {
		m.lo.Error("error deleting ticket form", "error", err)
		return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
	}
	return nil
}
