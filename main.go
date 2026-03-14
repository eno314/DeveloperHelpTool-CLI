package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

func run(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(errStream)

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Developer Help Tool CLI\n\n")
		fmt.Fprintf(errStream, "Usage:\n")
		fmt.Fprintf(errStream, "  %s [command]\n\n", args[0])
		fmt.Fprintf(errStream, "Available Commands:\n")
		fmt.Fprintf(errStream, "  amidakuji   Generate an Amidakuji (Ghost Leg) lottery\n\n")
		fmt.Fprintf(errStream, "Flags:\n")
		flags.PrintDefaults()
	}

	if len(args) < 2 {
		flags.Usage()
		return 0
	}

	err := flags.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	command := args[1]
	switch command {
	case "amidakuji":
		return runAmidakuji(args[2:], outStream, errStream)
	case "-h", "--help":
		flags.Usage()
		return 0
	default:
		fmt.Fprintf(errStream, "Unknown command: %s\n", command)
		flags.Usage()
		return 1
	}
}

func runAmidakuji(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet("amidakuji", flag.ContinueOnError)
	flags.SetOutput(errStream)

	participantsStr := flags.String("participants", "", "Comma-separated list of participants")
	goalsStr := flags.String("goals", "", "Comma-separated list of goals")

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Usage: amidakuji --participants A,B,C --goals X,Y,Z\n\nFlags:\n")
		flags.PrintDefaults()
	}

	err := flags.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	if *participantsStr == "" || *goalsStr == "" {
		fmt.Fprintf(errStream, "Error: Both --participants and --goals are required.\n")
		flags.Usage()
		return 1
	}

	participants := strings.Split(*participantsStr, ",")
	goals := strings.Split(*goalsStr, ",")

	if len(participants) < 2 {
		fmt.Fprintf(errStream, "Error: At least 2 participants are required.\n")
		return 1
	}

	if len(participants) != len(goals) {
		fmt.Fprintf(errStream, "Error: The number of participants (%d) must match the number of goals (%d).\n", len(participants), len(goals))
		return 1
	}

	// Make sure no empty names exist
	for i, p := range participants {
		participants[i] = strings.TrimSpace(p)
		if participants[i] == "" {
			fmt.Fprintf(errStream, "Error: Participant names cannot be empty.\n")
			return 1
		}
	}
	for i, g := range goals {
		goals[i] = strings.TrimSpace(g)
		if goals[i] == "" {
			fmt.Fprintf(errStream, "Error: Goal names cannot be empty.\n")
			return 1
		}
	}

	board := generateAmidakujiBoard(len(participants))
	results := evaluateAmidakuji(board, participants, goals)
	renderAmidakuji(outStream, board, participants, goals, results)

	return 0
}

func generateAmidakujiBoard(numParticipants int) [][]bool {
	// Let's create a board with enough horizontal steps to make it random.
	// For N participants, maybe max(10, N*2)
	steps := numParticipants * 2
	if steps < 10 {
		steps = 10
	}

	// board[step][line] represents whether there's a horizontal line connecting
	// vertical line `line` and `line+1` at horizontal step `step`.
	board := make([][]bool, steps)
	for s := 0; s < steps; s++ {
		board[s] = make([]bool, numParticipants-1)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for s := 0; s < steps; s++ {
		// Place a horizontal line randomly. Ensure no two adjacent horizontal lines
		// at the same step.
		for l := 0; l < numParticipants-1; l++ {
			// Probability of adding a horizontal line is roughly 30-40%
			if r.Intn(3) == 0 {
				board[s][l] = true
				// Skip the next line to prevent two adjacent horizontal lines meeting
				// at the same node, which is invalid in Amidakuji.
				l++
			}
		}
	}

	return board
}

func evaluateAmidakuji(board [][]bool, participants []string, goals []string) map[string]string {
	results := make(map[string]string)

	for pIdx, participant := range participants {
		currentLine := pIdx
		for _, row := range board {
			// Check left
			if currentLine > 0 && row[currentLine-1] {
				currentLine--
			} else if currentLine < len(participants)-1 && row[currentLine] {
				// Check right
				currentLine++
			}
		}
		results[participant] = goals[currentLine]
	}

	return results
}

func renderAmidakuji(outStream io.Writer, board [][]bool, participants []string, goals []string, results map[string]string) {
	// Padding for each column. We will make them equal width for nice rendering.
	maxLen := 0
	for _, p := range participants {
		if len(p) > maxLen {
			maxLen = len(p)
		}
	}
	for _, g := range goals {
		if len(g) > maxLen {
			maxLen = len(g)
		}
	}

	// Ensure at least 3 chars padding for nicer line drawing
	if maxLen < 3 {
		maxLen = 3
	}

	colWidth := maxLen + 2

	// Render participants
	for _, p := range participants {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", p))
	}
	fmt.Fprintf(outStream, "\n")

	// Helper func to center the vertical lines
	paddingLeft := (colWidth / 2)
	paddingRight := colWidth - paddingLeft - 1

	// Render top vertical lines before the board
	for range participants {
		fmt.Fprintf(outStream, "%*s|%*s", paddingLeft, "", paddingRight, "")
	}
	fmt.Fprintf(outStream, "\n")

	// Render the board
	for _, row := range board {
		// Print initial left padding for the first vertical line
		fmt.Fprintf(outStream, "%*s", paddingLeft, "")

		for i := 0; i < len(participants); i++ {
			// Print vertical line
			fmt.Fprintf(outStream, "|")

			// Determine what connects to the next vertical line (if not the last one)
			if i < len(participants)-1 {
				// We need to fill paddingRight, plus the next paddingLeft, with either spaces or hyphens
				connectionLen := paddingRight + paddingLeft
				if row[i] {
					fmt.Fprintf(outStream, "%s", strings.Repeat("-", connectionLen))
				} else {
					fmt.Fprintf(outStream, "%s", strings.Repeat(" ", connectionLen))
				}
			} else {
				// Last column, just print right padding
				fmt.Fprintf(outStream, "%*s", paddingRight, "")
			}
		}
		fmt.Fprintf(outStream, "\n")

		// Add an empty vertical step between horizontal rows to make it look nicer
		for range participants {
			fmt.Fprintf(outStream, "%*s|%*s", paddingLeft, "", paddingRight, "")
		}
		fmt.Fprintf(outStream, "\n")
	}

	// Render goals
	for _, g := range goals {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", g))
	}
	fmt.Fprintf(outStream, "\n\n")

	// Render results
	fmt.Fprintf(outStream, "Results:\n")
	for _, p := range participants {
		fmt.Fprintf(outStream, "  %s -> %s\n", p, results[p])
	}
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
