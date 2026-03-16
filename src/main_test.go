package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStatus int
		expectedOutput string
	}{
		{
			name:           "No arguments",
			args:           []string{"developer-help-tool-cli"},
			expectedStatus: 0,
			expectedOutput: "Developer Help Tool CLI\n\nUsage:\n  developer-help-tool-cli [command]\n\nAvailable Commands:\n  amidakuji   Generate an Amidakuji (Ghost Leg) lottery\n  httpdiff    Compare HTTP responses from two hosts\n\nFlags:\n",
		},
		{
			name:           "Help flag (-h)",
			args:           []string{"developer-help-tool-cli", "-h"},
			expectedStatus: 0,
			expectedOutput: "Developer Help Tool CLI\n\nUsage:\n  developer-help-tool-cli [command]\n\nAvailable Commands:\n  amidakuji   Generate an Amidakuji (Ghost Leg) lottery\n  httpdiff    Compare HTTP responses from two hosts\n\nFlags:\n",
		},
		{
			name:           "Unknown command",
			args:           []string{"developer-help-tool-cli", "unknown"},
			expectedStatus: 1,
			expectedOutput: "Unknown command: unknown\nDeveloper Help Tool CLI\n\nUsage:\n  developer-help-tool-cli [command]\n\nAvailable Commands:\n  amidakuji   Generate an Amidakuji (Ghost Leg) lottery\n  httpdiff    Compare HTTP responses from two hosts\n\nFlags:\n",
		},
		{
			name:           "amidakuji without args",
			args:           []string{"developer-help-tool-cli", "amidakuji"},
			expectedStatus: 1,
			expectedOutput: "Error: Both --participants and --goals are required.",
		},
		{
			name:           "amidakuji length mismatch",
			args:           []string{"developer-help-tool-cli", "amidakuji", "--participants", "A,B", "--goals", "X"},
			expectedStatus: 1,
			expectedOutput: "Error: The number of participants (2) must match the number of goals (1).",
		},
		{
			name:           "amidakuji insufficient participants",
			args:           []string{"developer-help-tool-cli", "amidakuji", "--participants", "A", "--goals", "X"},
			expectedStatus: 1,
			expectedOutput: "Error: At least 2 participants are required.",
		},
		{
			name:           "amidakuji empty names",
			args:           []string{"developer-help-tool-cli", "amidakuji", "--participants", "A,", "--goals", "X,Y"},
			expectedStatus: 1,
			expectedOutput: "Error: Participant names cannot be empty.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			status := run(tt.args, outStream, errStream)

			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}

			if !strings.Contains(errStream.String(), tt.expectedOutput) {
				t.Errorf("expected output to contain %q, got %q", tt.expectedOutput, errStream.String())
			}
		})
	}
}
