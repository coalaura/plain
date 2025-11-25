package plain

import "io"

type option func(*Plain)

func WithTarget(out io.Writer) option {
	return func(p *Plain) {
		p.out = out
	}
}

func WithDate(format string) option {
	return func(p *Plain) {
		p.format = format
	}
}
