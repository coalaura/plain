package plain

import "os"

// Debugf formats according to a format specifier and writes to the target output as a debug log.
func (p *Plain) Debugf(format string, a ...any) {
	p.Write(p.theme.Dimmed, sprintf(format, a...), true, false)
}

// Debug formats using the default formats for its operands and writes to the target output as a debug log.
func (p *Plain) Debug(a ...any) {
	p.Write(p.theme.Dimmed, sprint(a...), true, false)
}

// Debugln formats using the default formats for its operands and writes to the target output as a debug log with a trailing newline.
func (p *Plain) Debugln(a ...any) {
	p.Write(p.theme.Dimmed, sprint(a...), true, true)
}

// Printf formats according to a format specifier and writes to the target output.
func (p *Plain) Printf(format string, a ...any) {
	p.Write(Reset, sprintf(format, a...), true, false)
}

// Print formats using the default formats for its operands and writes to the target output.
func (p *Plain) Print(a ...any) {
	p.Write(Reset, sprint(a...), true, false)
}

// Println formats using the default formats for its operands and writes to the target output with a trailing newline.
func (p *Plain) Println(a ...any) {
	p.Write(Reset, sprint(a...), true, true)
}

// Warnf formats according to a format specifier and writes to the target output as a warning.
func (p *Plain) Warnf(format string, a ...any) {
	p.Write(p.theme.Warn, sprintf(format, a...), true, false)
}

// Warn formats using the default formats for its operands and writes to the target output as a warning.
func (p *Plain) Warn(a ...any) {
	p.Write(p.theme.Warn, sprint(a...), true, false)
}

// Warnln formats using the default formats for its operands and writes to the target output as a warning with a trailing newline.
func (p *Plain) Warnln(a ...any) {
	p.Write(p.theme.Warn, sprint(a...), true, true)
}

// Errorf formats according to a format specifier and writes to the target output as an error.
func (p *Plain) Errorf(format string, a ...any) {
	p.Write(p.theme.Error, sprintf(format, a...), true, false)
}

// Error formats using the default formats for its operands and writes to the target output as an error.
func (p *Plain) Error(a ...any) {
	p.Write(p.theme.Error, sprint(a...), true, false)
}

// Errorln formats using the default formats for its operands and writes to the target output as an error with a trailing newline.
func (p *Plain) Errorln(a ...any) {
	p.Write(p.theme.Error, sprint(a...), true, true)
}

// MustFail panics if err is not nil
func (p *Plain) MustFail(err error) {
	if err == nil {
		return
	}

	panic(err)
}

// MustExit logs the error and exits with code 1 if the error is not nil.
func (p *Plain) MustExit(err error) {
	if err == nil {
		return
	}

	p.Errorln(err)

	os.Exit(1)
}
