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
	"time"

	"golang.org/x/term"
)

const (
	// RFC3339Local is an RFC3339-like time format without timezone information
	RFC3339Local = "2006-01-02T15:04:05"

	ansiReset = "\x1b[0m"

	Dimmed themeColor = iota
	Success
	Highlight
	Input
	Warn
	Error
	Reset
)

type themeColor int

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
	term *terminal

	writeLock sync.Mutex
	readLock  sync.Mutex

	color bool
	mode  int

	theme  Theme
	format string

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

	fd, ok := getWriterFd(p.out)

	if ok && term.IsTerminal(fd) {
		p.mode = detectColorLevel(fd)
		p.color = p.mode > ModeNone

		p.theme.Dimmed = color(p.mode, "\x1b[90m", c256(244), rgb(145, 145, 145))
		p.theme.Success = color(p.mode, "\x1b[32m", c256(114), rgb(120, 210, 130))
		p.theme.Highlight = color(p.mode, "\x1b[94m", c256(111), rgb(100, 180, 255))
		p.theme.Input = color(p.mode, "\x1b[36m", c256(152), rgb(130, 220, 220))
		p.theme.Warn = color(p.mode, "\x1b[33m", c256(215), rgb(255, 190, 80))
		p.theme.Error = color(p.mode, "\x1b[31m", c256(210), rgb(255, 110, 110))
	} else {
		p.mode = ModeNone
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

	return ansiReset
}

// WaitForInterrupt blocks until SIGINT or SIGTERM is received
func (p *Plain) WaitForInterrupt() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	<-ctx.Done()

	return nil
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
	if !p.color && p.format == "" && strings.IndexByte(msg, '\x1b') == -1 {
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
		buf = append(buf, ansiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
	if !p.color && p.format == "" && strings.IndexByte(msg, '\x1b') == -1 {
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
		buf = append(buf, ansiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
		buf = append(buf, ansiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
		buf = append(buf, ansiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
		buf = append(buf, ansiReset...)
	}

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
		buf = append(buf, ansiReset...)
	}

	buf = append(buf, '\n')

	if !p.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = p.stripANSI(buf)
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
	if p.format == "" {
		if p.color {
			dst = append(dst, code...)
		}

		return dst
	}

	if p.color {
		dst = append(dst, p.theme.Dimmed...)
	}

	dst = time.Now().AppendFormat(dst, p.format)

	if p.color {
		if code != "" {
			dst = append(dst, code...)
		} else {
			dst = append(dst, ansiReset...)
		}
	}

	dst = append(dst, ' ')

	return dst
}

func (p *Plain) stripANSI(buf []byte) []byte {
	var j int

	for i := 0; i < len(buf); {
		// ESC sequence?
		if buf[i] == '\x1b' && i+1 < len(buf) {
			switch buf[i+1] {
			case '[':
				// CSI: ESC [ ... final_byte (0x40-0x7E)
				i += 2

				for i < len(buf) && (buf[i] < 0x40 || buf[i] > 0x7E) {
					i++
				}

				if i < len(buf) {
					i++ // skip the final byte too
				}

				continue
			case ']':
				// OSC: ESC ] ... BEL (0x07) or ST (ESC \)
				i += 2

				for i < len(buf) {
					if buf[i] == '\x07' {
						i++

						break
					}

					if buf[i] == '\x1b' && i+1 < len(buf) && buf[i+1] == '\\' {
						i += 2

						break
					}

					i++
				}

				continue
			default:
				// Other 2-byte escapes (ESC ( etc.)
				i += 2

				continue
			}
		}

		buf[j] = buf[i]

		j++
		i++
	}

	return buf[:j]
}

func color(mode int, some, bit8, full string) string {
	switch mode {
	case ModeSome:
		return some
	case Mode8Bit:
		return bit8
	case ModeFull:
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
