package plain

import "io"

type option func(*Plain)

// WithTarget sets the output writer used by the logger
func WithTarget(out io.Writer) option {
	return func(p *Plain) {
		p.out = out
	}
}

// WithDate sets the timestamp format used in log headers (empty disables timestamps)
func WithDate(format string) option {
	return func(p *Plain) {
		p.format = format
	}
}

// SetTarget updates the output writer used by the logger
func (p *Plain) SetTarget(out io.Writer) {
	p.writeLock.Lock()
	defer p.writeLock.Unlock()

	p.out = out
}

// SetDate updates the timestamp format used in log headers (empty disables timestamps)
func (p *Plain) SetDate(format string) {
	p.format = format
}
