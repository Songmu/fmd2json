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

	"github.com/Songmu/skillsmith"
	"github.com/goccy/go-yaml"
)

const cmdName = "fmd2json"

var defaultPropertyNames = []string{"dir", "filename", "body", "mtime"}

// Run the fmd2json
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	if len(argv) > 0 && argv[0] == "skills" {
		s, err := skillsmith.New("fmd2json", version, skillsFS)
		if err != nil {
			return err
		}
		s.OutWriter = outStream
		s.ErrWriter = errStream
		return s.Run(ctx, argv[1:])
	}
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.Usage = func() {
		fmt.Fprintf(errStream, "Usage: %s [options] [file...]\n\nSubcommands:\n  skills  Manage agent skills\n\nOptions:\n", cmdName)
		fs.PrintDefaults()
	}
	ver := fs.Bool("version", false, "display version")
	filenameFlag := fs.String("filename", "", "specify filename for stdin input (used with -)")
	jqExpr := fs.String("jq", "", "jq expression to apply to each JSON output")
	rawOutput := fs.Bool("raw-output", false, "output raw strings instead of JSON encoded strings (with --jq)")
	fs.BoolVar(rawOutput, "r", false, "shorthand for --raw-output")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}
	if *rawOutput && *jqExpr == "" {
		return fmt.Errorf("--raw-output (-r) requires --jq to be specified")
	}

	outputFunc := func(w io.Writer, v any) error {
		if *jqExpr != "" {
			return applyJQ(v, *jqExpr, w, *rawOutput)
		}
		return writeJSON(w, v)
	}

	args := fs.Args()
	switch {
	case len(args) > 0:
		hasStdin := false
		for _, arg := range args {
			if arg == "-" {
				hasStdin = true
				break
			}
		}
		if *filenameFlag != "" && !hasStdin {
			log.Println("warning: -filename is only used with stdin input (-), ignoring")
		}
		for _, arg := range args {
			if err := processArg(arg, *filenameFlag, outStream, errStream, outputFunc); err != nil {
				return err
			}
		}
	default:
		if *filenameFlag != "" {
			log.Println("warning: -filename is only used with stdin input (-), ignoring")
		}
		// Read file list from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if err := processFile(line, outStream, errStream, outputFunc); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
	}
	return nil
}

func processArg(arg, nameOverride string, outStream, errStream io.Writer, outputFunc func(io.Writer, any) error) error {
	if arg == "-" {
		return processStdin(nameOverride, outStream, errStream, outputFunc)
	}
	return processFile(arg, outStream, errStream, outputFunc)
}

func processStdin(nameOverride string, outStream, errStream io.Writer, outputFunc func(io.Writer, any) error) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	props, body := parseFrontmatter(data)
	warnConflicts(props, errStream)

	var dir, filename string
	if nameOverride != "" {
		cleaned := filepath.Clean(nameOverride)
		filename = strings.TrimSuffix(filepath.Base(cleaned), ".md")
		dir = extractDir(cleaned)
	}
	result := buildResult(props, dir, filename, body, nil)
	return outputFunc(outStream, result)
}

func processFile(path string, outStream, errStream io.Writer, outputFunc func(io.Writer, any) error) error {
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

	cleaned := filepath.Clean(path)
	filename := strings.TrimSuffix(filepath.Base(cleaned), ".md")
	dir := extractDir(cleaned)
	mtime := fi.ModTime().Format(time.RFC3339)

	result := buildResult(props, dir, filename, body, &mtime)
	return outputFunc(outStream, result)
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

func buildResult(props map[string]any, dir, filename, body string, mtime *string) map[string]any {
	result := make(map[string]any)
	// Copy frontmatter properties first
	for k, v := range props {
		result[k] = v
	}
	// Override with default properties
	delete(result, "dir")
	if dir != "" {
		result["dir"] = dir
	}
	result["filename"] = filename
	result["body"] = body
	if mtime != nil {
		result["mtime"] = *mtime
	}
	return result
}

// extractDir returns the directory part of a cleaned path using forward slashes.
// Returns empty string if the directory is "." (current directory).
func extractDir(cleaned string) string {
	dir := filepath.Dir(cleaned)
	if dir == "." {
		return ""
	}
	return filepath.ToSlash(dir)
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
