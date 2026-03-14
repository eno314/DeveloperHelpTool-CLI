package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func run(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(errStream)

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Developer Help Tool CLI\n\n")
		fmt.Fprintf(errStream, "Usage:\n")
		fmt.Fprintf(errStream, "  %s [command]\n\n", args[0])
		fmt.Fprintf(errStream, "Available Commands:\n")
		fmt.Fprintf(errStream, "  (Currently no commands are available)\n\n")
		fmt.Fprintf(errStream, "Flags:\n")
		flags.PrintDefaults()
	}

	err := flags.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	if len(flags.Args()) == 0 {
		flags.Usage()
		return 0
	}

	command := flags.Arg(0)
	if command != "" {
		fmt.Fprintf(errStream, "Unknown command: %s\n", command)
		flags.Usage()
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
