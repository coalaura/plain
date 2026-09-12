package minimal

import "os"

func (m *Minimal) Sub(a ...any) {
	m.writeArgs(AnsiSub, prefixSub, a...)
}

func (m *Minimal) Subf(format string, a ...any) {
	m.writeFormat(AnsiSub, prefixSub, format, a...)
}

func (m *Minimal) Subln(a ...any) {
	m.writeArgsLine(AnsiSub, prefixSub, a...)
}

func (m *Minimal) Info(a ...any) {
	m.writeArgs(AnsiInfo, prefixInfo, a...)
}

func (m *Minimal) Infof(format string, a ...any) {
	m.writeFormat(AnsiInfo, prefixInfo, format, a...)
}

func (m *Minimal) Infoln(a ...any) {
	m.writeArgsLine(AnsiInfo, prefixInfo, a...)
}

func (m *Minimal) Success(a ...any) {
	m.writeArgs(AnsiSuccess, prefixSuccess, a...)
}

func (m *Minimal) Successf(format string, a ...any) {
	m.writeFormat(AnsiSuccess, prefixSuccess, format, a...)
}

func (m *Minimal) Successln(a ...any) {
	m.writeArgsLine(AnsiSuccess, prefixSuccess, a...)
}

func (m *Minimal) Error(a ...any) {
	m.writeArgs(AnsiError, prefixError, a...)
}

func (m *Minimal) Errorf(format string, a ...any) {
	m.writeFormat(AnsiError, prefixError, format, a...)
}

func (m *Minimal) Errorln(a ...any) {
	m.writeArgsLine(AnsiError, prefixError, a...)
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
