package fstorage

type FStorageConfig struct {
	StoreInterval int    `env:"STORE_INTERVAL" json:"store_interval"`
	FPath         string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Restore       bool   `env:"RESTORE" json:"restore"`
}
