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
	cfg, err := parseAndValidateHttpDiffArgs(args, errStream)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if compareAllPaths(cfg, outStream, errStream) {
		return 0
	}
	return 1
}

func parseAndValidateHttpDiffArgs(args []string, errStream io.Writer) (*httpDiffConfig, error) {
	flags, opts := newFlagSet(errStream)
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	cfg, err := newHttpDiffConfig(opts)
	if err != nil {
		fmt.Fprintf(errStream, "Error: %v\n", err)
		if errors.Is(err, errMissingFlags) {
			flags.Usage()
		}
		return nil, err
	}
	return cfg, nil
}

func compareAllPaths(cfg *httpDiffConfig, outStream, errStream io.Writer) bool {
	client := &http.Client{}
	success := true
	for _, p := range cfg.paths {
		if !comparePath(client, cfg, p, outStream, errStream) {
			success = false
		}
	}
	return success
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
	if err := validateRequiredOpts(opts); err != nil {
		return nil, err
	}
	if err := validateHeaders(opts.headers); err != nil {
		return nil, err
	}

	bodyBytes, err := resolveBodyBytes(opts)
	if err != nil {
		return nil, err
	}

	return &httpDiffConfig{
		host1:     opts.host1,
		host2:     opts.host2,
		paths:     parsePathsStr(opts.pathsStr),
		method:    strings.ToUpper(opts.method),
		bodyBytes: bodyBytes,
		headers:   opts.headers,
	}, nil
}

func validateRequiredOpts(opts *httpDiffOptions) error {
	if opts.host1 == "" || opts.host2 == "" || opts.pathsStr == "" {
		return errMissingFlags
	}
	if opts.body != "" && opts.bodyFile != "" {
		return fmt.Errorf("cannot specify both --body and --body-file")
	}
	return nil
}

func validateHeaders(headers []string) error {
	for _, h := range headers {
		if !strings.Contains(h, ":") {
			return fmt.Errorf("invalid header format: %s", h)
		}
	}
	return nil
}

func resolveBodyBytes(opts *httpDiffOptions) ([]byte, error) {
	if opts.bodyFile != "" {
		b, err := os.ReadFile(opts.bodyFile)
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		return b, nil
	}
	return []byte(opts.body), nil
}

func parsePathsStr(pathsStr string) []string {
	var paths []string
	for _, p := range strings.Split(pathsStr, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

func comparePath(client *http.Client, cfg *httpDiffConfig, path string, outStream, errStream io.Writer) bool {
	u1, u2, err := buildCompareURLs(cfg.host1, cfg.host2, path)
	if err != nil {
		fmt.Fprintf(errStream, "Error: %v\n", err)
		return false
	}

	fmt.Fprintf(outStream, "Comparing %s vs %s\n", u1, u2)

	resp1, body1, err1 := doRequest(client, u1, cfg.method, cfg.bodyBytes, cfg.headers)
	resp2, body2, err2 := doRequest(client, u2, cfg.method, cfg.bodyBytes, cfg.headers)
	if err := checkRequestErrors(u1, u2, err1, err2, errStream); err != nil {
		return false
	}

	warnStatusDifferences(outStream, resp1.StatusCode, resp2.StatusCode)
	printBodyDiff(outStream, body1, body2)
	return true
}

func buildCompareURLs(host1, host2, path string) (string, string, error) {
	u1, err := url.JoinPath(host1, path)
	if err != nil {
		return "", "", fmt.Errorf("joining path for host1 (%s) and path (%s): %w", host1, path, err)
	}
	u2, err := url.JoinPath(host2, path)
	if err != nil {
		return "", "", fmt.Errorf("joining path for host2 (%s) and path (%s): %w", host2, path, err)
	}
	return u1, u2, nil
}

func checkRequestErrors(u1, u2 string, err1, err2 error, errStream io.Writer) error {
	if err1 != nil {
		fmt.Fprintf(errStream, "Error requesting %s: %v\n", u1, err1)
		return err1
	}
	if err2 != nil {
		fmt.Fprintf(errStream, "Error requesting %s: %v\n", u2, err2)
		return err2
	}
	return nil
}

func warnStatusDifferences(outStream io.Writer, status1, status2 int) {
	if status1 != status2 {
		fmt.Fprintf(outStream, "Warning: Status codes differ. host1: %d, host2: %d\n", status1, status2)
	}
}

func printBodyDiff(outStream io.Writer, body1, body2 string) {
	diff := cmp.Diff(body1, body2)
	if diff == "" {
		fmt.Fprintf(outStream, "No differences found.\n")
	} else {
		fmt.Fprintf(outStream, "Differences found:\n%s\n", diff)
	}
	fmt.Fprintln(outStream, "--------------------------------------------------")
}

func doRequest(client *http.Client, urlStr string, method string, body []byte, headers []string) (*http.Response, string, error) {
	req, err := createRequest(method, urlStr, body)
	if err != nil {
		return nil, "", err
	}

	applyHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("executing request to %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	bodyStr, err := readResponseBody(resp.Body, urlStr)
	if err != nil {
		return nil, "", err
	}

	return resp, bodyStr, nil
}

func createRequest(method, urlStr string, body []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", urlStr, err)
	}
	return req, nil
}

func applyHeaders(req *http.Request, headers []string) {
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

func readResponseBody(body io.ReadCloser, urlStr string) (string, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("reading response body from %s: %w", urlStr, err)
	}
	return string(bodyBytes), nil
}

type headersValue []string

func (h *headersValue) String() string {
	return strings.Join(*h, ", ")
}

func (h *headersValue) Set(value string) error {
	*h = append(*h, value)
	return nil
}
