package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	agentcfg "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	repo "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
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
				"ADDRESS":           ":8081",
				"STORE_INTERVAL":    "500",
				"FILE_STORAGE_PATH": "./test.json",
				"RESTORE":           "false",
				"DATABASE_DSN":      "host=localhost user=postgres password=secret dbname=test sslmode=disable",
				"LOG_LEVEL":         "error",
			},
			args: []string{"-a=:7070", "-i=400", "-f=./non.json", "-d=user:password@/dbname", "-r=false"},
			expectedConfig: ServerConfig{
				Connection: serverConn{Host: ":8081"},
				Repository: repo.RepositoryConfig{
					FileConfig: repo.RepositoryConfig.FileConfig{
						StoreInterval: 500,
						FPath:         "./test.json",
						Restore:       false,
					},
					DBConfig: repo.DBConfig{
						DatabaseDSN: "host=localhost user=postgres password=secret dbname=test sslmode=disable",
					},
				},
				LogLevel: "error",
			},
			wantErr: false,
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-i=400", "-f=./test1.json", "-d=user:password@/dbname", "-r=false"},
			expectedConfig: ServerConfig{
				Connection: serverConn{Host: ":7070"},
				Repository: repo.RepositoryConfig{
					FileConfig: repo.FileRepositoryConfig{
						StoreInterval: 400,
						FPath:         "./test1.json",
						Restore:       false,
					},
					DBConfig: repo.DBRepositoryConfig{
						DatabaseDSN: "user:password@/dbname",
					},
				},
				LogLevel: "debug",
			},
			wantErr: false,
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: ServerConfig{
				Connection: serverConn{Host: "localhost:8080"},
				Repository: repo.RepositoryConfig{
					FileConfig: repo.FileRepositoryConfig{
						StoreInterval: 300,
						FPath:         "",
						Restore:       false,
					},
					DBConfig: repo.DBRepositoryConfig{
						DatabaseDSN: "",
					},
				},
				LogLevel: "debug",
			},
			wantErr: false,
		},
		{
			name: "negative StoreInterval",
			envVars: map[string]string{
				"ADDRESS":           ":8081",
				"STORE_INTERVAL":    "-1",
				"FILE_STORAGE_PATH": "./test.json",
				"RESTORE":           "false",
				"LOG_LEVEL":         "error",
			},
			args: []string{"-a=:7070", "-i=400", "-f=./non.json", "-d=user:password@/dbname", "-r=false"},
			expectedConfig: ServerConfig{
				Connection: serverConn{Host: ":8081"},
				Repository: repo.RepositoryConfig{
					FileConfig: repo.FileRepositoryConfig{
						StoreInterval: 500,
						FPath:         "./test.json",
						Restore:       false,
					},
					DBConfig: repo.DBRepositoryConfig{
						DatabaseDSN: "user:password@/dbname",
					},
				},
				LogLevel: "error",
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

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
				"ADDRESS":            ":8081",
				"REPORT_INTERVAL":    "5",
				"POLL_INTERVAL":      "1",
				"LOG_LEVEL":          "error",
				"ENABLE_GZIP":        "true",
				"ENABLE_GET_METRICS": "true",
			},
			args: []string{"-a=:7070", "-r=30", "-p=10", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: agentConn{Host: ":8081"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     true,
					EnableTestGet:  true,
				},
				LogLevel: "error",
			},
			wantErr: false,
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-r=5", "-p=1", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: agentConn{Host: ":7070"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     false,
					EnableTestGet:  false,
				},
				LogLevel: "debug",
			},
			wantErr: false,
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: AgentConfig{
				Connection: agentConn{Host: ":8080"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 10,
					PollInterval:   2,
					EnableGzip:     true,
					EnableTestGet:  false,
				},
				LogLevel: "debug",
			},
			wantErr: false,
		},
		{
			name: "negative PollInterval",
			envVars: map[string]string{
				"ADDRESS":            ":8081",
				"REPORT_INTERVAL":    "5",
				"POLL_INTERVAL":      "-1",
				"LOG_LEVEL":          "error",
				"ENABLE_GZIP":        "true",
				"ENABLE_GET_METRICS": "true",
			},
			args: []string{"-a=:7070", "-r=30", "-p=10", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Connection: agentConn{Host: ":8081"},
				Agent: agentcfg.AgentConfig{
					ReportInterval: 5,
					PollInterval:   1,
					EnableGzip:     true,
					EnableTestGet:  true,
				},
				LogLevel: "error",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

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

// ...existing code...
