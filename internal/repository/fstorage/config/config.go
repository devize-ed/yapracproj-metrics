package fstorage

type FStorageConfig struct {
	StoreInterval int    `env:"STORE_INTERVAL"`
	FPath         string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
}
