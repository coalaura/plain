package minimal

import "os"

type option func(*Minimal)

// ToStdErr sets the output writer used by the logger to standard error
func ToStdErr() option {
	return func(m *Minimal) {
		m.out = os.Stderr
	}
}
