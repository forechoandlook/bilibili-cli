package formatter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{65, "01:05"},
		{3665, "1:01:05"},
		{0, "00:00"},
		{-10, "00:00"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("FormatDuration(%d) = %s; expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{100, "100"},
		{10000, "1.0万"},
		{15500, "1.6万"},
	}

	for _, tt := range tests {
		result := FormatCount(tt.input)
		if result != tt.expected {
			t.Errorf("FormatCount(%d) = %s; expected %s", tt.input, result, tt.expected)
		}
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestEmitStructuredJSON(t *testing.T) {
	data := map[string]string{"foo": "bar"}

	output := captureStdout(func() {
		EmitStructured(data, OutputFormatJSON)
	})

	if !strings.Contains(output, `"ok": true`) {
		t.Errorf("Expected 'ok: true' in JSON output")
	}
	if !strings.Contains(output, `"foo": "bar"`) {
		t.Errorf("Expected 'foo: bar' in JSON output")
	}
}

func TestEmitStructuredYAML(t *testing.T) {
	data := map[string]string{"foo": "bar"}

	output := captureStdout(func() {
		EmitStructured(data, OutputFormatYAML)
	})

	if !strings.Contains(output, "ok: true") {
		t.Errorf("Expected 'ok: true' in YAML output")
	}
	if !strings.Contains(output, "foo: bar") {
		t.Errorf("Expected 'foo: bar' in YAML output")
	}
}
