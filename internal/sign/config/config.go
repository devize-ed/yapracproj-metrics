// Package config provides configuration structures for the sign component.
package config

type SignConfig struct {
	Key string `env:"KEY"` // Secret key for the Hash.
}
