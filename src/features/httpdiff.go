package features

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
)

var errMissingFlags = errors.New("--host1, --host2, and --paths are required")

type httpDiffConfig struct {
	host1     string
	host2     string
	paths     []string
	method    string
	bodyBytes []byte
	headers   headersValue
}

type httpDiffOptions struct {
	host1    string
	host2    string
	pathsStr string
	method   string
	body     string
	bodyFile string
	headers  headersValue
}

// RunHttpDiff is the entry point for the httpdiff command.
func RunHttpDiff(args []string, outStream, errStream io.Writer) int {
	flags, opts := newFlagSet(errStream)

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	cfg, err := newHttpDiffConfig(opts)
	if err != nil {
		fmt.Fprintf(errStream, "Error: %v\n", err)
		if err == errMissingFlags {
			flags.Usage()
		}
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

func newFlagSet(errStream io.Writer) (*flag.FlagSet, *httpDiffOptions) {
	opts := &httpDiffOptions{}
	flags := flag.NewFlagSet("httpdiff", flag.ContinueOnError)
	flags.SetOutput(errStream)

	flags.StringVar(&opts.host1, "host1", "", "First host URL (e.g., http://example.com)")
	flags.StringVar(&opts.host2, "host2", "", "Second host URL (e.g., http://example.org)")
	flags.StringVar(&opts.pathsStr, "paths", "", "Comma-separated list of paths (e.g., /api/v1/users,/api/v1/posts)")
	flags.StringVar(&opts.method, "method", "GET", "HTTP method (e.g., GET, POST)")
	flags.StringVar(&opts.body, "body", "", "HTTP request body string")
	flags.StringVar(&opts.bodyFile, "body-file", "", "Path to file containing HTTP request body")
	flags.Var(&opts.headers, "header", "Custom HTTP request header (can be specified multiple times, e.g. -header 'Name: Value')")

	flags.Usage = func() {
		fmt.Fprintf(errStream, "Usage of httpdiff:\n")
		flags.PrintDefaults()
	}

	return flags, opts
}

func newHttpDiffConfig(opts *httpDiffOptions) (*httpDiffConfig, error) {
	if opts.host1 == "" || opts.host2 == "" || opts.pathsStr == "" {
		return nil, errMissingFlags
	}

	if opts.body != "" && opts.bodyFile != "" {
		return nil, fmt.Errorf("cannot specify both --body and --body-file")
	}

	for _, h := range opts.headers {
		if !strings.Contains(h, ":") {
			return nil, fmt.Errorf("invalid header format: %s", h)
		}
	}

	var bodyBytes []byte
	if opts.bodyFile != "" {
		b, err := os.ReadFile(opts.bodyFile)
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		bodyBytes = b
	} else if opts.body != "" {
		bodyBytes = []byte(opts.body)
	}

	var paths []string
	for _, p := range strings.Split(opts.pathsStr, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}

	return &httpDiffConfig{
		host1:     opts.host1,
		host2:     opts.host2,
		paths:     paths,
		method:    strings.ToUpper(opts.method),
		bodyBytes: bodyBytes,
		headers:   opts.headers,
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
