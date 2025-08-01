package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(DSN string) error {
	m, err := migrate.New(
		"file://migrations",
		DSN,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	return m.Up()
}
