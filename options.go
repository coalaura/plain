package plain

import "os"

type option func(*Plain)

func WithTarget(out *os.File) option {
	return func(p *Plain) {
		p.out = out
	}
}

func WithDate(format string) option {
	return func(p *Plain) {
		p.format = format
	}
}
