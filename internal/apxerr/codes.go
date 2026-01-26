// Package apxerr provides structured error codes for AgentPlexus Console.
//
// Error codes follow HTTP semantics:
//   - APX-4xx: Content/user errors (fixable by the user editing their specs)
//   - APX-5xx: System errors (bugs in AgentPlexus Console)
package apxerr

import (
	"fmt"
	"log/slog"
)

// Code represents an APX error code.
type Code string

// Content errors (4xx) - issues with user-provided specs
const (
	// InvalidStructure indicates the spec has an invalid structure.
	InvalidStructure Code = "APX-400"

	// MissingField indicates a required or expected field is missing.
	MissingField Code = "APX-401"

	// InvalidFieldType indicates a field has an unexpected type.
	InvalidFieldType Code = "APX-402"

	// InvalidReference indicates a reference to a non-existent resource.
	InvalidReference Code = "APX-403"

	// ResourceNotFound indicates a requested resource was not found.
	ResourceNotFound Code = "APX-404"
)

// System errors (5xx) - issues with AgentPlexus Console itself
const (
	// InternalError indicates an internal system error.
	InternalError Code = "APX-500"

	// NotImplemented indicates a feature is not yet implemented.
	NotImplemented Code = "APX-501"
)

// Warning represents a validation warning with structured context.
type Warning struct {
	Code    Code
	Message string
	Team    string
	Step    string
	Field   string
	File    string
}

// String returns a formatted warning string for console output.
func (w Warning) String() string {
	return fmt.Sprintf("[%s] %s", w.Code, w.Message)
}

// LogAttrs returns slog attributes for structured logging.
func (w Warning) LogAttrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("code", string(w.Code)),
	}
	if w.Team != "" {
		attrs = append(attrs, slog.String("team", w.Team))
	}
	if w.Step != "" {
		attrs = append(attrs, slog.String("step", w.Step))
	}
	if w.Field != "" {
		attrs = append(attrs, slog.String("field", w.Field))
	}
	if w.File != "" {
		attrs = append(attrs, slog.String("file", w.File))
	}
	return attrs
}

// LogWarn logs the warning using slog.
func (w Warning) LogWarn(logger *slog.Logger) {
	attrs := make([]any, 0, len(w.LogAttrs())*2)
	for _, a := range w.LogAttrs() {
		attrs = append(attrs, a.Key, a.Value.Any())
	}
	logger.Warn(w.Message, attrs...)
}

// ToJSON returns a JSON-serializable representation for API responses.
func (w Warning) ToJSON() map[string]string {
	m := map[string]string{
		"code":    string(w.Code),
		"message": w.Message,
	}
	if w.Team != "" {
		m["team"] = w.Team
	}
	if w.Step != "" {
		m["step"] = w.Step
	}
	if w.Field != "" {
		m["field"] = w.Field
	}
	if w.File != "" {
		m["file"] = w.File
	}
	return m
}

// FormatConsoleJS returns a string suitable for browser console.warn().
func (w Warning) FormatConsoleJS() string {
	base := fmt.Sprintf("[%s] Warning: %s", w.Code, w.Message)
	details := ""
	if w.Team != "" {
		details += fmt.Sprintf("\n  Team: %s", w.Team)
	}
	if w.Step != "" {
		details += fmt.Sprintf(", Step: %s", w.Step)
	}
	if w.Field != "" {
		details += fmt.Sprintf(", Field: %s", w.Field)
	}
	if w.File != "" {
		details += fmt.Sprintf("\n  File: %s", w.File)
	}
	return base + details
}
