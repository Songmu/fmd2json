package fmd2json

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/itchyny/gojq"
)

// applyJQ applies a jq expression to the input value and writes results to w.
// When rawOutput is true, scalar values are written as raw text (like jq --raw-output).
// When rawOutput is false, all values are written as JSON (like jq default).
func applyJQ(input any, expr string, w io.Writer, rawOutput bool) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("failed to parse jq expression: %w", err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		return fmt.Errorf("failed to compile jq expression: %w", err)
	}

	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return err
		}
		if rawOutput {
			if text, e := scalarToString(v); e == nil {
				if _, err := fmt.Fprintln(w, text); err != nil {
					return err
				}
				continue
			}
		}
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

func scalarToString(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		if math.Trunc(val) == val {
			return strconv.FormatFloat(val, 'f', 0, 64), nil
		}
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(val), nil
	case bool:
		return strconv.FormatBool(val), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("not a scalar: %T", v)
	}
}
