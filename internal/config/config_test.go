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
		envVars        []string
		args           []string
		expectedConfig ServerConfig
	}{
		{
			name:    "Environment variables",
			envVars: []string{"ADDRESS", ":8081"},
			args:    []string{"-a=:7070"},
			expectedConfig: ServerConfig{
				Host:     ":8081",
				LogLevel: "debug",
			},
		},
		{
			name:    "CLI flags",
			envVars: []string{},
			args:    []string{"-a=:7070"},
			expectedConfig: ServerConfig{
				Host:     ":7070",
				LogLevel: "debug",
			},
		},
		{
			name:    "Defaults",
			envVars: []string{},
			args:    []string{},
			expectedConfig: ServerConfig{
				Host:     "localhost:8080",
				LogLevel: "debug",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)

			if len(tc.envVars) != 0 {
				os.Setenv(tc.envVars[0], tc.envVars[1])
				defer os.Unsetenv(tc.envVars[0])
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
				"ADDRESS":         ":8081",
				"REPORT_INTERVAL": "5",
				"POLL_INTERVAL":   "1",
				"LOG_LEVEL":       "error",
			},
			args: []string{"-a=:7070", "-r=30", "-p=10"},
			expectedConfig: AgentConfig{
				Host:           ":8081",
				ReportInterval: 5,
				PollInterval:   1,
				LogLevel:       "error",
			},
		},
		{
			name:    "CLI flags",
			envVars: map[string]string{},
			args:    []string{"-a=:7070", "-r=5", "-p=1"},
			expectedConfig: AgentConfig{
				Host:           ":7070",
				ReportInterval: 5,
				PollInterval:   1,
				LogLevel:       "debug",
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
