package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V2_18_0 adds the onboarding wizard status. Workspaces that already have inboxes are
// treated as set up, so the wizard does not open for their admins after upgrading.
func V2_18_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf) error {
	_, err := db.Exec(`
		INSERT INTO settings (key, value)
		SELECT 'onboarding.status',
			CASE WHEN EXISTS (SELECT 1 FROM inboxes WHERE deleted_at IS NULL)
				THEN '"dismissed"'::jsonb ELSE '""'::jsonb END
		ON CONFLICT (key) DO NOTHING;
	`)
	return err
}
