package plain

import (
	"os"
)

// Printf formats according to a format specifier and writes to the target output.
func (p *Plain) Printf(format string, a ...any) {
	p.Write(Text, sprintf(format, a...), true)
}

// Print formats using the default formats for its operands and writes to the target output.
func (p *Plain) Print(a ...any) {
	p.Write(Text, sprint(a...), true)
}

// Println formats using the default formats for its operands and writes to the target output with a trailing newline.
func (p *Plain) Println(a ...any) {
	p.Write(Text, sprint(a...)+"\n", true)
}

// Warnf formats according to a format specifier and writes to the target output as a warning.
func (p *Plain) Warnf(format string, a ...any) {
	p.Write(Warn, sprintf(format, a...), true)
}

// Warn formats using the default formats for its operands and writes to the target output as a warning.
func (p *Plain) Warn(a ...any) {
	p.Write(Warn, sprint(a...), true)
}

// Warnln formats using the default formats for its operands and writes to the target output as a warning with a trailing newline.
func (p *Plain) Warnln(a ...any) {
	p.Write(Warn, sprint(a...)+"\n", true)
}

// Errorf formats according to a format specifier and writes to the target output as an error.
func (p *Plain) Errorf(format string, a ...any) {
	p.Write(Error, sprintf(format, a...), true)
}

// Error formats using the default formats for its operands and writes to the target output as an error.
func (p *Plain) Error(a ...any) {
	p.Write(Error, sprint(a...), true)
}

// Errorln formats using the default formats for its operands and writes to the target output as an error with a trailing newline.
func (p *Plain) Errorln(a ...any) {
	p.Write(Error, sprint(a...)+"\n", true)
}

// MustFail logs the error and exits with code 1 if the error is not nil.
func (p *Plain) MustFail(err error) {
	if err == nil {
		return
	}

	p.Errorln(err)

	os.Exit(1)
}
