package features

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

func RunAmidakuji(args []string, outStream, errStream io.Writer) int {
	participants, goals, err := parseAndValidateAmidakujiArgs(args, errStream)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := generateAmidakujiBoard(len(participants), r)
	results := evaluateAmidakuji(board, participants, goals)
	renderAmidakuji(outStream, board, participants, goals, results)

	return 0
}

func parseAndValidateAmidakujiArgs(args []string, errStream io.Writer) ([]string, []string, error) {
	flags, participantsStr, goalsStr := setupAmidakujiFlags(errStream)

	if err := flags.Parse(args); err != nil {
		return nil, nil, err
	}

	if *participantsStr == "" || *goalsStr == "" {
		fmt.Fprintf(errStream, "Error: Both --participants and --goals are required.\n")
		flags.Usage()
		return nil, nil, fmt.Errorf("missing flags")
	}

	return validateParticipantsAndGoals(*participantsStr, *goalsStr, errStream)
}

func setupAmidakujiFlags(errStream io.Writer) (*flag.FlagSet, *string, *string) {
	flags := flag.NewFlagSet("amidakuji", flag.ContinueOnError)
	flags.SetOutput(errStream)

	participantsStr := flags.String("participants", "", "Comma-separated list of participants")
	goalsStr := flags.String("goals", "", "Comma-separated list of goals")

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Usage: amidakuji --participants A,B,C --goals X,Y,Z\n\nFlags:\n")
		flags.PrintDefaults()
	}
	return flags, participantsStr, goalsStr
}

func validateParticipantsAndGoals(pStr, gStr string, errStream io.Writer) ([]string, []string, error) {
	participants := splitAndTrim(pStr)
	goals := splitAndTrim(gStr)

	if err := validateLengths(participants, goals, errStream); err != nil {
		return nil, nil, err
	}

	if err := validateUniqueAndNotEmpty(participants, errStream); err != nil {
		return nil, nil, err
	}

	if err := validateNotEmpty(goals, errStream); err != nil {
		return nil, nil, err
	}

	return participants, goals, nil
}

func splitAndTrim(str string) []string {
	parts := strings.Split(str, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func validateLengths(participants, goals []string, errStream io.Writer) error {
	if len(participants) < 2 {
		fmt.Fprintf(errStream, "Error: At least 2 participants are required.\n")
		return fmt.Errorf("insufficient participants")
	}
	if len(participants) != len(goals) {
		fmt.Fprintf(errStream, "Error: The number of participants (%d) must match the number of goals (%d).\n", len(participants), len(goals))
		return fmt.Errorf("length mismatch")
	}
	return nil
}

func validateUniqueAndNotEmpty(names []string, errStream io.Writer) error {
	seen := make(map[string]bool)
	for _, name := range names {
		if name == "" {
			fmt.Fprintf(errStream, "Error: Participant names cannot be empty.\n")
			return fmt.Errorf("empty name")
		}
		if seen[name] {
			fmt.Fprintf(errStream, "Error: Duplicate participant name found: %s\n", name)
			return fmt.Errorf("duplicate name")
		}
		seen[name] = true
	}
	return nil
}

func validateNotEmpty(names []string, errStream io.Writer) error {
	for _, name := range names {
		if name == "" {
			fmt.Fprintf(errStream, "Error: Goal names cannot be empty.\n")
			return fmt.Errorf("empty name")
		}
	}
	return nil
}

func generateAmidakujiBoard(numParticipants int, r *rand.Rand) [][]bool {
	steps := calculateSteps(numParticipants)
	board := makeEmptyBoard(steps, numParticipants-1)
	populateRandomLines(board, numParticipants, r)
	return board
}

func calculateSteps(numParticipants int) int {
	steps := numParticipants * 2
	if steps < 10 {
		return 10
	}
	return steps
}

func makeEmptyBoard(steps, width int) [][]bool {
	board := make([][]bool, steps)
	for s := 0; s < steps; s++ {
		board[s] = make([]bool, width)
	}
	return board
}

func populateRandomLines(board [][]bool, numParticipants int, r *rand.Rand) {
	for s := range board {
		populateRow(board[s], numParticipants-1, r)
	}
}

func populateRow(row []bool, width int, r *rand.Rand) {
	for l := 0; l < width; l++ {
		if r.Intn(3) == 0 {
			row[l] = true
			l++
		}
	}
}

func evaluateAmidakuji(board [][]bool, participants []string, goals []string) map[string]string {
	results := make(map[string]string)
	for pIdx, participant := range participants {
		endLine := traceParticipant(board, pIdx, len(participants))
		results[participant] = goals[endLine]
	}
	return results
}

func traceParticipant(board [][]bool, startLine, numParticipants int) int {
	curr := startLine
	for _, row := range board {
		curr = stepNextLine(row, curr, numParticipants)
	}
	return curr
}

func stepNextLine(row []bool, curr, numParticipants int) int {
	if curr > 0 && row[curr-1] {
		return curr - 1
	}
	if curr < numParticipants-1 && row[curr] {
		return curr + 1
	}
	return curr
}

func renderAmidakuji(outStream io.Writer, board [][]bool, participants []string, goals []string, results map[string]string) {
	colWidth := calculateColWidth(participants, goals)
	paddingLeft := colWidth / 2
	paddingRight := colWidth - paddingLeft - 1

	renderParticipants(outStream, participants, colWidth)
	renderTopSpacers(outStream, len(participants), paddingLeft, paddingRight)
	renderBoard(outStream, board, len(participants), paddingLeft, paddingRight)
	renderGoals(outStream, goals, colWidth)
	renderResultsText(outStream, participants, results)
}

func calculateColWidth(participants, goals []string) int {
	maxLen := 3
	for _, p := range participants {
		maxLen = maxInt(maxLen, len(p))
	}
	for _, g := range goals {
		maxLen = maxInt(maxLen, len(g))
	}
	return maxLen + 2
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderParticipants(outStream io.Writer, participants []string, colWidth int) {
	for _, p := range participants {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", p))
	}
	fmt.Fprintf(outStream, "\n")
}

func renderTopSpacers(outStream io.Writer, numParticipants, paddingLeft, paddingRight int) {
	for i := 0; i < numParticipants; i++ {
		fmt.Fprintf(outStream, "%*s|%*s", paddingLeft, "", paddingRight, "")
	}
	fmt.Fprintf(outStream, "\n")
}

func renderGoals(outStream io.Writer, goals []string, colWidth int) {
	for _, g := range goals {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", g))
	}
	fmt.Fprintf(outStream, "\n\n")
}

func renderResultsText(outStream io.Writer, participants []string, results map[string]string) {
	fmt.Fprintf(outStream, "Results:\n")
	for _, p := range participants {
		fmt.Fprintf(outStream, "  %s -> %s\n", p, results[p])
	}
}

func renderBoard(outStream io.Writer, board [][]bool, numParticipants, paddingLeft, paddingRight int) {
	for _, row := range board {
		renderBoardRow(outStream, row, numParticipants, paddingLeft, paddingRight)
		renderSpacerRow(outStream, numParticipants, paddingLeft, paddingRight)
	}
}

func renderBoardRow(outStream io.Writer, row []bool, numParticipants, paddingLeft, paddingRight int) {
	fmt.Fprintf(outStream, "%*s", paddingLeft, "")
	for i := 0; i < numParticipants; i++ {
		fmt.Fprintf(outStream, "|")
		renderRowConnection(outStream, row, i, numParticipants, paddingLeft, paddingRight)
	}
	fmt.Fprintf(outStream, "\n")
}

func renderRowConnection(outStream io.Writer, row []bool, index, numParticipants, paddingLeft, paddingRight int) {
	if index >= numParticipants-1 {
		fmt.Fprintf(outStream, "%*s", paddingRight, "")
		return
	}
	connectionLen := paddingRight + paddingLeft
	if row[index] {
		fmt.Fprintf(outStream, "%s", strings.Repeat("-", connectionLen))
		return
	}
	fmt.Fprintf(outStream, "%s", strings.Repeat(" ", connectionLen))
}

func renderSpacerRow(outStream io.Writer, numParticipants, paddingLeft, paddingRight int) {
	for i := 0; i < numParticipants; i++ {
		fmt.Fprintf(outStream, "%*s|%*s", paddingLeft, "", paddingRight, "")
	}
	fmt.Fprintf(outStream, "\n")
}
