package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	agentcfg "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	repo "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	sign "github.com/devize-ed/yapracproj-metrics.git/internal/sign/config"
)

func TestGetServerConfig(t *testing.T) {
	tests := []struct {
		name           string
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
				Connection: ServerConn{Host: "localhost:8081"},
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
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-i=400", "-f=./test1.json", "-d=user:password@/dbname", "-r=false", "-k=test2_key"},
			expectedConfig: ServerConfig{
				Connection: ServerConn{Host: ":7070"},
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
				Connection: ServerConn{Host: "localhost:8080"},
				Repository: repo.RepositoryConfig{
					FSConfig: fs.FStorageConfig{
						StoreInterval: 300,
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
				Connection: ServerConn{Host: ":7070"},
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

			for _, k := range []string{
				"ADDRESS", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN",
				"LOG_LEVEL", "KEY",
			} {
				t.Setenv(k, "")
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
		envVars        map[string]string
		args           []string
		expectedConfig AgentConfig
		wantErr        bool
	}{
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":            "localhost:8081",
				"REPORT_INTERVAL":    "5",
				"POLL_INTERVAL":      "1",
				"LOG_LEVEL":          "error",
				"ENABLE_GZIP":        "true",
				"ENABLE_GET_METRICS": "true",
				"KEY":                "test_key",
			},
			args: []string{"-a=:7070", "-r=30", "-p=10", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: "localhost:8081"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     true,
					EnableTestGet:  true,
				},
				Sign: sign.SignConfig{
					Key: "test_key",
				},
				LogLevel: "error",
			},
			wantErr: false,
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-r=5", "-p=1", "-c=false", "-g=false", "-k=test2_key"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":7070"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     false,
					EnableTestGet:  false,
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
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: "localhost:8080"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 10,
					PollInterval:   2,
					EnableGzip:     true,
					EnableTestGet:  false,
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				LogLevel: "",
			},
			wantErr: false,
		},
		{
			name: "negative PollInterval",
			envVars: map[string]string{
				"ADDRESS":            ":8081",
				"LOG_LEVEL":          "error",
				"ENABLE_GZIP":        "true",
				"ENABLE_GET_METRICS": "true",
			},
			args: []string{"-a=:7070", "-r=30", "-p=-1", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: AgentConn{Host: ":7070"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 30,
					PollInterval:   -1,
					EnableGzip:     false,
					EnableTestGet:  false,
				},
				Sign: sign.SignConfig{
					Key: "",
				},
				LogLevel: "",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

			for _, k := range []string{
				"ADDRESS", "REPORT_INTERVAL", "POLL_INTERVAL", "LOG_LEVEL",
				"ENABLE_GZIP", "ENABLE_GET_METRICS",
			} {
				t.Setenv(k, "")
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
