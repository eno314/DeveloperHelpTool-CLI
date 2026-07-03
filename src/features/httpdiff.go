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

type httpDiffConfig struct {
	host1     string
	host2     string
	paths     []string
	method    string
	bodyBytes []byte
	headers   headersValue
}

// RunHttpDiff is the entry point for the httpdiff command.
func RunHttpDiff(args []string, outStream, errStream io.Writer) int {
	var host1, host2, pathsStr string
	var method, body, bodyFile string
	var headers headersValue

	flags := flag.NewFlagSet("httpdiff", flag.ContinueOnError)
	flags.SetOutput(errStream)

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

	cfg, err := newHttpDiffConfig(host1, host2, pathsStr, method, body, bodyFile, headers)
	if err != nil {
		fmt.Fprintf(errStream, "Error: %v\n", err)
		return 1
	}

	client := &http.Client{}
	hasErrors := false

	for _, p := range cfg.paths {
		if !comparePath(client, cfg.host1, cfg.host2, p, cfg.method, cfg.bodyBytes, cfg.headers, outStream, errStream) {
			hasErrors = true
		}
	}

	if hasErrors {
		return 1
	}
	return 0
}

func newHttpDiffConfig(host1, host2, pathsStr, method, body, bodyFile string, headers headersValue) (*httpDiffConfig, error) {
	if body != "" && bodyFile != "" {
		return nil, fmt.Errorf("cannot specify both --body and --body-file")
	}

	for _, h := range headers {
		if !strings.Contains(h, ":") {
			return nil, fmt.Errorf("invalid header format: %s", h)
		}
	}

	var bodyBytes []byte
	if bodyFile != "" {
		b, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		bodyBytes = b
	} else if body != "" {
		bodyBytes = []byte(body)
	}

	var paths []string
	for _, p := range strings.Split(pathsStr, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}

	return &httpDiffConfig{
		host1:     host1,
		host2:     host2,
		paths:     paths,
		method:    strings.ToUpper(method),
		bodyBytes: bodyBytes,
		headers:   headers,
	}, nil
}

func comparePath(client *http.Client, host1, host2, path, method string, bodyBytes []byte, headers []string, outStream, errStream io.Writer) bool {
	u1, err := url.JoinPath(host1, path)
	if err != nil {
		fmt.Fprintf(errStream, "Error joining path for host1 (%s) and path (%s): %v\n", host1, path, err)
		return false
	}

	u2, err := url.JoinPath(host2, path)
	if err != nil {
		fmt.Fprintf(errStream, "Error joining path for host2 (%s) and path (%s): %v\n", host2, path, err)
		return false
	}

	fmt.Fprintf(outStream, "Comparing %s vs %s\n", u1, u2)

	resp1, body1, err := doRequest(client, u1, method, bodyBytes, headers)
	if err != nil {
		fmt.Fprintf(errStream, "Error requesting %s: %v\n", u1, err)
		return false
	}

	resp2, body2, err := doRequest(client, u2, method, bodyBytes, headers)
	if err != nil {
		fmt.Fprintf(errStream, "Error requesting %s: %v\n", u2, err)
		return false
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

	return true
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
