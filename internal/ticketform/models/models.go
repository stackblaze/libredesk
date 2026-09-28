package models

import (
	"encoding/json"
	"time"
)

// TicketForm is a public request form that files a conversation in an inbox.
type TicketForm struct {
	ID          int             `db:"id" json:"id"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	InboxID     int             `db:"inbox_id" json:"inbox_id"`
	Enabled     bool            `db:"enabled" json:"enabled"`
	Fields      json.RawMessage `db:"fields" json:"fields"`
	InboxName   string          `db:"inbox_name" json:"inbox_name"`
}

// Field is one extra question on a ticket form.
type Field struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}
