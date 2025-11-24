package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	agentcfg "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	repo "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	sign "github.com/devize-ed/yapracproj-metrics.git/internal/sign/config"
)

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	fp := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	return fp
}

func TestGetServerConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupFileJSON  string
		envVars        map[string]string
		args           []string
		expectedConfig ServerConfig
		wantErr        bool
	}{
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":           "localhost:8081",
				"STORE_INTERVAL":    "500",
				"FILE_STORAGE_PATH": "./test.json",
				"RESTORE":           "false",
				"DATABASE_DSN":      "user:password@/dbname",
				"LOG_LEVEL":         "error",
				"KEY":               "test_key",
			},
			args: []string{"-a=:7070", "-i=400", "-f=./non.json", "-d=user:password@/dbname", "-r=true"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: "localhost:8081", GRPCHost: "localhost:3200"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 500,
						FPath:         "./test.json",
						Restore:       false,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "user:password@/dbname",
					},
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				LogLevel: "error",
			},
			wantErr: false,
		},
		{
			name:    "CLI flags (no env, no file)",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-i=400", "-f=./test1.json", "-d=user:password@/dbname", "-r=false", "-k=test2_key"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: ":7070", GRPCHost: "localhost:3200"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 400,
						FPath:         "./test1.json",
						Restore:       false,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "user:password@/dbname",
					},
				},
				Sign: sign.SignConfig{
					Key: "test2_key",
				},
				LogLevel: "",
			},
			wantErr: false,
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: "localhost:8080", GRPCHost: "localhost:3200"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 0,
						FPath:         "",
						Restore:       false,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "",
					},
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				LogLevel: "",
			},
			wantErr: false,
		},
		{
			name: "negative StoreInterval",
			envVars: map[string]string{
				"ADDRESS":           ":8081",
				"FILE_STORAGE_PATH": "./test.json",
				"RESTORE":           "false",
				"LOG_LEVEL":         "error",
			},
			args: []string{"-a=:7070", "-i=-1", "-f=./non.json", "-d=user:password@/dbname", "-r=false"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: ":7070", GRPCHost: "localhost:3200"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: -1,
						FPath:         "./non.json",
						Restore:       false,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "user:password@/dbname",
					},
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				LogLevel: "",
			},
			wantErr: true,
		},
		{
			name: "Config file from env",
			setupFileJSON: `{
				"connection": {"host": "localhost:9090"},
				"repository": {
					"fs": {"store_interval": 1, "file_storage_path": "test.json", "restore": true},
					"db": {"database_dsn": ""}
				},
				"sign": {"key": "test_key"},
				"log_level": "debug"
			}`,
			envVars: map[string]string{
				"CONFIG":            "",
				"ADDRESS":           ":9999",
				"STORE_INTERVAL":    "3",
				"FILE_STORAGE_PATH": "test.json",
				"RESTORE":           "true",
				"DATABASE_DSN":      "",
				"KEY":               "test_key",
				"LOG_LEVEL":         "warn",
			},
			args: []string{"-a=:8088", "-i=2", "-f=test.json", "-d=test.db", "-r=false", "-k=test_key"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: ":9999"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 3,
						FPath:         "test.json",
						Restore:       true,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "",
					},
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				LogLevel: "warn",
			},
			wantErr: false,
		},
		{
			name: "Config file from flag",
			setupFileJSON: `{
				"connection": {"host": "localhost:8085"},
				"repository": {
					"fs": {"store_interval": 10, "file_storage_path": "test.json", "restore": false},
					"db": {"database_dsn": ""}
				},
				"sign": {"key": "test_key"},
				"log_level": "debug"
			}`,
			envVars: map[string]string{
				"ADDRESS": ":5050",
			},
			args: []string{"-c", "configpath.json", "-a=:6060", "-i=11", "-f=test.json", "-d=test.db", "-r=true", "-k=test_key"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: ":5050"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 11,
						FPath:         "test.json",
						Restore:       true,
					},
					DBConfig: db.DBConfig{
						DatabaseDSN: "test.db",
					},
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				LogLevel: "debug",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, k := range []string{
				"ADDRESS", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN",
				"LOG_LEVEL", "KEY", "CONFIG", "CRYPTO_KEY", "AUDIT_FILE", "AUDIT_URL",
			} {
				t.Setenv(k, "")
			}

			if tc.setupFileJSON != "" {
				fp := writeTempJSON(t, tc.setupFileJSON)
				if _, has := tc.envVars["CONFIG"]; has {
					tc.envVars["CONFIG"] = fp
				} else {
					for i := range tc.args {
						if tc.args[i] == "configpath.json" {
							tc.args[i] = fp
						}
					}
				}
			}

			for key, val := range tc.envVars {
				t.Setenv(key, val)
			}

			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := GetServerConfig()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedConfig, cfg)
		})
	}
}

func TestGetAgentConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupFileJSON  string
		envVars        map[string]string
		args           []string
		expectedConfig AgentConfig
		wantErr        bool
	}{
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":         "localhost:8081",
				"REPORT_INTERVAL": "5",
				"POLL_INTERVAL":   "1",
				"LOG_LEVEL":       "debug",
				"ENABLE_GZIP":     "true",
				"ENABLE_TEST_GET": "true",
				"KEY":             "test_key",
				"RATE_LIMIT":      "10",
			},
			args: []string{"-a=:7070", "-r=30", "-p=10", "--gzip=false", "-g=false", "-l=5"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: "localhost:8081", GRPCHost: "localhost:3200"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     true,
					EnableTestGet:  true,
					RateLimit:      10,
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				ShutdownTimeout: 5,
				LogLevel:        "debug",
			},
			wantErr: false,
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-r=5", "-p=1", "--gzip=false", "-g=false", "-k=test_key", "-l=5"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":7070", GRPCHost: "localhost:3200"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     false,
					EnableTestGet:  false,
					RateLimit:      5,
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				ShutdownTimeout: 5,
				LogLevel:        "",
			},
			wantErr: false,
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: "localhost:8080", GRPCHost: "localhost:3200"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 10,
					PollInterval:   2,
					EnableGzip:     true,
					EnableTestGet:  false,
					RateLimit:      10,
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				ShutdownTimeout: 5,
				LogLevel:        "",
			},
			wantErr: false,
		},
		{
			name: "negative PollInterval",
			envVars: map[string]string{
				"ADDRESS":         ":8081",
				"LOG_LEVEL":       "error",
				"ENABLE_GZIP":     "true",
				"ENABLE_TEST_GET": "true",
			},
			args: []string{"-a=:7070", "-r=30", "-p=-1", "--gzip=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":7070", GRPCHost: "localhost:3200"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 30,
					PollInterval:   -1,
					EnableGzip:     false,
					EnableTestGet:  false,
					RateLimit:      10,
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				ShutdownTimeout: 5,
				LogLevel:        "",
			},
			wantErr: true,
		},
		{
			name: "Precedence: file < flags < env (CONFIG via env)",
			setupFileJSON: `{
				"connection": {"host": "from-file:8085"},
				"agent": {
					"report_interval": 11,
					"poll_interval": 4,
					"enable_gzip": false,
					"enable_get_metrics": true,
					"rate_limit": 4
				},
				"sign": {"key": "filekey"},
				"log_level": "info"
			}`,
			envVars: map[string]string{
				"CONFIG":          "",
				"ADDRESS":         ":9091",
				"REPORT_INTERVAL": "13",
				"POLL_INTERVAL":   "6",
				"ENABLE_GZIP":     "false",
				"ENABLE_TEST_GET": "true",
				"RATE_LIMIT":      "6",
				"KEY":             "envkey",
				"LOG_LEVEL":       "error",
			},
			args: []string{"-a=:7070", "-r=12", "-p=5", "--gzip=true", "-g=false", "-l=5", "-k=flagkey"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":9091"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 13,
					PollInterval:   6,
					EnableGzip:     false,
					EnableTestGet:  true,
					RateLimit:      6,
				},
				Sign: sign.SignConfig{
					Key: "envkey",
				},
				ShutdownTimeout: 5,
				LogLevel:        "error",
			},
			wantErr: false,
		},
		{
			name: "Precedence: file < flags < env (config via -c flag)",
			setupFileJSON: `{
				"connection": {"host": "file-host:8082"},
				"agent": {
					"report_interval": 20,
					"poll_interval": 7,
					"enable_gzip": true,
					"enable_get_metrics": false,
					"rate_limit": 2
				},
				"sign": {"key": "filek"},
				"log_level": "debug"
			}`,
			envVars: map[string]string{
				"ADDRESS": ":8000",
			},
			args: []string{"-c", "WILL_BE_REPLACED", "-a=:7000", "-r=21", "-p=8", "--gzip=false", "-g=true", "-l=3", "-k=flagk"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":8000"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 21,
					PollInterval:   8,
					EnableGzip:     false,
					EnableTestGet:  true,
					RateLimit:      3,
				},
				Sign: sign.SignConfig{
					Key: "flagk",
				},
				ShutdownTimeout: 5,
				LogLevel:        "debug",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { // Clear environment variables
			for _, k := range []string{
				"ADDRESS", "REPORT_INTERVAL", "POLL_INTERVAL", "LOG_LEVEL",
				"ENABLE_GZIP", "ENABLE_TEST_GET",
				"KEY", "RATE_LIMIT", "CONFIG", "CRYPTO_KEY", "SHUTDOWN_TIMEOUT",
			} {
				t.Setenv(k, "")
			}

			if tc.setupFileJSON != "" {
				fp := writeTempJSON(t, tc.setupFileJSON)
				if _, has := tc.envVars["CONFIG"]; has {
					tc.envVars["CONFIG"] = fp
				} else {
					for i := range tc.args {
						if tc.args[i] == "WILL_BE_REPLACED" {
							tc.args[i] = fp
						}
					}
				}
			}

			for key, val := range tc.envVars {
				t.Setenv(key, val)
			}

			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := GetAgentConfig()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedConfig, cfg)
		})
	}
}

func TestGetValueOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", "N/A"},
		{"non-empty string (version)", "1.2.3", "1.2.3"},
		{"non-empty string (commit)", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBuildTag(tt.input)
			if got != tt.want {
				t.Errorf("GetBuildTag() = %q, want %q", got, tt.want)
			}
		})
	}
}
