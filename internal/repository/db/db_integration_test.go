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

	"github.com/jackc/pgerrcode"
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

func TestSave(t *testing.T) {
	dsn := getDSN()
	cases := []struct {
		Name       string
		InGauges   map[string]float64
		InCounters map[string]int64
		WantErr    bool
	}{
		{
			Name: "save_data",
			InCounters: map[string]int64{
				"test_counter": 5,
			},
			InGauges: map[string]float64{
				"test_gauge": 3.14,
			},
			WantErr: false,
		},
	}

	db, err := NewDB(context.Background(), dsn)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.Name), func(t *testing.T) {
			err := db.Save(context.Background(), tc.InGauges, tc.InCounters)
			assert.NoError(t, err, "error saving Data: %w", err)
		})
	}
}

func TestLoad(t *testing.T) {
	dsn := getDSN()
	cases := []struct {
		name        string
		expGauges   map[string]float64
		expCounters map[string]int64
		wantErr     bool
		expErr      string
	}{
		{
			name: "load_data",
			expGauges: map[string]float64{
				"test_gauge": 3.14,
			},
			expCounters: map[string]int64{
				"test_counter": 5,
			},
			wantErr: false,
			expErr:  "",
		},
		{
			name:        "load_data_empty",
			expGauges:   map[string]float64{},
			expCounters: map[string]int64{},
			wantErr:     true,
			expErr:      pgerrcode.NoData,
		},
	}

	db, err := NewDB(context.Background(), dsn)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	for i, tc := range cases {
		i, tc := i, tc

		t.Run(fmt.Sprintf("test #%d: %s", i, tc.name), func(t *testing.T) {
			counters, gauges, err := db.Load(context.Background())
			if !tc.wantErr {
				assert.NoError(t, err, "error loading data: %w", err)
				assert.Equal(t, counters, tc.expCounters, "expected counters do not match")
				assert.Equal(t, gauges, tc.expGauges, "expected gauges do not match")
				return
			} else {
				assert.Error(t, err, "expected error but got none")
				assert.Equal(t, counters, tc.expCounters, "expected counters do not match")
				assert.Equal(t, gauges, tc.expGauges, "expected gauges do not match")
				if err != nil {
					assert.Equal(t, tc.expErr, err.Error(), "expected error does not match")
				}
			}
		})
	}
}
