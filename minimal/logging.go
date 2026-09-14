package minimal

import "os"

// Sub writes its operands to the standard target as a sub-message.
func (m *Minimal) Sub(a ...any) error {
	return m.writeArgs(false, true, AnsiSub, prefixSub, a...)
}

// Subf formats according to a format specifier and writes to the standard target as a sub-message.
func (m *Minimal) Subf(format string, a ...any) error {
	return m.writeFormat(false, true, AnsiSub, prefixSub, format, a...)
}

// Subln writes its operands to the standard target as a sub-message with a trailing newline.
func (m *Minimal) Subln(a ...any) error {
	return m.writeArgsLine(false, true, AnsiSub, prefixSub, a...)
}

// Step writes its operands to the standard target as a step message.
func (m *Minimal) Step(a ...any) error {
	return m.writeArgs(false, false, AnsiStep, prefixStep, a...)
}

// Stepf formats according to a format specifier and writes to the standard target as a step message.
func (m *Minimal) Stepf(format string, a ...any) error {
	return m.writeFormat(false, false, AnsiStep, prefixStep, format, a...)
}

// Stepln writes its operands to the standard target as a step message with a trailing newline.
func (m *Minimal) Stepln(a ...any) error {
	return m.writeArgsLine(false, false, AnsiStep, prefixStep, a...)
}

// Info writes its operands to the standard target as an info message.
func (m *Minimal) Info(a ...any) error {
	return m.writeArgs(false, false, AnsiInfo, prefixInfo, a...)
}

// Infof formats according to a format specifier and writes to the standard target as an info message.
func (m *Minimal) Infof(format string, a ...any) error {
	return m.writeFormat(false, false, AnsiInfo, prefixInfo, format, a...)
}

// Infoln writes its operands to the standard target as an info message with a trailing newline.
func (m *Minimal) Infoln(a ...any) error {
	return m.writeArgsLine(false, false, AnsiInfo, prefixInfo, a...)
}

// Success writes its operands to the standard target as a success message.
func (m *Minimal) Success(a ...any) error {
	return m.writeArgs(false, false, AnsiSuccess, prefixSuccess, a...)
}

// Successf formats according to a format specifier and writes to the standard target as a success message.
func (m *Minimal) Successf(format string, a ...any) error {
	return m.writeFormat(false, false, AnsiSuccess, prefixSuccess, format, a...)
}

// Successln writes its operands to the standard target as a success message with a trailing newline.
func (m *Minimal) Successln(a ...any) error {
	return m.writeArgsLine(false, false, AnsiSuccess, prefixSuccess, a...)
}

// Warn writes its operands to the standard target as a warning.
func (m *Minimal) Warn(a ...any) error {
	return m.writeArgs(false, false, AnsiWarn, prefixWarn, a...)
}

// Warnf formats according to a format specifier and writes to the standard target as a warning.
func (m *Minimal) Warnf(format string, a ...any) error {
	return m.writeFormat(false, false, AnsiWarn, prefixWarn, format, a...)
}

// Warnln writes its operands to the standard target as a warning with a trailing newline.
func (m *Minimal) Warnln(a ...any) error {
	return m.writeArgsLine(false, false, AnsiWarn, prefixWarn, a...)
}

// Error writes its operands to the error target as an error message.
func (m *Minimal) Error(a ...any) error {
	return m.writeArgs(true, false, AnsiError, prefixError, a...)
}

// Errorf formats according to a format specifier and writes to the error target as an error message.
func (m *Minimal) Errorf(format string, a ...any) error {
	return m.writeFormat(true, false, AnsiError, prefixError, format, a...)
}

// Errorln writes its operands to the error target as an error message with a trailing newline.
func (m *Minimal) Errorln(a ...any) error {
	return m.writeArgsLine(true, false, AnsiError, prefixError, a...)
}

// Write writes its operands to the standard target using the provided ANSI color code.
func (m *Minimal) Write(color string, a ...any) error {
	return m.writeArgs(false, false, color, "", a...)
}

// Writef formats according to a format specifier and writes to the standard target using the provided ANSI color code.
func (m *Minimal) Writef(color, format string, a ...any) error {
	return m.writeFormat(false, false, color, "", format, a...)
}

// Writeln writes its operands to the standard target using the provided ANSI color code with a trailing newline.
func (m *Minimal) Writeln(color string, a ...any) error {
	return m.writeArgsLine(false, false, color, "", a...)
}

// MustFail logs and panics if err is not nil.
func (m *Minimal) MustFail(err error) {
	if err == nil {
		return
	}

	m.writeArgsLine(true, false, AnsiError, prefixError, err)

	panic(err)
}

// MustExit logs and exits with code 1 if err is not nil.
func (m *Minimal) MustExit(err error) {
	if err == nil {
		return
	}

	m.writeArgsLine(true, false, AnsiError, prefixError, err)

	os.Exit(1)
}
