package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServerConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		args           []string
		expectedConfig ServerConfig
	}{
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":           ":8081",
				"STORE_INTERVAL":    "500",
				"FILE_STORAGE_PATH": "./test.json",
				"RESTORE":           "false",
				"LOG_LEVEL":         "error",
			},
			args: []string{"-a=:7070", "-i=400", "-f=./non.json", "-r=false"},
			expectedConfig: ServerConfig{
				Host:          ":8081",
				StoreInterval: 500,
				FPath:         "./test.json",
				Restore:       false,
				LogLevel:      "error",
			},
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-i=400", "-f=./test1.json", "-r=false"},
			expectedConfig: ServerConfig{
				Host:          ":7070",
				StoreInterval: 400,
				FPath:         "./test1.json",
				Restore:       false,
				LogLevel:      "debug",
			},
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: ServerConfig{
				Host:          "localhost:8080",
				StoreInterval: 300,
				FPath:         "./metrics_storage.json",
				Restore:       false,
				LogLevel:      "debug",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

			for key, val := range tc.envVars {
				os.Setenv(key, val)
				defer os.Unsetenv(key)
			}

			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := GetServerConfig()
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
				Host:           ":8081",
				ReportInterval: 5,
				PollInterval:   1,
				LogLevel:       "error",
				EnableGzip:     true,
				EnableTestGet:  true,
			},
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-r=5", "-p=1", "-c=false", "-g=false"},
			expectedConfig: AgentConfig{
				Host:           ":7070",
				ReportInterval: 5,
				PollInterval:   1,
				LogLevel:       "debug",
				EnableGzip:     false,
				EnableTestGet:  false,
			},
		},
		{
			name:    "Defaults",
			envVars: map[string]string{},
			args:    []string{},
			expectedConfig: AgentConfig{
				Host:           ":8080",
				ReportInterval: 10,
				PollInterval:   2,
				LogLevel:       "debug",
				EnableGzip:     false,
				EnableTestGet:  false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

			for key, val := range tc.envVars {
				os.Setenv(key, val)
				defer os.Unsetenv(key)
			}

			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := GetAgentConfig()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedConfig, cfg)
		})
	}
}
