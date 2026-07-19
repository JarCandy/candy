package sdk

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type DiagnosticError struct {
	Message string
	Hint    string
	Detail  string
	Cause   error
}

func NewError(text string) *DiagnosticError {
	return &DiagnosticError{Message: strings.TrimSpace(text)}
}

func NewErrorf(format string, args ...any) *DiagnosticError {
	return NewError(fmt.Sprintf(format, args...))
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	if strings.TrimSpace(err.Message) == "" {
		if err.Cause != nil {
			return err.Cause.Error()
		}
		return "unknown error"
	}
	if err.Cause != nil {
		return err.Message + ": " + err.Cause.Error()
	}
	return err.Message
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

func (err *DiagnosticError) WithHint(text string) *DiagnosticError {
	if err == nil {
		return nil
	}
	err.Hint = strings.TrimSpace(text)
	return err
}

func (err *DiagnosticError) WithDetail(text string) *DiagnosticError {
	if err == nil {
		return nil
	}
	err.Detail = strings.TrimSpace(text)
	return err
}

func (err *DiagnosticError) WithCause(cause error) *DiagnosticError {
	if err == nil {
		return nil
	}
	err.Cause = cause
	return err
}

func PrintError(writer io.Writer, err error) error {
	if writer == nil || err == nil {
		return nil
	}

	var diagnostic *DiagnosticError
	ok := errors.As(err, &diagnostic)
	if !ok {
		return writeLine(writer, err.Error())
	}

	if message := strings.TrimSpace(diagnostic.Message); message != "" {
		if err := writeLine(writer, message); err != nil {
			return err
		}
	} else if diagnostic.Cause != nil {
		if err := writeLine(writer, diagnostic.Cause.Error()); err != nil {
			return err
		}
	}
	if detail := strings.TrimSpace(diagnostic.Detail); detail != "" {
		if err := writeLine(writer, detail); err != nil {
			return err
		}
	}
	if hint := strings.TrimSpace(diagnostic.Hint); hint != "" {
		if err := writeLine(writer, "hint: "+hint); err != nil {
			return err
		}
	}
	if diagnostic.Cause != nil && strings.TrimSpace(diagnostic.Detail) == "" {
		if err := writeLine(writer, "cause: "+diagnostic.Cause.Error()); err != nil {
			return err
		}
	}
	return nil
}

func writeLine(writer io.Writer, text string) error {
	_, err := fmt.Fprintln(writer, text)
	return err
}

func Text(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}
