package plain

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/coalaura/atom"
	"github.com/coalaura/plain/internal"
	"golang.org/x/term"
)

// RFC3339Local is an RFC3339-like time format without timezone information
const RFC3339Local = "2006-01-02T15:04:05"

type Level uint8

const (
	LevelDebug Level = iota
	LevelPrint
	LevelWarn
	LevelError
)

type themeColor uint8

const (
	Dimmed themeColor = iota
	Success
	Highlight
	Input
	Warn
	Error
	Reset
)

// Theme defines the ANSI color sequences used by the logger
type Theme struct {
	Success   string
	Highlight string
	Input     string

	Dimmed string
	Warn   string
	Error  string
}

// Plain is a small, allocation-conscious logger with optional ANSI color output
type Plain struct {
	out  io.Writer
	term *internal.Terminal

	writeLock sync.Mutex
	readLock  sync.Mutex

	color bool
	mode  int

	level  atom.Value[Level]
	format atom.Value[string]
	theme  Theme

	readBuf []byte
}

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

// New creates a Plain logger configured by the provided options
func New(opts ...option) *Plain {
	p := &Plain{
		out: os.Stdout,
	}

	for _, opt := range opts {
		opt(p)
	}

	fd, ok := internal.GetWriterFd(p.out)

	if ok && term.IsTerminal(fd) {
		p.mode = internal.DetectColorLevel(fd)
		p.color = p.mode > internal.ModeNone

		p.theme.Dimmed = color(p.mode, "\x1b[90m", c256(244), rgb(145, 145, 145))
		p.theme.Success = color(p.mode, "\x1b[32m", c256(114), rgb(120, 210, 130))
		p.theme.Highlight = color(p.mode, "\x1b[94m", c256(111), rgb(100, 180, 255))
		p.theme.Input = color(p.mode, "\x1b[36m", c256(152), rgb(130, 220, 220))
		p.theme.Warn = color(p.mode, "\x1b[33m", c256(215), rgb(255, 190, 80))
		p.theme.Error = color(p.mode, "\x1b[31m", c256(210), rgb(255, 110, 110))
	} else {
		p.mode = internal.ModeNone
	}

	return p
}

// Theme returns a theme color as ansi code
func (p *Plain) Theme(c themeColor) string {
	switch c {
	case Dimmed:
		return p.theme.Dimmed
	case Success:
		return p.theme.Success
	case Highlight:
		return p.theme.Highlight
	case Input:
		return p.theme.Input
	case Warn:
		return p.theme.Warn
	case Error:
		return p.theme.Error
	}

	if !p.color {
		return ""
	}

	return internal.AnsiReset
}

// WaitForInterrupt blocks until SIGINT or SIGTERM is received
func (p *Plain) WaitForInterrupt() {
	internal.WaitForInterrupt()
}

// OnSignal registers a non-blocking handler that executes the provided callback
// every time the specified signal is received.
func (p *Plain) OnSignal(sig os.Signal, handler func()) {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, sig)

	go func() {
		for range ch {
			handler()
		}
	}()
}

// Write writes a formatted log line with an optional reset code
func (p *Plain) Write(code, msg string, reset, newline, noHeader bool) {
	if newline {
		p.writeLine(code, msg, reset, noHeader)

		return
	}

	p.writeString(code, msg, reset, noHeader)
}

// Writeln writes a formatted log line with an optional reset code and a newline
func (p *Plain) Writeln(code, msg string, reset, noHeader bool) {
	p.Write(code, msg, reset, true, noHeader)
}

func (p *Plain) writeString(code, msg string, reset, noHeader bool) {
	if !p.color && p.format.Load() == "" && strings.IndexByte(msg, '\x1b') == -1 {
		if sw, ok := p.out.(io.StringWriter); ok {
			p.writeLock.Lock()
			sw.WriteString(msg)
			p.writeLock.Unlock()

			return
		}
	}

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	if noHeader {
		buf = append(buf, code...)
	} else {
		buf = p.appendHeader(buf, code)
	}

	buf = append(buf, msg...)

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeLine(code, msg string, reset, noHeader bool) {
	if !p.color && p.format.Load() == "" && strings.IndexByte(msg, '\x1b') == -1 {
		if sw, ok := p.out.(io.StringWriter); ok {
			p.writeLock.Lock()
			sw.WriteString(msg)
			sw.WriteString("\n")
			p.writeLock.Unlock()

			return
		}
	}

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	if noHeader {
		buf = append(buf, code...)
	} else {
		buf = p.appendHeader(buf, code)
	}

	buf = append(buf, msg...)

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (p *Plain) writeArgs(code string, reset, nl bool, a ...any) {
	if nl {
		p.writeArgsLine(code, reset, a...)

		return
	}

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeArgsLine(code string, reset bool, a ...any) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (p *Plain) writeFormat(code string, reset, nl bool, format string, a ...any) {
	if nl {
		p.writeFormatLine(code, reset, format, a...)

		return
	}

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeFormatLine(code string, reset bool, format string, a ...any) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	if p.color && reset {
		buf = append(buf, internal.AnsiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	p.writeLock.Lock()
	p.out.Write(buf)
	p.writeLock.Unlock()

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (p *Plain) appendHeader(dst []byte, code string) []byte {
	format := p.format.Load()

	if format == "" {
		if p.color {
			dst = append(dst, code...)
		}

		return dst
	}

	if p.color {
		dst = append(dst, p.theme.Dimmed...)
	}

	dst = time.Now().AppendFormat(dst, format)

	if p.color {
		if code != "" {
			dst = append(dst, code...)
		} else {
			dst = append(dst, internal.AnsiReset...)
		}
	}

	dst = append(dst, ' ')

	return dst
}

func color(mode int, some, bit8, full string) string {
	switch mode {
	case internal.ModeSome:
		return some
	case internal.Mode8Bit:
		return bit8
	case internal.ModeFull:
		return full
	}

	return ""
}

func c256(id int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", id)
}

func rgb(r, g, b int) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
