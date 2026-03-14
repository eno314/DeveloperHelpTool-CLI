package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateAmidakujiBoard(t *testing.T) {
	numParticipants := 4
	board := generateAmidakujiBoard(numParticipants)

	// Ensure there are at least 10 steps
	if len(board) < 10 {
		t.Errorf("Expected at least 10 steps, got %d", len(board))
	}

	// Ensure each row has numParticipants-1 elements
	for i, row := range board {
		if len(row) != numParticipants-1 {
			t.Errorf("Expected row %d to have %d elements, got %d", i, numParticipants-1, len(row))
		}
	}

	// Ensure no two adjacent horizontal lines in the same row
	for i, row := range board {
		for j := 0; j < len(row)-1; j++ {
			if row[j] && row[j+1] {
				t.Errorf("Found adjacent horizontal lines in row %d at positions %d and %d", i, j, j+1)
			}
		}
	}
}

func TestEvaluateAmidakuji(t *testing.T) {
	participants := []string{"A", "B", "C"}
	goals := []string{"X", "Y", "Z"}

	// Custom board for testing
	// A B C
	// |-| |
	// | |-|
	// |-| |
	// X Y Z
	board := [][]bool{
		{true, false},
		{false, true},
		{true, false},
	}

	// Path of A: start 0 -> step 1 (move to 1) -> step 2 (move to 2) -> step 3 (stay 2) -> goal Z
	// Path of B: start 1 -> step 1 (move to 0) -> step 2 (stay 0) -> step 3 (move to 1) -> goal Y
	// Path of C: start 2 -> step 1 (stay 2) -> step 2 (move to 1) -> step 3 (move to 0) -> goal X

	results := evaluateAmidakuji(board, participants, goals)

	expected := map[string]string{
		"A": "Z",
		"B": "Y",
		"C": "X",
	}

	for p, expectedG := range expected {
		if g, ok := results[p]; !ok || g != expectedG {
			t.Errorf("Participant %s: expected %s, got %s", p, expectedG, g)
		}
	}
}

func TestRenderAmidakuji(t *testing.T) {
	outStream := new(bytes.Buffer)
	participants := []string{"A", "B"}
	goals := []string{"X", "Y"}
	board := [][]bool{
		{true},
		{false},
	}
	results := map[string]string{
		"A": "Y",
		"B": "X",
	}

	renderAmidakuji(outStream, board, participants, goals, results)
	output := outStream.String()

	if !strings.Contains(output, "A    B") {
		t.Errorf("Output missing participants line")
	}
	if !strings.Contains(output, "X    Y") {
		t.Errorf("Output missing goals line")
	}
	if !strings.Contains(output, "Results:") {
		t.Errorf("Output missing Results section")
	}
	if !strings.Contains(output, "A -> Y") {
		t.Errorf("Output missing result for A")
	}
	if !strings.Contains(output, "B -> X") {
		t.Errorf("Output missing result for B")
	}
}
