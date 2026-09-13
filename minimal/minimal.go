package minimal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/coalaura/plain/internal"
	"golang.org/x/term"
)

const (
	AnsiSub     = "\033[90m"
	AnsiInfo    = "\033[36m"
	AnsiSuccess = "\033[32m"
	AnsiWarn    = "\033[33m"
	AnsiError   = "\033[31m"
	AnsiReset   = internal.AnsiReset

	prefixSub     = "   -> "
	prefixInfo    = ":: "
	prefixSuccess = ":: "
	prefixWarn    = "?? "
	prefixError   = "!! "
)

type ColorMode uint8

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

type Minimal struct {
	mx sync.Mutex

	out    io.Writer
	errOut io.Writer

	outColor  bool
	errColor  bool
	colorMode ColorMode
}

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

// New creates a Minimal logger configured by the provided options.
func New(opts ...option) *Minimal {
	m := &Minimal{
		out:       os.Stdout,
		errOut:    os.Stderr,
		colorMode: ColorAuto,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.outColor = detectColor(m.out, m.colorMode)
	m.errColor = detectColor(m.errOut, m.colorMode)

	return m
}

func (m *Minimal) writeArgs(errorOutput, colorMessage bool, code, prefix string, a ...any) error {
	out, colored := m.target(errorOutput)

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = appendPrefix(buf, colored, colorMessage, code, prefix)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	buf = finishMessage(buf, colored, colorMessage, code, prefix)
	buf = stripColor(buf, colored)

	err := m.write(out, buf)

	putBuffer(bp, buf)

	return err
}

func (m *Minimal) writeArgsLine(errorOutput, colorMessage bool, code, prefix string, a ...any) error {
	out, colored := m.target(errorOutput)

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = appendPrefix(buf, colored, colorMessage, code, prefix)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	buf = finishMessage(buf, colored, colorMessage, code, prefix)
	buf = append(buf, '\n')
	buf = stripColor(buf, colored)

	err := m.write(out, buf)

	putBuffer(bp, buf)

	return err
}

func (m *Minimal) writeFormat(errorOutput, colorMessage bool, code, prefix, format string, a ...any) error {
	out, colored := m.target(errorOutput)

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = appendPrefix(buf, colored, colorMessage, code, prefix)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	buf = finishMessage(buf, colored, colorMessage, code, prefix)
	buf = stripColor(buf, colored)

	err := m.write(out, buf)

	putBuffer(bp, buf)

	return err
}

func (m *Minimal) target(errorOutput bool) (io.Writer, bool) {
	if errorOutput {
		return m.errOut, m.errColor
	}

	return m.out, m.outColor
}

func (m *Minimal) write(out io.Writer, buf []byte) error {
	m.mx.Lock()
	n, err := out.Write(buf)
	m.mx.Unlock()

	if err == nil && n != len(buf) {
		return io.ErrShortWrite
	}

	return err
}

func appendPrefix(dst []byte, colored, colorMessage bool, code, prefix string) []byte {
	if colored && code != "" {
		dst = append(dst, code...)
	}

	dst = append(dst, prefix...)

	if colored && code != "" && prefix != "" && !colorMessage {
		dst = append(dst, AnsiReset...)
	}

	return dst
}

func finishMessage(dst []byte, colored, colorMessage bool, code, prefix string) []byte {
	if colored && code != "" && (prefix == "" || colorMessage) {
		return append(dst, AnsiReset...)
	}

	return dst
}

func stripColor(buf []byte, colored bool) []byte {
	if !colored && bytes.IndexByte(buf, '\x1b') >= 0 {
		return internal.StripANSI(buf)
	}

	return buf
}

func putBuffer(bp *[]byte, buf []byte) {
	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func detectColor(out io.Writer, mode ColorMode) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		fd, ok := internal.GetWriterFd(out)
		if !ok || !term.IsTerminal(fd) {
			return false
		}

		return internal.DetectColorLevel(fd) != internal.ModeNone
	default:
		panic("minimal: invalid color mode")
	}
}
