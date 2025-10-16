// Package audit provides configuration structures for the audit component.
package audit

type AuditConfig struct {
	AuditFile string `env:"AUDIT_FILE"` // File path for storing audit data
	AuditURL  string `env:"AUDIT_URL"`  // URL for sending audit data to the remote server
}
