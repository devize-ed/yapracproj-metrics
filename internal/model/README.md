# internal/model

This package defines the data structures for metrics.

## Data Structures

### Metrics

The main `Metrics` struct represents a metric with:
- `ID`: Unique identifier for the metric
- `MType`: Type of metric (counter or gauge)
- `Delta`: Value for counter metrics (pointer to distinguish 0 from unset)
- `Value`: Value for gauge metrics (pointer to distinguish 0 from unset)
- `Hash`: Optional hash for integrity verification


