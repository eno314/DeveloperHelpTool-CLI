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
	switch command {
	case "amidakuji":
		return runAmidakuji(args[2:], outStream, errStream)
	case "httpdiff":
		return runHttpDiff(args[2:], outStream, errStream)
	case "-h", "--help":
		flags.Usage()
		return 0
	default:
		fmt.Fprintf(errStream, "Unknown command: %s\n", command)
		flags.Usage()
		return 1
	}
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
