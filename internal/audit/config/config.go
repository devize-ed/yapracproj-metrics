// Package audit provides configuration structures for the audit component.
package audit

type AuditConfig struct {
	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`
}
