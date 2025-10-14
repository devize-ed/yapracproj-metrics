// Package models defines the data structures for metrics.
// It provides types for counter and gauge metrics with JSON serialization support.
package models

// Metric type constants.
const (
	Counter = "counter" // Counter metric type.
	Gauge   = "gauge"   // Gauge metric type.
)

// Metrics represents a metric with its type, value, and optional hash.
// Delta and Value are declared as pointers to distinguish between "0" and unset values.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // Delta value for counter metrics.
	Value *float64 `json:"value,omitempty"` // Value for gauge metrics.
	Hash  string   `json:"hash,omitempty"` // Optional hash for integrity verification.
}
