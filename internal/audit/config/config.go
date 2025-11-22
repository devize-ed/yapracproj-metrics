// Package audit provides configuration structures for the audit component.
package audit

type AuditConfig struct {
	AuditFile string `env:"AUDIT_FILE" json:"audit_file"` // File path for storing audit data
	AuditURL  string `env:"AUDIT_URL" json:"audit_url"`   // URL for sending audit data to the remote server
}
