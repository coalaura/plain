package plain

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	RFC3339Local = "2006-01-02T15:04:05"

	Reset = "\x1b[0m"

	Dimmed    = "\x1b[90m"
	Success   = "\x1b[32m"
	Highlight = "\x1b[94m"
	Input     = "\x1b[36m"

	Text  = "\x1b[37m"
	Warn  = "\x1b[33m"
	Error = "\x1b[31m"
)

type Plain struct {
	out    io.Writer
	color  bool
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

	p.color = isWriterTerminal(p.out)

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
		buf.WriteString(Dimmed)
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
