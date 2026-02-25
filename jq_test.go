package fmd2json

import (
	"bytes"
	"strings"
	"testing"
)

func TestApplyJQ(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expr      string
		rawOutput bool
		want      string
		wantErr   bool
	}{
		{
			name:  "select string property (default JSON)",
			input: map[string]any{"name": "hello", "count": 42},
			expr:  ".name",
			want:  "\"hello\"\n",
		},
		{
			name:      "select string property (raw)",
			input:     map[string]any{"name": "hello", "count": 42},
			expr:      ".name",
			rawOutput: true,
			want:      "hello\n",
		},
		{
			name:  "select numeric property",
			input: map[string]any{"name": "hello", "count": 42},
			expr:  ".count",
			want:  "42\n",
		},
		{
			name:  "construct object",
			input: map[string]any{"name": "hello", "count": 42},
			expr:  `{n: .name}`,
			want:  "{\"n\":\"hello\"}\n",
		},
		{
			name:  "multiple outputs",
			input: map[string]any{"a": 1, "b": 2},
			expr:  ".a, .b",
			want:  "1\n2\n",
		},
		{
			name:  "array construction",
			input: map[string]any{"x": "hello"},
			expr:  "[.x, .x]",
			want:  "[\"hello\",\"hello\"]\n",
		},
		{
			name:  "pipe and length",
			input: map[string]any{"body": "hello world"},
			expr:  ".body | length",
			want:  "11\n",
		},
		{
			name:  "boolean result (default JSON)",
			input: map[string]any{"a": 1},
			expr:  ".a == 1",
			want:  "true\n",
		},
		{
			name:  "null result (default JSON)",
			input: map[string]any{"a": 1},
			expr:  ".missing",
			want:  "null\n",
		},
		{
			name:      "null result (raw)",
			input:     map[string]any{"a": 1},
			expr:      ".missing",
			rawOutput: true,
			want:      "\n",
		},
		{
			name:  "multiline string (default JSON)",
			input: map[string]any{"body": "line1\nline2\n"},
			expr:  ".body",
			want:  "\"line1\\nline2\\n\"\n",
		},
		{
			name:      "multiline string (raw)",
			input:     map[string]any{"body": "line1\nline2\n"},
			expr:      ".body",
			rawOutput: true,
			want:      "line1\nline2\n\n",
		},
		{
			name:    "invalid expression",
			input:   map[string]any{},
			expr:    ".foo[",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := applyJQ(tt.input, tt.expr, &buf, tt.rawOutput)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("got %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestScalarToString(t *testing.T) {
	tests := []struct {
		input   any
		want    string
		wantErr bool
	}{
		{input: "hello", want: "hello"},
		{input: 42.0, want: "42"},
		{input: 3.14, want: "3.14"},
		{input: true, want: "true"},
		{input: false, want: "false"},
		{input: nil, want: ""},
		{input: 123, want: "123"},
		{input: []any{1, 2}, wantErr: true},
		{input: map[string]any{"a": 1}, wantErr: true},
	}

	for _, tt := range tests {
		got, err := scalarToString(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("scalarToString(%v): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("scalarToString(%v): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("scalarToString(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRunWithJQ(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(t.Context(), []string{"--jq", ".filename", "testdata/basic.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(outBuf.String())
	if got != `"basic"` {
		t.Errorf("got %q, want %q", got, `"basic"`)
	}
}

func TestRunWithJQRawOutput(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(t.Context(), []string{"--jq", ".filename", "-r", "testdata/basic.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(outBuf.String())
	if got != "basic" {
		t.Errorf("got %q, want %q", got, "basic")
	}
}

func TestRunWithJQMultipleFiles(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(t.Context(), []string{"--jq", ".filename", "-r", "testdata/basic.md", "testdata/second.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), outBuf.String())
	}
	if lines[0] != "basic" {
		t.Errorf("line 0 = %q, want %q", lines[0], "basic")
	}
	if lines[1] != "second" {
		t.Errorf("line 1 = %q, want %q", lines[1], "second")
	}
}

func TestRunWithJQObject(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(t.Context(), []string{"--jq", `{name: .filename, prop: .prop1}`, "testdata/basic.md"}, &outBuf, &errBuf)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(outBuf.String())
	want := `{"name":"basic","prop":"aaa"}`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunWithJQInvalidExpr(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := Run(t.Context(), []string{"--jq", ".foo[", "testdata/basic.md"}, &outBuf, &errBuf)
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}
