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
	AnsiError   = "\033[31m"
	AnsiReset   = internal.AnsiReset

	prefixSub     = "   -> "
	prefixInfo    = ":: "
	prefixSuccess = ":: "
	prefixError   = "!! "
)

type Minimal struct {
	mx    sync.Mutex
	out   io.Writer
	color bool
}

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

func NewMinimal(opts ...option) *Minimal {
	m := &Minimal{
		out: os.Stdout,
	}

	for _, opt := range opts {
		opt(m)
	}

	fd, ok := internal.GetWriterFd(m.out)

	if ok && term.IsTerminal(fd) {
		m.color = internal.DetectColorLevel(fd) != internal.ModeNone
	}

	return m
}

func (m *Minimal) writeArgs(code, prefix string, a ...any) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	if m.color {
		buf = append(buf, code...)
	}

	buf = append(buf, prefix...)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if m.color {
		buf = append(buf, AnsiReset...)
	}

	if !m.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	m.mx.Lock()
	m.out.Write(buf)
	m.mx.Unlock()

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}

func (m *Minimal) writeArgsLine(code, prefix string, a ...any) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	if m.color {
		buf = append(buf, code...)
	}

	buf = append(buf, prefix...)

	if len(a) > 0 {
		buf = fmt.Append(buf, a...)
	}

	if m.color {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, '\n')

	if !m.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	m.mx.Lock()
	m.out.Write(buf)
	m.mx.Unlock()

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (m *Minimal) writeFormat(code, prefix string, format string, a ...any) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	if m.color {
		buf = append(buf, code...)
	}

	buf = append(buf, prefix...)

	if len(a) == 0 {
		buf = append(buf, format...)
	} else {
		buf = fmt.Appendf(buf, format, a...)
	}

	if m.color {
		buf = append(buf, AnsiReset...)
	}

	if !m.color && bytes.IndexByte(buf, '\x1b') >= 0 {
		buf = internal.StripANSI(buf)
	}

	m.mx.Lock()
	m.out.Write(buf)
	m.mx.Unlock()

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}
}
