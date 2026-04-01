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
	Code       ResponseCode
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

func wrapError(op string, err error, statusCode int, code ResponseCode) error {
	if err == nil {
		return nil
	}
	function, file, line, _ := callerOutsideHttpx()

	return &WrappedError{
		Op:         op,
		StatusCode: statusCode,
		Code:       code,
		Function:   function,
		File:       file,
		Line:       line,
		Err:        err,
	}
}

func callerOutsideHttpx() (function, file string, line int, ok bool) {
	// skip: callerOutsideHttpx(0) → wrapError(1) → WrapError/WrapErrorWithCode(2) → actual caller(3+)
	for skip := 3; skip < 8; skip++ {
		pc, f, l, found := runtime.Caller(skip)
		if !found {
			return "", "", 0, false
		}
		if isHttpxSourceFile(f) {
			continue
		}
		if fn := runtime.FuncForPC(pc); fn != nil {
			return fn.Name(), f, l, true
		}
		return "", f, l, true
	}
	return "", "", 0, false
}

func isHttpxSourceFile(file string) bool {
	return strings.Contains(file, "/pkg/httpx/") || strings.Contains(file, `\pkg\httpx\`)
}

func WrapError(op string, err error, statusCode int) error {
	return wrapError(op, err, statusCode, "")
}

func WrapErrorWithCode(op string, err error, statusCode int, code ResponseCode) error {
	return wrapError(op, err, statusCode, code)
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

func ResponseCodeFromError(err error) ResponseCode {
	var code ResponseCode
	for current := err; current != nil; current = errors.Unwrap(current) {
		if w, ok := current.(*WrappedError); ok && w.Code != "" {
			code = w.Code
		}
	}
	return code
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

func formatWrappedErrorLine(e *WrappedError) string {
	if e == nil || e.Op == "" {
		return ""
	}
	fileLine := ""
	if e.File != "" && e.Line > 0 {
		fileLine = fmt.Sprintf("%s:%d", filepath.Base(e.File), e.Line)
	}
	switch {
	case e.Function != "" && fileLine != "":
		return fmt.Sprintf("%s (%s @ %s)", e.Op, e.Function, fileLine)
	case e.Function != "":
		return fmt.Sprintf("%s (%s)", e.Op, e.Function)
	case fileLine != "":
		return fmt.Sprintf("%s (@ %s)", e.Op, fileLine)
	default:
		return e.Op
	}
}

func BuildErrorStackTraces(err error) []string {
	if err == nil {
		return nil
	}

	traces := make([]string, 0, 4)
	for current := err; current != nil; {
		switch e := current.(type) {
		case *WrappedError:
			if line := formatWrappedErrorLine(e); line != "" {
				traces = append(traces, line)
			}
			if e.Err == nil {
				return traces
			}
			current = e.Err
		default:
			traces = append(traces, current.Error())
			return traces
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
