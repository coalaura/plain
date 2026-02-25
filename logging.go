package plain

import "os"

// Debugf formats according to a format specifier and writes to the target output as a debug log.
func (p *Plain) Debugf(format string, a ...any) {
	p.writeFormat(p.theme.Dimmed, true, false, format, a...)
}

// Debug formats using the default formats for its operands and writes to the target output as a debug log.
func (p *Plain) Debug(a ...any) {
	p.writeArgs(p.theme.Dimmed, true, false, a...)
}

// Debugln formats using the default formats for its operands and writes to the target output as a debug log with a trailing newline.
func (p *Plain) Debugln(a ...any) {
	p.writeArgsLine(p.theme.Dimmed, true, a...)
}

// Printf formats according to a format specifier and writes to the target output.
func (p *Plain) Printf(format string, a ...any) {
	p.writeFormat(ansiReset, true, false, format, a...)
}

// Print formats using the default formats for its operands and writes to the target output.
func (p *Plain) Print(a ...any) {
	p.writeArgs(ansiReset, true, false, a...)
}

// Println formats using the default formats for its operands and writes to the target output with a trailing newline.
func (p *Plain) Println(a ...any) {
	p.writeArgsLine(ansiReset, true, a...)
}

// Warnf formats according to a format specifier and writes to the target output as a warning.
func (p *Plain) Warnf(format string, a ...any) {
	p.writeFormat(p.theme.Warn, true, false, format, a...)
}

// Warn formats using the default formats for its operands and writes to the target output as a warning.
func (p *Plain) Warn(a ...any) {
	p.writeArgs(p.theme.Warn, true, false, a...)
}

// Warnln formats using the default formats for its operands and writes to the target output as a warning with a trailing newline.
func (p *Plain) Warnln(a ...any) {
	p.writeArgsLine(p.theme.Warn, true, a...)
}

// Errorf formats according to a format specifier and writes to the target output as an error.
func (p *Plain) Errorf(format string, a ...any) {
	p.writeFormat(p.theme.Error, true, false, format, a...)
}

// Error formats using the default formats for its operands and writes to the target output as an error.
func (p *Plain) Error(a ...any) {
	p.writeArgs(p.theme.Error, true, false, a...)
}

// Errorln formats using the default formats for its operands and writes to the target output as an error with a trailing newline.
func (p *Plain) Errorln(a ...any) {
	p.writeArgsLine(p.theme.Error, true, a...)
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
