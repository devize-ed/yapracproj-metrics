package db

type DBConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"` //Data Source Name for the database connection.
}
