//go:build integration_tests
// +build integration_tests

package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	cfg "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	"github.com/jackc/pgx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	code, err := runMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

const (
	testDBName       = "test"
	testUserName     = "test"
	testUserPassword = "test"
)

var (
	getDSN          func() string
	getSUConnection func() (*pgx.Conn, error)
)

func initGetDSN(hostAndPort string) {
	getDSN = func() string {
		return fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			testUserName,
			testUserPassword,
			hostAndPort,
			testDBName,
		)
	}
}

func initGetSUConnection(hostPort string) error {
	host, port, err := getHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("failed to extract the host and port parts from the string %s: %w", hostPort, err)
	}
	getSUConnection = func() (*pgx.Conn, error) {
		conn, err := pgx.Connect(pgx.ConnConfig{
			Host:     host,
			Port:     port,
			Database: "postgres",
			User:     "postgres",
			Password: "postgres",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get a super user connection: %w", err)
		}
		return conn, nil
	}
	return nil
}

func runMain(m *testing.M) (int, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 1, fmt.Errorf("failed to initialize a pool: %w", err)
	}

	pg, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "17.2",
			Name:       "migrations-integration-tests",
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{"5432/tcp"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return 1, fmt.Errorf("failed to run the postgres container: %w", err)
	}

	defer func() {
		if err := pool.Purge(pg); err != nil {
			log.Printf("failed to purge the postgres container: %v", err)
		}
	}()

	hostPort := pg.GetHostPort("5432/tcp")
	initGetDSN(hostPort)
	if err := initGetSUConnection(hostPort); err != nil {
		return 1, err
	}

	pool.MaxWait = 10 * time.Second
	var conn *pgx.Conn
	if err := pool.Retry(func() error {
		conn, err = getSUConnection()
		if err != nil {
			return fmt.Errorf("failed to connect to the DB: %w", err)
		}
		return nil
	}); err != nil {
		return 1, err
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to correctly close the connection: %v", err)
		}
	}()

	if err := createTestDB(conn); err != nil {
		return 1, fmt.Errorf("failed to create a test DB: %w", err)
	}

	exitCode := m.Run()

	return exitCode, nil
}

func createTestDB(conn *pgx.Conn) error {
	_, err := conn.Exec(
		fmt.Sprintf(
			`CREATE USER %s PASSWORD '%s'`,
			testUserName,
			testUserPassword,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create a test user: %w", err)
	}

	_, err = conn.Exec(
		fmt.Sprintf(`
			CREATE DATABASE %s
				OWNER '%s'
				ENCODING 'UTF8'
				LC_COLLATE = 'en_US.utf8'
				LC_CTYPE = 'en_US.utf8'
			`, testDBName, testUserName,
		),
	)

	if err != nil {
		return fmt.Errorf("failed to create a test DB: %w", err)
	}

	return nil
}

func getHostPort(hostPort string) (string, uint16, error) {
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		return "", 0, fmt.Errorf("got an invalid host-port string: %s", hostPort)
	}

	portStr := hostPortParts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to cast the port %s to an int: %w", portStr, err)
	}
	return hostPortParts[0], uint16(port), nil
}

func TestAddCounter(t *testing.T) {
	testCounter := int64(1)
	testGauge := float64(1.11)

	cases := []struct {
		Name        string
		InMetrics   *models.Metrics
		wantErr     bool
		ExpectedErr string
	}{
		{
			Name: "add_counter",
			InMetrics: &models.Metrics{
				ID:    "testCounter",
				Delta: &testCounter,
			},
			wantErr:     false,
			ExpectedErr: "",
		},
		{
			Name: "add_wrong_counter",
			InMetrics: &models.Metrics{
				ID:    "testCounter",
				Value: &testGauge,
			},
			wantErr:     true,
			ExpectedErr: "null value in column \"delta\"",
		},
	}

	db, err := NewDB(context.Background(), &cfg.DBConfig{
		DatabaseDSN: getDSN(),
	})
	if err != nil {
		t.Fatalf("failed to create a DB: %v", err)
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			err := db.AddCounter(context.Background(), tc.InMetrics.ID, tc.InMetrics.Delta)
			if tc.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ExpectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetGauge(t *testing.T) {
	testCounter := int64(1)
	testGauge := float64(1.11)

	cases := []struct {
		Name        string
		InMetrics   *models.Metrics
		wantErr     bool
		ExpectedErr string
	}{
		{
			Name: "set_gauge",
			InMetrics: &models.Metrics{
				ID:    "testGauge",
				Value: &testGauge,
			},
			wantErr:     false,
			ExpectedErr: "",
		},
		{
			Name: "set_wrong_gauge",
			InMetrics: &models.Metrics{
				ID:    "testGauge",
				Delta: &testCounter,
			},
			wantErr:     true,
			ExpectedErr: "null value in column \"value\"",
		},
	}

	db, err := NewDB(context.Background(), &cfg.DBConfig{
		DatabaseDSN: getDSN(),
	})
	if err != nil {
		t.Fatalf("failed to create a DB: %v", err)
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			err := db.SetGauge(context.Background(), tc.InMetrics.ID, tc.InMetrics.Value)
			if tc.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ExpectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGauge(t *testing.T) {
	testGauge := float64(1.11)

	cases := []struct {
		Name            string
		ID              string
		ExpectedMetrics *models.Metrics
		wantErr         bool
		ExpectedErr     string
	}{
		{
			Name: "get_gauge",
			ID:   "testGauge",
			ExpectedMetrics: &models.Metrics{
				ID:    "testGauge",
				Value: &testGauge,
			},
			wantErr:     false,
			ExpectedErr: "",
		},
		{
			Name: "get_unexisting_gauge",
			ID:   "unexistingGauge",
			ExpectedMetrics: &models.Metrics{
				ID:    "unexistingGauge",
				Value: &testGauge,
			},
			wantErr:     true,
			ExpectedErr: "not found",
		},
	}

	db, err := NewDB(context.Background(), &cfg.DBConfig{
		DatabaseDSN: getDSN(),
	})
	if err != nil {
		t.Fatalf("failed to create a DB: %v", err)
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			got, err := db.GetGauge(context.Background(), tc.ID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ExpectedErr)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, 1.11, *got, 1e-9)
			}
		})
	}
}

func TestGetCounter(t *testing.T) {
	testCounter := int64(1)

	cases := []struct {
		Name            string
		ID              string
		ExpectedMetrics *models.Metrics
		wantErr         bool
		ExpectedErr     string
	}{
		{
			Name: "get_counter",
			ID:   "testCounter",
			ExpectedMetrics: &models.Metrics{
				ID: "testCounter",
			},
			wantErr:     false,
			ExpectedErr: "",
		},
		{
			Name: "get_unexisting_counter",
			ID:   "unexistingCounter",
			ExpectedMetrics: &models.Metrics{
				ID: "unexistingCounter",
			},
			wantErr:     true,
			ExpectedErr: "not found",
		},
	}

	db, err := NewDB(context.Background(), &cfg.DBConfig{
		DatabaseDSN: getDSN(),
	})
	if err != nil {
		t.Fatalf("failed to create a DB: %v", err)
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			got, err := db.GetCounter(context.Background(), tc.ID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ExpectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCounter, *got)
			}
		})
	}
}

func TestSaveBatchAndGetAll(t *testing.T) {
	testCounter := int64(5)
	testGauge := float64(5.55)

	cases := []struct {
		Name        string
		Metrics     *[]models.Metrics
		wantErr     bool
		ExpectedErr string
	}{
		{
			Name: "save_batch",
			Metrics: &[]models.Metrics{
				{
					ID:    "testCounter",
					MType: models.Counter,
					Delta: &testCounter,
				},
				{
					ID:    "testGauge",
					MType: models.Gauge,
					Value: &testGauge,
				},
			},
			wantErr:     false,
			ExpectedErr: "",
		},
		{
			Name:        "save_empty_batch",
			Metrics:     &[]models.Metrics{},
			wantErr:     true,
			ExpectedErr: "empty slice",
		},
	}

	db, err := NewDB(context.Background(), &cfg.DBConfig{
		DatabaseDSN: getDSN(),
	})
	if err != nil {
		t.Fatalf("failed to create a DB: %v", err)
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			err := db.SaveBatch(context.Background(), *tc.Metrics)
			if tc.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ExpectedErr)
			} else {
				assert.NoError(t, err)
			}
		})

		want := map[string]string{"testCounter": strconv.FormatInt(testCounter+1, 10), "testGauge": strconv.FormatFloat(testGauge, 'f', -1, 64)}
		t.Run("get_all", func(t *testing.T) {
			got, err := db.GetAll(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}
