package plain

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coalaura/atom"
	"github.com/coalaura/plain/ansi"
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

const AnsiReset = ansi.AnsiReset

// Theme defines the ANSI color sequences used by the logger
type Theme struct {
	Success   string
	Highlight string
	Input     string

	Dimmed string
	Warn   string
	Error  string
}

type outputState struct {
	out   io.Writer
	color bool
	theme Theme
}

type outputWriter struct {
	plain  *Plain
	target io.Writer
}

// Plain is a small, allocation-conscious logger with optional ANSI color output
type Plain struct {
	out io.Writer

	writeLock sync.Mutex
	readLock  sync.Mutex

	output atomic.Pointer[outputState]

	level  atom.Value[Level]
	format atom.Value[string]

	readBuf []byte
}

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

func (w outputWriter) Write(buf []byte) (int, error) {
	return w.plain.writeOutput(w.target, buf)
}

// New creates a Plain logger configured by the provided options
func New(opts ...option) *Plain {
	p := &Plain{
		out: os.Stdout,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.setTarget(p.out)

	return p
}

// Theme returns a theme color as ansi code
func (p *Plain) Theme(c themeColor) string {
	state := p.outputState()

	switch c {
	case Dimmed:
		return state.theme.Dimmed
	case Success:
		return state.theme.Success
	case Highlight:
		return state.theme.Highlight
	case Input:
		return state.theme.Input
	case Warn:
		return state.theme.Warn
	case Error:
		return state.theme.Error
	}

	if !state.color {
		return ""
	}

	return AnsiReset
}

// WaitForInterrupt blocks until SIGINT or SIGTERM is received
func (p *Plain) WaitForInterrupt() {
	internal.WaitForInterrupt()
}

// OnSignal registers a non-blocking handler that executes the provided callback
// every time the specified signal is received.
func (p *Plain) OnSignal(sig os.Signal, handler func()) {
	p.OnSignalContext(context.Background(), sig, handler)
}

// OnSignalContext registers a non-blocking handler until ctx is cancelled.
func (p *Plain) OnSignalContext(ctx context.Context, sig os.Signal, handler func()) {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, sig)

	go func() {
		defer signal.Stop(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				handler()
			}
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
	state := p.outputState()

	if !state.color && p.format.Load() == "" && strings.IndexByte(msg, '\x1b') == -1 {
		if sw, ok := state.out.(io.StringWriter); ok {
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
		buf = p.appendHeader(buf, code, state.color, state.theme)
	}

	buf = append(buf, msg...)

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeLine(code, msg string, reset, noHeader bool) {
	state := p.outputState()

	if !state.color && p.format.Load() == "" && strings.IndexByte(msg, '\x1b') == -1 {
		if sw, ok := state.out.(io.StringWriter); ok {
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
		buf = p.appendHeader(buf, code, state.color, state.theme)
	}

	buf = append(buf, msg...)

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, '\n')

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

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

	state := p.outputState()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code, state.color, state.theme)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeArgsLine(code string, reset bool, a ...any) {
	state := p.outputState()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code, state.color, state.theme)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, '\n')

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

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

	state := p.outputState()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code, state.color, state.theme)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (p *Plain) writeFormatLine(code string, reset bool, format string, a ...any) {
	state := p.outputState()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code, state.color, state.theme)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	if state.color && reset {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, '\n')

	if !state.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = ansi.StripANSI(buf)
	}

	p.writeBytes(state.out, buf)

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (p *Plain) appendHeader(dst []byte, code string, colored bool, theme Theme) []byte {
	format := p.format.Load()

	if format == "" {
		if colored {
			dst = append(dst, code...)
		}

		return dst
	}

	if colored {
		dst = append(dst, theme.Dimmed...)
	}

	dst = time.Now().AppendFormat(dst, format)

	if colored {
		if code != "" {
			dst = append(dst, code...)
		} else {
			dst = append(dst, AnsiReset...)
		}
	}

	dst = append(dst, ' ')

	return dst
}

func (p *Plain) setTarget(out io.Writer) {
	color, _, theme := detectWriterTheme(out)

	p.output.Store(&outputState{
		out:   out,
		color: color,
		theme: theme,
	})
}

func (p *Plain) outputState() outputState {
	return *p.output.Load()
}

func (p *Plain) writeBytes(out io.Writer, buf []byte) {
	_, _ = p.writeOutput(out, buf)
}

func (p *Plain) writeOutput(out io.Writer, buf []byte) (int, error) {
	p.writeLock.Lock()
	n, err := out.Write(buf)
	p.writeLock.Unlock()

	return n, err
}

func detectWriterTheme(out io.Writer) (bool, int, Theme) {
	fd, ok := internal.GetWriterFd(out)
	if !ok || !term.IsTerminal(fd) {
		return false, internal.ModeNone, Theme{}
	}

	mode := internal.DetectColorLevel(fd)
	if mode == internal.ModeNone {
		return false, mode, Theme{}
	}

	theme := Theme{
		Dimmed:    color(mode, ansi.AnsiHiBlack, c256(244), rgb(145, 145, 145)),
		Success:   color(mode, ansi.AnsiGreen, c256(114), rgb(120, 210, 130)),
		Highlight: color(mode, ansi.AnsiHiBlue, c256(111), rgb(100, 180, 255)),
		Input:     color(mode, ansi.AnsiCyan, c256(152), rgb(130, 220, 220)),
		Warn:      color(mode, ansi.AnsiYellow, c256(215), rgb(255, 190, 80)),
		Error:     color(mode, ansi.AnsiRed, c256(210), rgb(255, 110, 110)),
	}

	return true, mode, theme
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
