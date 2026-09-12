package minimal

import "os"

// Sub writes its operands to the target output as a sub-message.
func (m *Minimal) Sub(a ...any) {
	m.writeArgs(AnsiSub, prefixSub, a...)
}

// Subf formats according to a format specifier and writes to the target output as a sub-message.
func (m *Minimal) Subf(format string, a ...any) {
	m.writeFormat(AnsiSub, prefixSub, format, a...)
}

// Subln formats using the default formats for its operands and writes to the target output as a sub-message with a trailing newline.
func (m *Minimal) Subln(a ...any) {
	m.writeArgsLine(AnsiSub, prefixSub, a...)
}

// Info writes its operands to the target output as an info message.
func (m *Minimal) Info(a ...any) {
	m.writeArgs(AnsiInfo, prefixInfo, a...)
}

// Infof formats according to a format specifier and writes to the target output as an info message.
func (m *Minimal) Infof(format string, a ...any) {
	m.writeFormat(AnsiInfo, prefixInfo, format, a...)
}

// Infoln formats using the default formats for its operands and writes to the target output as an info message with a trailing newline.
func (m *Minimal) Infoln(a ...any) {
	m.writeArgsLine(AnsiInfo, prefixInfo, a...)
}

// Success writes its operands to the target output as a success message.
func (m *Minimal) Success(a ...any) {
	m.writeArgs(AnsiSuccess, prefixSuccess, a...)
}

// Successf formats according to a format specifier and writes to the target output as a success message.
func (m *Minimal) Successf(format string, a ...any) {
	m.writeFormat(AnsiSuccess, prefixSuccess, format, a...)
}

// Successln formats using the default formats for its operands and writes to the target output as a success message with a trailing newline.
func (m *Minimal) Successln(a ...any) {
	m.writeArgsLine(AnsiSuccess, prefixSuccess, a...)
}

// Error writes its operands to the target output as an error message.
func (m *Minimal) Error(a ...any) {
	m.writeArgs(AnsiError, prefixError, a...)
}

// Errorf formats according to a format specifier and writes to the target output as an error message.
func (m *Minimal) Errorf(format string, a ...any) {
	m.writeFormat(AnsiError, prefixError, format, a...)
}

// Errorln formats using the default formats for its operands and writes to the target output as an error message with a trailing newline.
func (m *Minimal) Errorln(a ...any) {
	m.writeArgsLine(AnsiError, prefixError, a...)
}

// Write writes its operands to the target output using the provided ANSI color code.
func (m *Minimal) Write(color string, a ...any) {
	m.writeArgs(color, "", a...)
}

// Writef formats according to a format specifier and writes to the target output using the provided ANSI color code.
func (m *Minimal) Writef(color, format string, a ...any) {
	m.writeFormat(color, "", format, a...)
}

// Writeln formats using the default formats for its operands and writes to the target output using the provided ANSI color code with a trailing newline.
func (m *Minimal) Writeln(color string, a ...any) {
	m.writeArgsLine(color, "", a...)
}

// MustFail panics if err is not nil
func (m *Minimal) MustFail(err error) {
	if err == nil {
		return
	}

	m.writeArgsLine(AnsiError, prefixError, err)

	panic(err)
}

// MustExit logs the error and exits with code 1 if the error is not nil.
func (m *Minimal) MustExit(err error) {
	if err == nil {
		return
	}

	m.writeArgsLine(AnsiError, prefixError, err)

	os.Exit(1)
}
