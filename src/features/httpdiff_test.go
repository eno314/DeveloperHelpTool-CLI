package features

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunHttpDiff(t *testing.T) {
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/same", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello, World!")
	})
	mux1.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello, World 1!")
	})
	mux1.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	mux1.HandleFunc("/error1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error 1")
	})
	mux1.HandleFunc("/post-test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Test-Header") != "foo" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if buf.String() != "hello world" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success body")
	})
	mux1.HandleFunc("/post-file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Test-Header") != "bar" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if buf.String() != "hello file" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success file")
	})
	mux1.HandleFunc("/post-diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "diff1")
	})

	server1 := httptest.NewServer(mux1)
	defer server1.Close()

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/same", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello, World!")
	})
	mux2.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello, World 2!")
	})
	mux2.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	})
	mux2.HandleFunc("/error1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error 1") // Same error body to test we still diff it
	})
	mux2.HandleFunc("/post-test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Test-Header") != "foo" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if buf.String() != "hello world" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success body")
	})
	mux2.HandleFunc("/post-file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Test-Header") != "bar" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if buf.String() != "hello file" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success file")
	})
	mux2.HandleFunc("/post-diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "diff2")
	})
	server2 := httptest.NewServer(mux2)
	defer server2.Close()

	// Create a temp file for body-file testing
	err := os.WriteFile("test_body.txt", []byte("hello file"), 0644)
	if err != nil {
		t.Fatalf("failed to create test body file: %v", err)
	}
	defer os.Remove("test_body.txt")

	tests := []struct {
		name         string
		args         []string
		expectedCode int
		expectedOut  string
		expectedErr  string
	}{
		{
			name:         "Help flag",
			args:         []string{"httpdiff", "--help"},
			expectedCode: 0,
			expectedOut:  "",
			expectedErr:  "Usage of httpdiff:\n", // Will check for presence
		},
		{
			name:         "Missing flags",
			args:         []string{"httpdiff", "--host1", server1.URL},
			expectedCode: 1,
			expectedOut:  "",
			expectedErr:  "Error: --host1, --host2, and --paths are required\n", // Prefix check
		},
		{
			name:         "Same response",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/same"},
			expectedCode: 0,
			expectedOut:  "Comparing " + server1.URL + "/same vs " + server2.URL + "/same\nNo differences found.\n--------------------------------------------------\n",
			expectedErr:  "",
		},
		{
			name:         "Different response",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/diff"},
			expectedCode: 0,
			expectedOut:  "Differences found:\n", // Just check prefix/presence
			expectedErr:  "",
		},
		{
			name:         "Different status",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/status"},
			expectedCode: 0,
			expectedOut:  "Warning: Status codes differ. host1: 200, host2: 404\n",
			expectedErr:  "",
		},
		{
			name:         "Multiple paths",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/same,/diff"},
			expectedCode: 0,
			expectedOut:  "No differences found", // Check it contains this
			expectedErr:  "",
		},
		{
			name:         "Invalid URL error",
			args:         []string{"httpdiff", "--host1", "http://invalid-host-that-does-not-exist", "--host2", server2.URL, "--paths", "/same"},
			expectedCode: 1, // Has errors
			expectedOut:  "Comparing",
			expectedErr:  "Error requesting",
		},
		{
			name:         "Conflict validation: body and body-file",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/same", "--body", "hello", "--body-file", "test_body.txt"},
			expectedCode: 1,
			expectedOut:  "",
			expectedErr:  "Error: cannot specify both --body and --body-file",
		},
		{
			name:         "Header format validation: no colon",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/same", "--header", "InvalidHeader"},
			expectedCode: 1,
			expectedOut:  "",
			expectedErr:  "Error: invalid header format: InvalidHeader",
		},
		{
			name:         "POST request with custom header and inline body",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/post-test", "--method", "POST", "--header", "X-Test-Header: foo", "--body", "hello world"},
			expectedCode: 0,
			expectedOut:  "No differences found.",
			expectedErr:  "",
		},
		{
			name:         "POST request with custom header and body file",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/post-file", "--method", "post", "--header", "X-Test-Header: bar", "--body-file", "test_body.txt"},
			expectedCode: 0,
			expectedOut:  "No differences found.",
			expectedErr:  "",
		},
		{
			name:         "POST request with different responses",
			args:         []string{"httpdiff", "--host1", server1.URL, "--host2", server2.URL, "--paths", "/post-diff", "--method", "POST"},
			expectedCode: 0,
			expectedOut:  "Differences found:\n",
			expectedErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			code := RunHttpDiff(tt.args[1:], outBuf, errBuf) // drop the "httpdiff" command name as `args` usually does in `main` after switch

			if code != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, code)
			}

			outStr := outBuf.String()
			errStr := errBuf.String()

			if tt.expectedOut != "" && !strings.Contains(outStr, tt.expectedOut) {
				t.Errorf("stdout expected to contain %q, but got %q", tt.expectedOut, outStr)
			}

			if tt.expectedErr != "" && !strings.Contains(errStr, tt.expectedErr) {
				t.Errorf("stderr expected to contain %q, but got %q", tt.expectedErr, errStr)
			}
		})
	}
}
