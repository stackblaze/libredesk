package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V2_5_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf) error {
	// Grant the new `contacts:delete` permission to the Admin role.
	_, err := db.Exec(`
		UPDATE roles
		SET permissions = array_append(permissions, 'contacts:delete')
		WHERE name = 'Admin' AND NOT ('contacts:delete' = ANY(permissions));
	`)
	return err
}
