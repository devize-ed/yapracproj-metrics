package db

type DBConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN"` //Data Source Name for the database connection.
}
