package repository

import (
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
)

type RepositoryConfig struct {
	FSConfig fs.FStorageConfig // Configuration for file storage.
	DBConfig db.DBConfig       // Configuration for database storage.
}
