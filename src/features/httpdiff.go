package features

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
)

func RunHttpDiff(args []string, outStream, errStream io.Writer) int {
	flags := flag.NewFlagSet("httpdiff", flag.ContinueOnError)
	flags.SetOutput(errStream)

	var host1, host2, pathsStr string
	var method, body, bodyFile string
	var headers headersValue

	flags.StringVar(&host1, "host1", "", "First host URL (e.g., http://example.com)")
	flags.StringVar(&host2, "host2", "", "Second host URL (e.g., http://example.org)")
	flags.StringVar(&pathsStr, "paths", "", "Comma-separated list of paths (e.g., /api/v1/users,/api/v1/posts)")
	flags.StringVar(&method, "method", "GET", "HTTP method (e.g., GET, POST)")
	flags.StringVar(&body, "body", "", "HTTP request body string")
	flags.StringVar(&bodyFile, "body-file", "", "Path to file containing HTTP request body")
	flags.Var(&headers, "header", "Custom HTTP request header (can be specified multiple times, e.g. -header 'Name: Value')")

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

	if body != "" && bodyFile != "" {
		fmt.Fprintln(errStream, "Error: cannot specify both --body and --body-file")
		return 1
	}

	for _, h := range headers {
		if !strings.Contains(h, ":") {
			fmt.Fprintf(errStream, "Error: invalid header format: %s\n", h)
			return 1
		}
	}

	var bodyBytes []byte
	if bodyFile != "" {
		var err error
		bodyBytes, err = os.ReadFile(bodyFile)
		if err != nil {
			fmt.Fprintf(errStream, "Error reading body file: %v\n", err)
			return 1
		}
	} else if body != "" {
		bodyBytes = []byte(body)
	}

	method = strings.ToUpper(method)

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

		resp1, body1, err := doRequest(client, u1, method, bodyBytes, headers)
		if err != nil {
			fmt.Fprintf(errStream, "Error requesting %s: %v\n", u1, err)
			hasErrors = true
			continue
		}

		resp2, body2, err := doRequest(client, u2, method, bodyBytes, headers)
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

func doRequest(client *http.Client, urlStr string, method string, body []byte, headers []string) (*http.Response, string, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, "", fmt.Errorf("creating request for %s: %w", urlStr, err)
	}

	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("executing request to %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response body from %s: %w", urlStr, err)
	}

	return resp, string(bodyBytes), nil
}

type headersValue []string

func (h *headersValue) String() string {
	return strings.Join(*h, ", ")
}

func (h *headersValue) Set(value string) error {
	*h = append(*h, value)
	return nil
}
