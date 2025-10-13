package audit

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// AuditMsg is a struct that contains the audit message.
type AuditMsg struct {
	TimeStamp time.Time `json:"ts"`
	Addr      string    `json:"ip_address"`
	Metrics   []string  `json:"metrics"`
}

// Auditor is a struct that contains channels for events and registrations.
type Auditor struct {
	eventChan    chan AuditMsg
	registerChan chan chan AuditMsg
	logger       *zap.SugaredLogger
}

// NewAuditor creates a new auditor.
func NewAuditor(logger *zap.SugaredLogger) *Auditor {

	return &Auditor{
		eventChan:    make(chan AuditMsg),
		registerChan: make(chan chan AuditMsg, 2),
		logger:       logger,
	}
}

// Run starts the auditor.
func (a *Auditor) Run(ctx context.Context, auditFile string, auditURL string) {
	a.logger.Debugf("starting auditor")
	// Create a map of subscriptions.
	subs := make(map[chan AuditMsg]struct{})
	// if audit file is set, start the file auditor
	if auditFile != "" {
		ch := a.Register()
		go RunFileAudit(ctx, ch, auditFile, a.logger)
	}
	// if audit URL is set, start the URL auditor
	if auditURL != "" {
		ch := a.Register()
		go RunURLAudit(ctx, ch, auditURL, a.logger)
	}
	// if audit file and URL are not set, skip the auditors
	if auditFile == "" && auditURL == "" {
		a.logger.Debugf("audit file and URL are not set, skipping auditors")
		return
	}
	// start the auditors
	for {
		select {
		// If the context is done, close all subscriptions.
		case <-ctx.Done():
			// Close all subscriptions.
			for sub := range subs {
				close(sub)
			}
			return
		// If a new subscription is registered, add it to the map.
		case sub := <-a.registerChan:
			subs[sub] = struct{}{}
			a.logger.Debugf("new subscription for auditor: %v", sub)
		// If a new message is received, send it to all subscriptions.
		case msg := <-a.eventChan:
			a.logger.Debugf("received message from event channel: %v", msg)
			for sub := range subs {
				a.logger.Debugf("sending message to subscription: %v", sub)
				sub <- msg
			}
		}
	}
}

// Send sends a message to the auditor.
func (a *Auditor) Send(addr string, metrics []string) {
	msg := AuditMsg{
		TimeStamp: time.Now(),
		Addr:      addr,
		Metrics:   metrics,
	}
	select {
	case a.eventChan <- msg:
		a.logger.Debugf("sent message to auditor: %v", msg)
	default:
		a.logger.Warn("audit queue is full; dropping message")
	}
}

// Register registers a new subscription to the auditor.
func (a *Auditor) Register() chan AuditMsg {
	sub := make(chan AuditMsg)
	a.registerChan <- sub
	return sub
}

// RunFileAudit runs the file auditor.
func RunFileAudit(ctx context.Context, ch <-chan AuditMsg, fname string, logger *zap.SugaredLogger) {
	// Open the audit file.
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		logger.Errorf("open audit file: %v", err)
		return
	}
	// Close the audit file.
	defer func() {
		if err := f.Sync(); err != nil {
			logger.Errorf("sync audit file: %v", err)
		}
		if err := f.Close(); err != nil {
			logger.Errorf("close audit file: %v", err)
		}
	}()
	// Create a new JSON encoder.
	enc := json.NewEncoder(f)
	// Wait for messages from the channel and send them to the file.
	for {
		select {
		// If the context is done, exit
		case <-ctx.Done():
			logger.Debugf("context done, exiting")
			return
		// If a new message is received, send it to the file.
		case msg, ok := <-ch:
			// If the channel is closed, exit
			if !ok {
				return
			}
			// If the message is not encoded, exit
			if err := enc.Encode(msg); err != nil {
				logger.Errorf("write audit json: %v", err)
			}
		}
	}
}

// RunURLAudit runs the URL auditor.
func RunURLAudit(ctx context.Context, ch <-chan AuditMsg, url string, logger *zap.SugaredLogger) {
	// Create a new resty client.
	client := resty.New().SetTimeout(10 * time.Second)
	// Wait for messages from the channel and send them to the URL.
	for {
		select {
		// If the context is done, exit
		case <-ctx.Done():
			logger.Debugf("context done, exiting")
			return
		// If a new message is received, send it to the URL.
		case msg, ok := <-ch:
			// If the channel is closed, exit
			if !ok {
				return
			}
			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(msg).
				SetContext(ctx)
			if _, err := req.Post(url); err != nil {
				logger.Errorf("send audit to %s: %v", url, err)
			}
		}
	}
}
