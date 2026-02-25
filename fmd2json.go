package fmd2json

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const cmdName = "fmd2json"

var defaultPropertyNames = []string{"filename", "body", "mtime"}

// Run the fmd2json
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}

	args := fs.Args()
	switch {
	case len(args) > 0:
		for _, arg := range args {
			if err := processArg(arg, outStream, errStream); err != nil {
				return err
			}
		}
	default:
		// Read file list from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if err := processFile(line, outStream, errStream); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
	}
	return nil
}

func processArg(arg string, outStream, errStream io.Writer) error {
	if arg == "-" {
		return processStdin(outStream, errStream)
	}
	return processFile(arg, outStream, errStream)
}

func processStdin(outStream, errStream io.Writer) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	props, body := parseFrontmatter(data)
	warnConflicts(props, errStream)

	result := buildResult(props, "", body, nil)
	return writeJSON(outStream, result)
}

func processFile(path string, outStream, errStream io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file %s: %w", path, err)
	}
	props, body := parseFrontmatter(data)
	warnConflicts(props, errStream)

	filename := strings.TrimSuffix(filepath.Base(path), ".md")
	mtime := fi.ModTime().Format(time.RFC3339)

	result := buildResult(props, filename, body, &mtime)
	return writeJSON(outStream, result)
}

// parseFrontmatter splits markdown content into frontmatter properties and body.
// The frontmatter must be delimited by "---" lines at the beginning of the content.
func parseFrontmatter(data []byte) (map[string]any, string) {
	content := string(data)

	// Check for frontmatter delimiter at the start
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, content
	}

	// Find the closing delimiter
	rest := content[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---\n")
	idxR := strings.Index(rest, "\r\n---\r\n")
	if idx < 0 && idxR < 0 {
		// Check if closing delimiter is at the very end
		if strings.HasSuffix(rest, "\n---") {
			idx = len(rest) - 4
		} else if strings.HasSuffix(rest, "\r\n---") {
			idx = len(rest) - 5
		}
		if idx < 0 {
			return nil, content
		}
		yamlContent := rest[:idx]
		props := parseYAML(yamlContent)
		return props, ""
	}

	var yamlContent string
	var body string
	if idx >= 0 && (idxR < 0 || idx <= idxR) {
		yamlContent = rest[:idx]
		body = rest[idx+5:] // skip "\n---\n"
	} else {
		yamlContent = rest[:idxR]
		body = rest[idxR+7:] // skip "\r\n---\r\n"
	}

	props := parseYAML(yamlContent)
	return props, body
}

func parseYAML(content string) map[string]any {
	var props map[string]any
	if err := yaml.NewDecoder(
		bytes.NewReader([]byte(content)),
		yaml.UseOrderedMap(),
	).Decode(&props); err != nil {
		return nil
	}
	return props
}

func warnConflicts(props map[string]any, errStream io.Writer) {
	if props == nil {
		return
	}
	for _, name := range defaultPropertyNames {
		if _, ok := props[name]; ok {
			fmt.Fprintf(errStream, "warning: frontmatter property %q conflicts with default property, using default value\n", name)
		}
	}
}

func buildResult(props map[string]any, filename, body string, mtime *string) map[string]any {
	result := make(map[string]any)
	// Copy frontmatter properties first
	for k, v := range props {
		result[k] = v
	}
	// Override with default properties
	result["filename"] = filename
	result["body"] = body
	if mtime != nil {
		result["mtime"] = *mtime
	}
	return result
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}
