package plain

import (
	"fmt"
	"io"
	"os"
	"os/signal"
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

	readLock sync.Mutex
	closer   *runner

	color bool
	mode  int

	theme Theme

	format string
}

var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

func New(opts ...option) *Plain {
	p := &Plain{
		out:    os.Stdout,
		closer: &runner{},
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

func (p *Plain) HandleInterrupt(fn ...func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch

		p.Close()

		for _, f := range fn {
			f()
		}

		os.Exit(1)
	}()
}

func (p *Plain) Write(code, msg string, reset, nl bool) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, code)

	buf = append(buf, msg...)

	if p.color && reset {
		buf = append(buf, Reset...)
	}

	if nl {
		buf = append(buf, '\n')
	}

	p.out.Write(buf)

	if cap(buf) > 4096 {
		return
	}

	*bp = buf

	pool.Put(bp)
}

func (p *Plain) Close() error {
	p.closer.RunAndClear()

	if cl, ok := p.out.(closer); ok {
		err := cl.Close()
		if err != nil {
			return err
		}
	}

	return nil
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
			dst = append(dst, Reset...)
		}
	}

	dst = append(dst, ' ')

	return dst
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
