package fmd2json

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantProps map[string]any
		wantBody  string
	}{
		{
			name:  "basic frontmatter",
			input: "---\nprop1: aaa\n---\nbody body\n",
			wantProps: map[string]any{
				"prop1": "aaa",
			},
			wantBody: "body body\n",
		},
		{
			name:      "no frontmatter",
			input:     "no frontmatter here\njust plain markdown\n",
			wantProps: nil,
			wantBody:  "no frontmatter here\njust plain markdown\n",
		},
		{
			name:      "empty content",
			input:     "",
			wantProps: nil,
			wantBody:  "",
		},
		{
			name:  "frontmatter only",
			input: "---\nkey: val\n---",
			wantProps: map[string]any{
				"key": "val",
			},
			wantBody: "",
		},
		{
			name:  "numeric value",
			input: "---\ncount: 42\n---\nbody\n",
			wantProps: map[string]any{
				"count": uint64(42),
			},
			wantBody: "body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, body := parseFrontmatter([]byte(tt.input))
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if tt.wantProps == nil {
				if props != nil {
					t.Errorf("props = %v, want nil", props)
				}
				return
			}
			for k, want := range tt.wantProps {
				got, ok := props[k]
				if !ok {
					t.Errorf("missing property %q", k)
					continue
				}
				if got != want {
					t.Errorf("props[%q] = %v (%T), want %v (%T)", k, got, got, want, want)
				}
			}
		})
	}
}

func TestWarnConflicts(t *testing.T) {
	var errBuf bytes.Buffer
	props := map[string]any{
		"filename": "custom",
		"body":     "custom body",
		"other":    "value",
	}
	warnConflicts(props, &errBuf)
	output := errBuf.String()
	if !strings.Contains(output, `"filename"`) {
		t.Error("expected warning for filename conflict")
	}
	if !strings.Contains(output, `"body"`) {
		t.Error("expected warning for body conflict")
	}
	if strings.Contains(output, `"other"`) {
		t.Error("unexpected warning for other property")
	}
}

func TestRunWithFiles(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(context.Background(), []string{"testdata/basic.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, outBuf.String())
	}

	if result["filename"] != "basic" {
		t.Errorf("filename = %v, want %q", result["filename"], "basic")
	}
	if result["prop1"] != "aaa" {
		t.Errorf("prop1 = %v, want %q", result["prop1"], "aaa")
	}
	if result["body"] != "body body\n" {
		t.Errorf("body = %q, want %q", result["body"], "body body\n")
	}
	if _, ok := result["mtime"]; !ok {
		t.Error("expected mtime property")
	}
}

func TestRunNoFrontmatter(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(context.Background(), []string{"testdata/no_frontmatter.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if result["filename"] != "no_frontmatter" {
		t.Errorf("filename = %v, want %q", result["filename"], "no_frontmatter")
	}
	if result["body"] != "no frontmatter here\njust plain markdown\n" {
		t.Errorf("body = %q", result["body"])
	}
}

func TestRunConflictWarning(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(context.Background(), []string{"testdata/conflict.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}

	// Check warnings on stderr
	warnings := errBuf.String()
	if !strings.Contains(warnings, `"filename"`) {
		t.Error("expected warning for filename conflict")
	}
	if !strings.Contains(warnings, `"body"`) {
		t.Error("expected warning for body conflict")
	}
	if !strings.Contains(warnings, `"mtime"`) {
		t.Error("expected warning for mtime conflict")
	}

	// Default properties should override
	var result map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["filename"] != "conflict" {
		t.Errorf("filename = %v, want %q", result["filename"], "conflict")
	}
	if result["body"] != "actual body\n" {
		t.Errorf("body = %q, want %q", result["body"], "actual body\n")
	}
	if result["other"] != "value" {
		t.Errorf("other = %v, want %q", result["other"], "value")
	}
}

func TestRunMultipleFiles(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(context.Background(), []string{"testdata/basic.md", "testdata/second.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (ndjson), got %d: %s", len(lines), outBuf.String())
	}

	var first, second map[string]any
	json.Unmarshal([]byte(lines[0]), &first)
	json.Unmarshal([]byte(lines[1]), &second)

	if first["filename"] != "basic" {
		t.Errorf("first filename = %v", first["filename"])
	}
	if second["filename"] != "second" {
		t.Errorf("second filename = %v", second["filename"])
	}
	if second["title"] != "second" {
		t.Errorf("second title = %v", second["title"])
	}
}

func TestRunVersion(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(context.Background(), []string{"--version"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(outBuf.String(), "fmd2json") {
		t.Errorf("version output = %q", outBuf.String())
	}
}
