package plain

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

const (
	RFC3339Local = "2006-01-02T15:04:05"

	Reset = "\x1b[0m"
)

type Theme struct {
	Success   string
	Highlight string
	Input     string

	Dimmed string
	Warn   string
	Error  string
}

type Plain struct {
	out io.Writer

	color bool
	mode  int

	theme Theme

	format string
}

var pool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

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

		// Dimmed: Steel Gray (Timestamp/Meta)
		p.theme.Dimmed = color(p.mode, "\x1b[90m", c256(244), rgb(145, 145, 145))

		// Success: Soft Mint / Pastel Green
		p.theme.Success = color(p.mode, "\x1b[32m", c256(114), rgb(120, 210, 130))

		// Highlight: Sky Blue
		p.theme.Highlight = color(p.mode, "\x1b[94m", c256(111), rgb(100, 180, 255))

		// Input: Soft Cyan / Teal
		p.theme.Input = color(p.mode, "\x1b[36m", c256(152), rgb(130, 220, 220))

		// Warn: Soft Orange / Amber
		p.theme.Warn = color(p.mode, "\x1b[33m", c256(215), rgb(255, 178, 90))

		// Error: Pastel Red / Coral
		p.theme.Error = color(p.mode, "\x1b[31m", c256(210), rgb(255, 110, 110))
	} else {
		p.mode = ModeNone
	}

	return p
}

func (p *Plain) Write(code, msg string, reset bool) {
	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)

	buf.Reset()

	p.writeHeader(buf, code)

	buf.WriteString(msg)

	if p.color && reset {
		buf.WriteString(Reset)
	}

	p.out.Write(buf.Bytes())
}

func (p *Plain) writeHeader(buf *bytes.Buffer, code string) {
	if p.format == "" {
		if p.color {
			buf.WriteString(code)
		}

		return
	}

	if p.color {
		buf.WriteString(p.theme.Dimmed)
	}

	buf.WriteString(time.Now().Format(p.format))

	if p.color {
		buf.WriteString(code)
	}

	buf.WriteByte(' ')
}

func sprint(a ...any) string {
	if len(a) == 0 {
		return ""
	}

	return fmt.Sprint(a...)
}

func sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return format
	}

	return fmt.Sprintf(format, a...)
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
