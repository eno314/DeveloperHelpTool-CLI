package features

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

func RunAmidakuji(args []string, outStream, errStream io.Writer) int {
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

	seenParticipants := make(map[string]bool)
	for i, p := range participants {
		participants[i] = strings.TrimSpace(p)
		if participants[i] == "" {
			fmt.Fprintf(errStream, "Error: Participant names cannot be empty.\n")
			return 1
		}
		if seenParticipants[participants[i]] {
			fmt.Fprintf(errStream, "Error: Duplicate participant name found: %s\n", participants[i])
			return 1
		}
		seenParticipants[participants[i]] = true
	}
	for i, g := range goals {
		goals[i] = strings.TrimSpace(g)
		if goals[i] == "" {
			fmt.Fprintf(errStream, "Error: Goal names cannot be empty.\n")
			return 1
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := generateAmidakujiBoard(len(participants), r)
	results := evaluateAmidakuji(board, participants, goals)
	renderAmidakuji(outStream, board, participants, goals, results)

	return 0
}

func generateAmidakujiBoard(numParticipants int, r *rand.Rand) [][]bool {
	// Let's create a board with enough horizontal steps to make it random.
	// For N participants, maybe max(10, N*2)
	steps := numParticipants * 2
	if steps < 10 {
		steps = 10
	}

	board := make([][]bool, steps)
	for s := 0; s < steps; s++ {
		board[s] = make([]bool, numParticipants-1)
	}

	for s := 0; s < steps; s++ {
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
			if currentLine > 0 && row[currentLine-1] {
				currentLine--
			} else if currentLine < len(participants)-1 && row[currentLine] {
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

	for _, p := range participants {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", p))
	}
	fmt.Fprintf(outStream, "\n")

	paddingLeft := (colWidth / 2)
	paddingRight := colWidth - paddingLeft - 1

	for range participants {
		fmt.Fprintf(outStream, "%*s|%*s", paddingLeft, "", paddingRight, "")
	}
	fmt.Fprintf(outStream, "\n")

	for _, row := range board {
		fmt.Fprintf(outStream, "%*s", paddingLeft, "")

		for i := 0; i < len(participants); i++ {
			fmt.Fprintf(outStream, "|")

			if i < len(participants)-1 {
				connectionLen := paddingRight + paddingLeft
				if row[i] {
					fmt.Fprintf(outStream, "%s", strings.Repeat("-", connectionLen))
				} else {
					fmt.Fprintf(outStream, "%s", strings.Repeat(" ", connectionLen))
				}
			} else {
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

	for _, g := range goals {
		fmt.Fprintf(outStream, "%-*s", colWidth, fmt.Sprintf(" %s", g))
	}
	fmt.Fprintf(outStream, "\n\n")

	fmt.Fprintf(outStream, "Results:\n")
	for _, p := range participants {
		fmt.Fprintf(outStream, "  %s -> %s\n", p, results[p])
	}
}
