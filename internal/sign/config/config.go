// Package config provides configuration structures for the sign component.
package config

type SignConfig struct {
	Key string `env:"KEY" json:"key"` // Secret key for the Hash.
}
