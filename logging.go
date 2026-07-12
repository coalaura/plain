package plain

import "os"

// Debugf formats according to a format specifier and writes to the target output as a debug log.
func (p *Plain) Debugf(format string, a ...any) {
	if p.level.Load() > LevelDebug {
		return
	}

	p.writeFormat(p.theme.Dimmed, true, false, format, a...)
}

// Debug formats using the default formats for its operands and writes to the target output as a debug log.
func (p *Plain) Debug(a ...any) {
	if p.level.Load() > LevelDebug {
		return
	}

	p.writeArgs(p.theme.Dimmed, true, false, a...)
}

// Debugln formats using the default formats for its operands and writes to the target output as a debug log with a trailing newline.
func (p *Plain) Debugln(a ...any) {
	if p.level.Load() > LevelDebug {
		return
	}

	p.writeArgsLine(p.theme.Dimmed, true, a...)
}

// Printf formats according to a format specifier and writes to the target output.
func (p *Plain) Printf(format string, a ...any) {
	if p.level.Load() > LevelPrint {
		return
	}

	p.writeFormat(ansiReset, true, false, format, a...)
}

// Print formats using the default formats for its operands and writes to the target output.
func (p *Plain) Print(a ...any) {
	if p.level.Load() > LevelPrint {
		return
	}

	p.writeArgs(ansiReset, true, false, a...)
}

// Println formats using the default formats for its operands and writes to the target output with a trailing newline.
func (p *Plain) Println(a ...any) {
	if p.level.Load() > LevelPrint {
		return
	}

	p.writeArgsLine(ansiReset, true, a...)
}

// Warnf formats according to a format specifier and writes to the target output as a warning.
func (p *Plain) Warnf(format string, a ...any) {
	if p.level.Load() > LevelWarn {
		return
	}

	p.writeFormat(p.theme.Warn, true, false, format, a...)
}

// Warn formats using the default formats for its operands and writes to the target output as a warning.
func (p *Plain) Warn(a ...any) {
	if p.level.Load() > LevelWarn {
		return
	}

	p.writeArgs(p.theme.Warn, true, false, a...)
}

// Warnln formats using the default formats for its operands and writes to the target output as a warning with a trailing newline.
func (p *Plain) Warnln(a ...any) {
	if p.level.Load() > LevelWarn {
		return
	}

	p.writeArgsLine(p.theme.Warn, true, a...)
}

// Errorf formats according to a format specifier and writes to the target output as an error.
func (p *Plain) Errorf(format string, a ...any) {
	if p.level.Load() > LevelError {
		return
	}

	p.writeFormat(p.theme.Error, true, false, format, a...)
}

// Error formats using the default formats for its operands and writes to the target output as an error.
func (p *Plain) Error(a ...any) {
	if p.level.Load() > LevelError {
		return
	}

	p.writeArgs(p.theme.Error, true, false, a...)
}

// Errorln formats using the default formats for its operands and writes to the target output as an error with a trailing newline.
func (p *Plain) Errorln(a ...any) {
	if p.level.Load() > LevelError {
		return
	}

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

	p.writeArgsLine(p.theme.Error, true, err)

	os.Exit(1)
}

//go:fix inline
func (p *Plain) Infof(format string, a ...any) {
	p.Printf(format, a...)
}

//go:fix inline
func (p *Plain) Info(a ...any) {
	p.Print(a...)
}

//go:fix inline
func (p *Plain) Infoln(a ...any) {
	p.Println(a...)
}
