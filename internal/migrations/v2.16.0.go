package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V2_16_0 adds ticket forms used by the public request page.
func V2_16_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ticket_forms (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			inbox_id INT NOT NULL REFERENCES inboxes(id) ON DELETE RESTRICT ON UPDATE CASCADE,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			fields JSONB NOT NULL DEFAULT '[]'::jsonb
		);
	`)
	return err
}
