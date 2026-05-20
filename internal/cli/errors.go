package cli

import "fmt"

type commandError struct {
	code int
	err  error
}

func (e *commandError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *commandError) Unwrap() error {
	return e.err
}

func invalidError(format string, args ...any) error {
	return &commandError{code: 2, err: fmt.Errorf(format, args...)}
}

func userError(format string, args ...any) error {
	return &commandError{code: 1, err: fmt.Errorf(format, args...)}
}

func internalError(err error) error {
	if err == nil {
		return nil
	}
	return &commandError{code: 3, err: err}
}
