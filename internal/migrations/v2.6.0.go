package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V2_6_0 patches the default outgoing email template so Gmail can load the
// logo (public PNG via EmailLogoURL) and agent replies sign with the author's
// name instead of a hardcoded team label.
func V2_6_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf) error {
	_, err := db.Exec(`
		UPDATE templates
		SET body = replace(
			replace(
				body,
				'src="{{ LogoURL }}"',
				'src="{{ EmailLogoURL }}"'
			),
			'<div class="signature">Best,<br>Stackblaze Support</div>',
			'<div class="signature">Best,<br>{{ if .Author.FullName }}{{ .Author.FullName }}{{ else if ne SiteName "" }}{{ SiteName }} Support{{ else }}Stackblaze Support{{ end }}</div>'
		)
		WHERE is_default = true
		  AND type = 'email_outgoing'
		  AND (
			body LIKE '%src="{{ LogoURL }}"%'
			OR body LIKE '%<div class="signature">Best,<br>Stackblaze Support</div>%'
			OR body LIKE '%<div class="signature">Best regards,<br>Stackblaze Support</div>%'
		  );
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE templates
		SET body = replace(
			body,
			'<div class="signature">Best regards,<br>Stackblaze Support</div>',
			'<div class="signature">Best regards,<br>{{ if .Author.FullName }}{{ .Author.FullName }}{{ else if ne SiteName "" }}{{ SiteName }} Support{{ else }}Stackblaze Support{{ end }}</div>'
		)
		WHERE is_default = true
		  AND type = 'email_outgoing'
		  AND body LIKE '%<div class="signature">Best regards,<br>Stackblaze Support</div>%';
	`)
	return err
}
