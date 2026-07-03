package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/eno314/developer-help-tool-cli/src/features"
)

func run(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(errStream)

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Developer Help Tool CLI\n\n")
		fmt.Fprintf(errStream, "Usage:\n")
		fmt.Fprintf(errStream, "  %s [command]\n\n", args[0])
		fmt.Fprintf(errStream, "Available Commands:\n")
		fmt.Fprintf(errStream, "  amidakuji   Generate an Amidakuji (Ghost Leg) lottery\n")
		fmt.Fprintf(errStream, "  httpdiff    Compare HTTP responses from two hosts\n\n")
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
	if command == "-h" || command == "--help" {
		flags.Usage()
		return 0
	}

	type commandFunc func(args []string, outStream, errStream io.Writer) int
	commands := map[string]commandFunc{
		"amidakuji": features.RunAmidakuji,
		"httpdiff":  features.RunHttpDiff,
	}

	fn, ok := commands[command]
	if !ok {
		fmt.Fprintf(errStream, "Unknown command: %s\n", command)
		flags.Usage()
		return 1
	}

	return fn(args[2:], outStream, errStream)
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
