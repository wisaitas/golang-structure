package httpx

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type WrappedError struct {
	Op         string
	StatusCode int
	Function   string
	File       string
	Line       int
	Err        error
}

func (e *WrappedError) Error() string {
	if e == nil {
		return ""
	}
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *WrappedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func WrapError(op string, err error, statusCode int) error {
	if err == nil {
		return nil
	}
	pc, file, line, ok := runtime.Caller(1)
	function := ""
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			function = fn.Name()
		}
	}

	return &WrappedError{
		Op:         op,
		StatusCode: statusCode,
		Function:   function,
		File:       file,
		Line:       line,
		Err:        err,
	}
}

func StatusCodeFromError(err error, fallback int) int {
	for current := err; current != nil; current = errors.Unwrap(current) {
		wrappedErr, ok := current.(*WrappedError)
		if ok && wrappedErr.StatusCode > 0 {
			return wrappedErr.StatusCode
		}
	}
	return fallback
}

func FormatErrorChain(err error) string {
	if err == nil {
		return ""
	}

	parts := make([]string, 0, 4)
	for current := err; current != nil; current = errors.Unwrap(current) {
		switch e := current.(type) {
		case *WrappedError:
			if e.Op != "" {
				parts = append(parts, e.Op)
			}
		default:
			parts = append(parts, current.Error())
		}
	}

	return strings.Join(parts, " -> ")
}

func BuildErrorStackTraces(err error) []string {
	if err == nil {
		return nil
	}

	traces := make([]string, 0, 6)
	for current := err; current != nil; current = errors.Unwrap(current) {
		switch e := current.(type) {
		case *WrappedError:
			if e.Op != "" {
				fileLine := ""
				if e.File != "" && e.Line > 0 {
					fileLine = fmt.Sprintf("%s:%d", filepath.Base(e.File), e.Line)
				}
				switch {
				case e.Function != "" && fileLine != "":
					traces = append(traces, fmt.Sprintf("%s (%s @ %s)", e.Op, e.Function, fileLine))
				case e.Function != "":
					traces = append(traces, fmt.Sprintf("%s (%s)", e.Op, e.Function))
				case fileLine != "":
					traces = append(traces, fmt.Sprintf("%s (@ %s)", e.Op, fileLine))
				default:
					traces = append(traces, e.Op)
				}
			}
			if errors.Unwrap(current) == nil && e.Err != nil {
				traces = append(traces, e.Err.Error())
			}
		default:
			traces = append(traces, current.Error())
		}
	}

	if len(traces) == 0 {
		return nil
	}
	return traces
}

func RootErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	last := err
	for current := err; current != nil; current = errors.Unwrap(current) {
		last = current
	}
	return last.Error()
}
