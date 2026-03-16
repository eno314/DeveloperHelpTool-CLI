package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-cmp/cmp"
)

func runHttpDiff(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet("httpdiff", flag.ContinueOnError)
	flags.SetOutput(errStream)

	var host1, host2, pathsStr string
	flags.StringVar(&host1, "host1", "", "First host URL (e.g., http://example.com)")
	flags.StringVar(&host2, "host2", "", "Second host URL (e.g., http://example.org)")
	flags.StringVar(&pathsStr, "paths", "", "Comma-separated list of paths (e.g., /api/v1/users,/api/v1/posts)")

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Usage of httpdiff:\n")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	if host1 == "" || host2 == "" || pathsStr == "" {
		fmt.Fprintln(errStream, "Error: --host1, --host2, and --paths are required")
		flags.Usage()
		return 1
	}

	paths := strings.Split(pathsStr, ",")
	client := &http.Client{}
	hasErrors := false

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		u1, err := url.JoinPath(host1, p)
		if err != nil {
			fmt.Fprintf(errStream, "Error joining path for host1 (%s) and path (%s): %v\n", host1, p, err)
			hasErrors = true
			continue
		}

		u2, err := url.JoinPath(host2, p)
		if err != nil {
			fmt.Fprintf(errStream, "Error joining path for host2 (%s) and path (%s): %v\n", host2, p, err)
			hasErrors = true
			continue
		}

		fmt.Fprintf(outStream, "Comparing %s vs %s\n", u1, u2)

		resp1, body1, err := doRequest(client, u1)
		if err != nil {
			fmt.Fprintf(errStream, "Error requesting %s: %v\n", u1, err)
			hasErrors = true
			continue
		}

		resp2, body2, err := doRequest(client, u2)
		if err != nil {
			fmt.Fprintf(errStream, "Error requesting %s: %v\n", u2, err)
			hasErrors = true
			continue
		}

		if resp1.StatusCode != resp2.StatusCode {
			fmt.Fprintf(outStream, "Warning: Status codes differ. host1: %d, host2: %d\n", resp1.StatusCode, resp2.StatusCode)
		}

		diff := cmp.Diff(body1, body2)
		if diff == "" {
			fmt.Fprintf(outStream, "No differences found.\n")
		} else {
			fmt.Fprintf(outStream, "Differences found:\n%s\n", diff)
		}
		fmt.Fprintln(outStream, "--------------------------------------------------")
	}

	if hasErrors {
		return 1
	}
	return 0
}

func doRequest(client *http.Client, url string) (*http.Response, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return resp, string(bodyBytes), nil
}
