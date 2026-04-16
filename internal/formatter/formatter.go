package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "1"

type OutputFormat string

const (
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
	OutputFormatNone OutputFormat = "none"
)

type SuccessEnvelope struct {
	OK            bool        `json:"ok" yaml:"ok"`
	SchemaVersion string      `json:"schema_version" yaml:"schema_version"`
	Data          interface{} `json:"data" yaml:"data"`
}

type ErrorEnvelope struct {
	OK            bool        `json:"ok" yaml:"ok"`
	SchemaVersion string      `json:"schema_version" yaml:"schema_version"`
	Error         ErrorDetail `json:"error" yaml:"error"`
}

type ErrorDetail struct {
	Code    string      `json:"code" yaml:"code"`
	Message string      `json:"message" yaml:"message"`
	Details interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

func ResolveOutputFormat(asJSON, asYAML bool) OutputFormat {
	if asJSON && asYAML {
		fmt.Fprintln(os.Stderr, "❌ 不能同时使用 --json 和 --yaml。")
		os.Exit(1)
	}
	if asYAML {
		return OutputFormatYAML
	}
	if asJSON {
		return OutputFormatJSON
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("OUTPUT")))
	if env == "yaml" {
		return OutputFormatYAML
	}
	if env == "json" {
		return OutputFormatJSON
	}
	if env == "rich" {
		return OutputFormatNone
	}

	// Check if stdout is not a tty
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return OutputFormatYAML
	}

	return OutputFormatNone
}

func EmitStructured(data interface{}, format OutputFormat) bool {
	if format == OutputFormatNone {
		return false
	}

	envelope := SuccessEnvelope{
		OK:            true,
		SchemaVersion: SchemaVersion,
		Data:          data,
	}

	if format == OutputFormatJSON {
		b, _ := json.MarshalIndent(envelope, "", "  ")
		fmt.Println(string(b))
		return true
	} else if format == OutputFormatYAML {
		b, _ := yaml.Marshal(envelope)
		fmt.Print(string(b))
		return true
	}
	return false
}

func EmitError(code, message string, details interface{}, format OutputFormat) {
	envelope := ErrorEnvelope{
		OK:            false,
		SchemaVersion: SchemaVersion,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	if format == OutputFormatJSON {
		b, _ := json.MarshalIndent(envelope, "", "  ")
		fmt.Println(string(b))
	} else if format == OutputFormatYAML {
		b, _ := yaml.Marshal(envelope)
		fmt.Print(string(b))
	}
}

func ExitError(code, message string, format OutputFormat) {
	if format != OutputFormatNone {
		EmitError(code, message, nil, format)
	} else {
		fmt.Fprintf(os.Stderr, "❌ %s\n", message)
	}
	panic("exit:1")
}

func FormatDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	if seconds >= 3600 {
		h := seconds / 3600
		rem := seconds % 3600
		m := rem / 60
		s := rem % 60
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func FormatCount(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000.0)
	}
	return fmt.Sprintf("%d", n)
}
