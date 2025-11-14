package repository

import (
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
)

type RepositoryConfig struct {
	FSConfig fs.FStorageConfig `json:"fs"` // Configuration for file storage.
	DBConfig db.DBConfig       `json:"db"` // Configuration for database storage.
}
