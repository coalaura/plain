package minimal

import "os"

type option func(*Minimal)

func ToStdErr() option {
	return func(m *Minimal) {
		m.out = os.Stderr
	}
}
