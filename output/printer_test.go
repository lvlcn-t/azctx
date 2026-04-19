package output

import (
	"bytes"
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Format
		wantErr string
	}{
		{name: "text", raw: "text", want: FormatText},
		{name: "table", raw: "table", want: FormatTable},
		{name: "json", raw: "json", want: FormatJSON},
		{name: "trimmed and case insensitive", raw: "  JSON  ", want: FormatJSON},
		{name: "unsupported", raw: "yaml", wantErr: "unsupported output format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.raw)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{name: "map", input: map[string]any{"b": 2, "a": 1}},
		{name: "slice", input: []string{"one", "two"}},
		{name: "struct", input: struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{Name: "azctx", Age: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			err := PrintJSON(&output, tt.input)
			require.NoError(t, err)

			var decoded any
			require.NoError(t, json.Unmarshal(output.Bytes(), &decoded))
		})
	}
}

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		wantErr string
		wants   []string
	}{
		{
			name:    "empty headers",
			headers: nil,
			rows:    [][]string{{"a"}},
			wantErr: "table headers cannot be empty",
		},
		{
			name:    "writes headers and rows",
			headers: []string{"A", "B"},
			rows:    [][]string{{"1", "2"}, {"3", "4"}},
			wants:   []string{"A", "B", "1", "2", "3", "4"},
		},
		{
			name:    "writes only headers when no rows",
			headers: []string{"A", "B"},
			rows:    nil,
			wants:   []string{"A", "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			err := PrintTable(&output, tt.headers, tt.rows)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			got := output.String()
			for _, want := range tt.wants {
				assert.Contains(t, got, want)
			}
		})
	}
}
