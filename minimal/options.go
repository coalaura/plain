package minimal

import (
	"io"
	"os"
)

type option func(*Minimal)

// WithTarget sets the standard output writer used by the logger.
func WithTarget(out io.Writer) option {
	return func(m *Minimal) {
		m.out = out
	}
}

// WithErrorTarget sets the error output writer used by the logger.
func WithErrorTarget(out io.Writer) option {
	return func(m *Minimal) {
		m.errOut = out
	}
}

// WithColor controls when the logger emits ANSI color sequences.
func WithColor(mode ColorMode) option {
	return func(m *Minimal) {
		m.colorMode = mode
	}
}

// ToStdErr sets both output writers to standard error.
func ToStdErr() option {
	return func(m *Minimal) {
		m.out = os.Stderr
		m.errOut = os.Stderr
	}
}
