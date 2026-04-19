package output

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const (
	tableMinWidth = 0
	tableTabWidth = 8
	tablePadding  = 2
	tablePadChar  = ' '
	tableFlags    = 0
)

type Format string

const (
	// FormatText renders plain text output.
	FormatText Format = "text"
	// FormatTable renders tabular output.
	FormatTable Format = "table"
	// FormatJSON renders JSON output.
	FormatJSON Format = "json"
)

// ParseFormat parses and validates output format input.
func ParseFormat(raw string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(raw))) {
	case FormatText:
		return FormatText, nil
	case FormatTable:
		return FormatTable, nil
	case FormatJSON:
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported output format %q", raw)
	}
}

// PrintJSON writes pretty, deterministic JSON output.
func PrintJSON(writer io.Writer, value any) error {
	encoded, err := json.Marshal(
		value,
		json.Deterministic(true),
		jsontext.Multiline(true),
		jsontext.WithIndent("  "),
		jsontext.SpaceAfterColon(true),
	)
	if err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	if _, err := fmt.Fprintln(writer, string(encoded)); err != nil {
		return fmt.Errorf("write json output: %w", err)
	}

	return nil
}

// PrintTable writes a header and row matrix as aligned tabular output.
func PrintTable(writer io.Writer, headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return errors.New("table headers cannot be empty")
	}

	tableWriter := tabwriter.NewWriter(
		writer,
		tableMinWidth,
		tableTabWidth,
		tablePadding,
		tablePadChar,
		tableFlags,
	)

	if _, err := fmt.Fprintln(tableWriter, strings.Join(headers, "\t")); err != nil {
		return fmt.Errorf("write table headers: %w", err)
	}

	for _, row := range rows {
		if _, err := fmt.Fprintln(tableWriter, strings.Join(row, "\t")); err != nil {
			return fmt.Errorf("write table row: %w", err)
		}
	}

	if err := tableWriter.Flush(); err != nil {
		return fmt.Errorf("flush table output: %w", err)
	}

	return nil
}
