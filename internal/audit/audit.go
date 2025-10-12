package audit

import (
	"context"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// AuditMsg is a struct that contains the audit message.
type AuditMsg struct {
	TimeStamp time.Time `json:"ts"`
	Addr      string    `json:"ip_address"`
	Metrics   string    `json:"metrics"`
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
func (a *Auditor) Run(ctx context.Context) {
	a.logger.Debugf("starting auditor")
	// Create a map of subscriptions.
	subs := make(map[chan AuditMsg]struct{})
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
func (a *Auditor) Send(addr, metrics string) {
	msg := AuditMsg{
		TimeStamp: time.Now(),
		Addr:      addr,
		Metrics:   metrics,
	}
	a.eventChan <- msg
	a.logger.Debugf("sent message to auditor: %v", msg)
}

// Register registers a new subscription to the auditor.
func (a *Auditor) Register() chan AuditMsg {
	sub := make(chan AuditMsg)
	a.registerChan <- sub
	return sub
}

// RunFileAudit runs the file auditor.
func RunFileAudit(ch <-chan AuditMsg, fname string, logger *zap.SugaredLogger) {
	for msg := range ch {
		logger.Debugf("received message from event channel: %v", msg)
		// Save the message to the file.
		if err := os.WriteFile(fname, []byte(msg.Metrics), 0644); err != nil {
			logger.Errorf("failed to save message to file: %v", err)
		}
		logger.Debugf("message saved to file: %v", fname)
	}
}

// RunURLAudit runs the URL auditor.
func RunURLAudit(ch <-chan AuditMsg, url string, logger *zap.SugaredLogger) {
	// Create a new resty client.
	client := resty.New().
		SetBaseURL(url).
		SetTimeout(time.Duration(10) * time.Second)

	// Wait for messages from the channel and send them to the URL.
	for msg := range ch {
		logger.Debugf("received message from event channel: %v", msg)
		// Send the message to the URL.
		if _, err := client.R().
			SetBody(msg.Metrics).
			Post("/"); err != nil {
			logger.Errorf("failed to send message to URL: %v", err)
		}
		logger.Debugf("message sent to URL: %v", msg.Metrics)
	}
}
