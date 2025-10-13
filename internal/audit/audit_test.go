package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNewAuditor(t *testing.T) {
	var tests = []struct {
		name      string
		auditFile string
		auditURL  string
	}{
		{
			name:      "with_file_and_url",
			auditFile: "/tmp/test-audit.json",
			auditURL:  "http://192.168.1.1:8080",
		},
		{
			name:      "with_file_only",
			auditFile: "/tmp/test-audit.json",
			auditURL:  "",
		},
		{
			name:      "with_url_only",
			auditFile: "",
			auditURL:  "http://192.168.1.1:8080",
		},
		{
			name:      "no_config",
			auditFile: "",
			auditURL:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			auditor := NewAuditor(logger, tt.auditFile, tt.auditURL)

			assert.NotNil(t, auditor, "NewAuditor should not return nil")
			assert.Equal(t, tt.auditFile, auditor.auditFile, "auditFile should match")
			assert.Equal(t, tt.auditURL, auditor.auditURL, "auditURL should match")
			assert.NotNil(t, auditor.eventChan, "eventChan should not be nil")
			assert.NotNil(t, auditor.registerChan, "registerChan should not be nil")
			assert.NotNil(t, auditor.logger, "logger should not be nil")
		})
	}
}

func TestAuditor_Send(t *testing.T) {
	var tests = []struct {
		name    string
		addr    string
		metrics []string
	}{
		{
			name:    "single_metric",
			addr:    "192.168.1.1",
			metrics: []string{"cpu"},
		},
		{
			name:    "multiple_metrics",
			addr:    "192.168.1.111",
			metrics: []string{"testMetric1", "testMetric2", "testMetric3"},
		},
		{
			name:    "empty_metrics",
			addr:    "192.168.1.1",
			metrics: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			auditor := NewAuditor(logger, "", "")

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			go auditor.Run(ctx)

			time.Sleep(10 * time.Millisecond)
			auditor.Send(tt.addr, tt.metrics)
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestAuditor_Register(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	auditor := NewAuditor(logger, "", "")

	sub := auditor.Register()
	assert.NotNil(t, sub, "Register returned nil channel")

	select {
	case <-auditor.registerChan:
		assert.True(t, true)
	default:
		assert.Fail(t, "Subscription failed")
	}
}

func TestAuditor_Run(t *testing.T) {
	var tests = []struct {
		name      string
		auditFile string
		auditURL  string
		addr      string
		metrics   []string
		wantFile  bool
	}{
		{
			name:      "with_file_audit",
			auditFile: "test-audit.json",
			auditURL:  "",
			addr:      "192.168.1.1",
			metrics:   []string{"testMetric1", "testMetric2"},
			wantFile:  true,
		},
		{
			name:      "with_url_audit",
			auditFile: "",
			auditURL:  "http://example.com/audit",
			addr:      "192.168.1.1",
			metrics:   []string{"testMetric1", "testMetric2"},
			wantFile:  false,
		},
		{
			name:      "with_both_audits",
			auditFile: "test-audit.json",
			auditURL:  "http://example.com/audit",
			addr:      "192.168.1.1",
			metrics:   []string{"testMetric1", "testMetric2"},
			wantFile:  true,
		},
		{
			name:      "no_audit_config",
			auditFile: "",
			auditURL:  "",
			addr:      "192.168.1.1",
			metrics:   []string{"testMetric1", "testMetric2"},
			wantFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var auditFile string
			if tt.auditFile != "" {
				tmpDir := t.TempDir()
				auditFile = filepath.Join(tmpDir, tt.auditFile)
			}

			logger := zaptest.NewLogger(t).Sugar()
			auditor := NewAuditor(logger, auditFile, tt.auditURL)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			go auditor.Run(ctx)

			time.Sleep(10 * time.Millisecond)

			auditor.Send(tt.addr, tt.metrics)

			time.Sleep(50 * time.Millisecond)

			<-ctx.Done()

			if tt.wantFile {
				assert.FileExists(t, auditFile, "Audit file was not created")

				content, err := os.ReadFile(auditFile)
				assert.NoError(t, err, "Failed to read audit file")
				assert.NotEmpty(t, content, "Audit file is empty")

				var auditMsg AuditMsg
				err = json.Unmarshal(content, &auditMsg)
				assert.NoError(t, err, "Failed to unmarshal audit message")
				assert.Equal(t, tt.addr, auditMsg.Addr, "Address does not match")
				assert.Len(t, auditMsg.Metrics, len(tt.metrics), "Metrics length does not match")
			}
		})
	}
}

func TestRunFileAudit(t *testing.T) {
	var tests = []struct {
		name      string
		msg       AuditMsg
		filePath  string
		wantFile  bool
		wantError bool
	}{
		{
			name: "valid_message",
			msg: AuditMsg{
				TimeStamp: time.Now(),
				Addr:      "192.168.1.1",
				Metrics:   []string{"testMetric1", "testMetric2"},
			},
			filePath:  "audit.json",
			wantFile:  true,
			wantError: false,
		},
		{
			name: "empty_metrics",
			msg: AuditMsg{
				TimeStamp: time.Now(),
				Addr:      "192.168.1.1",
				Metrics:   []string{},
			},
			filePath:  "audit.json",
			wantFile:  true,
			wantError: false,
		},
		{
			name: "invalid_file_path",
			msg: AuditMsg{
				TimeStamp: time.Now(),
				Addr:      "192.168.1.1",
				Metrics:   []string{"testMetric1"},
			},
			filePath:  "/tmp",
			wantFile:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var auditFile string
			if tt.filePath == "audit.json" {
				tmpDir := t.TempDir()
				auditFile = filepath.Join(tmpDir, tt.filePath)
			} else {
				auditFile = tt.filePath
			}

			logger := zaptest.NewLogger(t).Sugar()
			msgChan := make(chan AuditMsg, 1)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			go RunFileAudit(ctx, msgChan, auditFile, logger)

			msgChan <- tt.msg

			<-ctx.Done()

			if tt.wantFile && !tt.wantError {
				assert.FileExists(t, auditFile, "Audit file should be created")

				content, err := os.ReadFile(auditFile)
				assert.NoError(t, err, "Failed to read audit file")
				assert.NotEmpty(t, content, "Audit file should contain data")

				var auditMsg AuditMsg
				err = json.Unmarshal(content, &auditMsg)
				assert.NoError(t, err, "Failed to unmarshal audit message")
				assert.Equal(t, tt.msg.Addr, auditMsg.Addr, "Address should match")
			}
		})
	}
}

func TestRunURLAudit(t *testing.T) {
	var tests = []struct {
		name string
		url  string
		msg  AuditMsg
	}{
		{
			name: "valid_url",
			url:  "http://192.268.1.1/audit",
			msg: AuditMsg{
				TimeStamp: time.Now(),
				Addr:      "192.168.1.1",
				Metrics:   []string{"testMetric1", "testMetric2"},
			},
		},
		{
			name: "invalid_url",
			url:  "http://none",
			msg: AuditMsg{
				TimeStamp: time.Now(),
				Addr:      "192.168.1.1",
				Metrics:   []string{"testMetric1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			msgChan := make(chan AuditMsg, 1)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			done := make(chan struct{})
			go func() {
				RunURLAudit(ctx, msgChan, tt.url, logger)
				close(done)
			}()

			msgChan <- tt.msg

			<-ctx.Done()

			select {
			case <-done:
			case <-time.After(200 * time.Millisecond):
				assert.Fail(t, "URL audit should have stopped after context cancellation")
			}
		})
	}
}

// Benchmark tests
func BenchmarkAuditor_Send(b *testing.B) {
	logger := zap.NewNop().Sugar()
	auditor := NewAuditor(logger, "", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auditor.Send("192.168.1.1", []string{"testMetric1", "testMetric2"})
	}
}
