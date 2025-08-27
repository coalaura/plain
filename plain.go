package plain

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	Reset = "\x1b[0m"

	Text  = "\x1b[37m"
	Input = "\x1b[97m"
	Warn  = "\x1b[33m"
	Error = "\x1b[31m"
)

type Plain struct {
	out   *os.File
	color bool
}

func New(out *os.File) *Plain {
	if out == nil {
		out = os.Stdout
	}

	return &Plain{
		out:   out,
		color: term.IsTerminal(int(out.Fd())),
	}
}

func (p *Plain) Write(code, msg string, reset bool) {
	if !p.color {
		p.out.WriteString(msg)

		return
	}

	if reset {
		msg += Reset
	}

	p.out.WriteString(code + msg)
}

func sprint(a ...any) string {
	if len(a) == 0 {
		return ""
	}

	return fmt.Sprint(a...)
}

func sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return format
	}

	return fmt.Sprintf(format, a...)
}
