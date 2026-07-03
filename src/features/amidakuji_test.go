package features

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

func TestGenerateAmidakujiBoard(t *testing.T) {
	tests := []struct {
		name            string
		numParticipants int
		seed            int64
	}{
		{"2 participants", 2, 42},
		{"4 participants", 4, 42},
		{"10 participants", 10, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(tt.seed))
			board := generateAmidakujiBoard(tt.numParticipants, r)

			if len(board) < 10 {
				t.Errorf("Expected at least 10 steps, got %d", len(board))
			}

			for i, row := range board {
				if len(row) != tt.numParticipants-1 {
					t.Errorf("Expected row %d to have %d elements, got %d", i, tt.numParticipants-1, len(row))
				}
			}

			for i, row := range board {
				for j := 0; j < len(row)-1; j++ {
					if row[j] && row[j+1] {
						t.Errorf("Found adjacent horizontal lines in row %d at positions %d and %d", i, j, j+1)
					}
				}
			}
		})
	}
}

func TestEvaluateAmidakuji(t *testing.T) {
	tests := []struct {
		name         string
		participants []string
		goals        []string
		board        [][]bool
		expected     map[string]string
	}{
		{
			name:         "3 participants cross",
			participants: []string{"A", "B", "C"},
			goals:        []string{"X", "Y", "Z"},
			board: [][]bool{
				{true, false},
				{false, true},
				{true, false},
			},
			expected: map[string]string{
				"A": "Z",
				"B": "Y",
				"C": "X",
			},
		},
		{
			name:         "2 participants straight",
			participants: []string{"A", "B"},
			goals:        []string{"X", "Y"},
			board: [][]bool{
				{false},
			},
			expected: map[string]string{
				"A": "X",
				"B": "Y",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := evaluateAmidakuji(tt.board, tt.participants, tt.goals)

			for p, expectedG := range tt.expected {
				if g, ok := results[p]; !ok || g != expectedG {
					t.Errorf("Participant %s: expected %s, got %s", p, expectedG, g)
				}
			}
		})
	}
}

func TestRenderAmidakuji(t *testing.T) {
	tests := []struct {
		name         string
		participants []string
		goals        []string
		board        [][]bool
		results      map[string]string
		contains     []string
	}{
		{
			name:         "2 participants cross",
			participants: []string{"A", "B"},
			goals:        []string{"X", "Y"},
			board: [][]bool{
				{true},
				{false},
			},
			results: map[string]string{
				"A": "Y",
				"B": "X",
			},
			contains: []string{
				"A    B",
				"X    Y",
				"Results:",
				"A -> Y",
				"B -> X",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outStream := new(bytes.Buffer)
			renderAmidakuji(outStream, tt.board, tt.participants, tt.goals, tt.results)
			output := outStream.String()

			for _, expectedStr := range tt.contains {
				if !strings.Contains(output, expectedStr) {
					t.Errorf("Output missing expected string: %s\nActual output:\n%s", expectedStr, output)
				}
			}
		})
	}
}
