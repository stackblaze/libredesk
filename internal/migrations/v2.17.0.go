package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V2_17_0 adds settings for passwordless email sign-in and SAML SSO.
func V2_17_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf) error {
	_, err := db.Exec(`
		INSERT INTO settings (key, value) VALUES
			('auth.magic_link_enabled', 'false'::jsonb),
			('saml.enabled', 'false'::jsonb),
			('saml.name', '"SAML SSO"'::jsonb),
			('saml.idp_metadata_url', '""'::jsonb),
			('saml.idp_metadata_xml', '""'::jsonb),
			('saml.allow_idp_initiated', 'false'::jsonb),
			('saml.email_attribute', '""'::jsonb),
			('saml.sp_certificate', '""'::jsonb),
			('saml.sp_private_key', '""'::jsonb)
		ON CONFLICT (key) DO NOTHING;
	`)
	return err
}
