package plain

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

// Read displays a prompt aligned with the logger's format and reads max bytes from stdin.
func (p *Plain) Read(prompt string, max int) (string, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	if p.color {
		buf = append(buf, p.theme.Input...)

		defer p.out.Write([]byte(ansiReset))
	}

	p.out.Write(buf)

	res := make([]byte, max)

	n, err := os.Stdin.Read(res)

	if cap(buf) < 4096 {
		pool.Put(bp)
	}

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(res[:n])), nil
}

// ReadOne displays a prompt aligned with the logger's format and reads 1 byte from stdin.
func (p *Plain) ReadOne(prompt string, echo bool) (rune, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)
	defer pool.Put(bp)

	buf := *bp
	buf = buf[:0]

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	p.out.Write(buf)

	defer p.out.Write([]byte("\n"))

	term, err := openTTY(false)
	if err != nil {
		return 0, err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	b, err := readCtx(ctx, p, term, (*terminal).ReadKey)
	if err != nil {
		return 0, err
	}

	if echo {
		buf = buf[:0]

		if p.color {
			buf = append(buf, p.theme.Input...)
		}

		buf = append(buf, byte(b))

		if p.color {
			buf = append(buf, ansiReset...)
		}

		p.out.Write(buf)
	}

	return b, nil
}

func (p *Plain) confirm(prompt string, defaultYes, echo bool, prefix string) (bool, error) {
	suffix := " [y/N]"

	if defaultYes {
		suffix = " [Y/n]"
	}

	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)
	defer pool.Put(bp)

	buf := *bp
	buf = buf[:0]

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)
	buf = append(buf, suffix...)

	p.out.Write(buf)

	term, err := openTTY(false)
	if err != nil {
		return false, err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var (
		done   bool
		result bool
	)

	for !done {
		b, err := readCtx(ctx, p, term, (*terminal).ReadKey)
		if err != nil {
			p.out.Write([]byte("\n"))

			return false, err
		}

		switch b {
		case 'y', 'Y':
			done = true
			result = true
		case 'n', 'N':
			done = true
		case '\r', '\n':
			done = true
			result = defaultYes
		}
	}

	if echo {
		buf = buf[:0]

		if prefix != "" {
			buf = append(buf, prefix...)
		}

		if p.color {
			buf = append(buf, p.theme.Input...)
		}

		if result {
			buf = append(buf, byte('y'))
		} else {
			buf = append(buf, byte('n'))
		}

		if p.color {
			buf = append(buf, ansiReset...)
		}

		p.out.Write(buf)
	}

	p.out.Write([]byte("\n"))

	return result, nil
}

// Confirm displays a prompt aligned with the logger's format, reads y/n confirmation from stdin and prints a newline
func (p *Plain) Confirm(prompt string, defaultYes bool) (bool, error) {
	return p.confirm(prompt, defaultYes, false, "")
}

// ConfirmWithEcho displays a prompt aligned with the logger's format, reads y/n confirmation from stdin and echoes the chosen input (optionally prefixed) before printing a newline.
func (p *Plain) ConfirmWithEcho(prompt string, defaultYes bool, prefix string) (bool, error) {
	return p.confirm(prompt, defaultYes, true, prefix)
}

// Select displays a cyclic selector aligned with the logger's format.
func (p *Plain) Select(prompt string, options []string) (int, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	var (
		index  int
		length int
	)

	bp := pool.Get().(*[]byte)
	defer pool.Put(bp)

	term, err := openTTY(true)
	if err != nil {
		return 0, err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for {
		buf := *bp
		buf = buf[:0]

		label := options[index]

		buf = append(buf, '\r')

		buf = p.appendPadding(buf)

		if p.color {
			buf = append(buf, ansiReset...)
		}

		buf = append(buf, prompt...)

		if p.color {
			buf = append(buf, p.theme.Input...)
		}

		buf = append(buf, label...)

		l := len(label)
		if l < length {
			remaining := length - l

			for range remaining {
				buf = append(buf, ' ')
			}
		}

		length = len(label)

		if p.color {
			buf = append(buf, ansiReset...)
		}

		p.out.Write(buf)

		i, err := readCtx(ctx, p, term, (*terminal).ReadArrow)
		if err != nil {
			return 0, err
		}

		switch i {
		case arrowRight, arrowDown:
			index++

			if index >= len(options) {
				index = 0
			}
		case arrowLeft, arrowUp:
			index--

			if index < 0 {
				index = len(options) - 1
			}
		case enter:
			p.out.Write([]byte("\n"))

			return index, nil
		}
	}
}

func (p *Plain) appendPadding(dst []byte) []byte {
	if p.format == "" {
		return dst
	}

	start := len(dst)

	dst = time.Now().AppendFormat(dst, p.format)

	for i := start; i < len(dst); i++ {
		dst[i] = ' '
	}

	dst = append(dst, ' ')

	return dst
}

// AsStrings converts a slice/array of any type to a slice of strings
func AsStrings[T any](input []T) []string {
	if strs, ok := any(input).([]string); ok {
		return strs
	}

	out := make([]string, len(input))

	for i, value := range input {
		if str, ok := any(value).(string); ok {
			out[i] = str
		} else {
			out[i] = fmt.Sprint(value)
		}
	}

	return out
}
