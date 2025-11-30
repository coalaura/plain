package plain

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Read displays a prompt aligned with the logger's format.
func (p *Plain) Read(in io.Reader, prompt string, max int) (string, error) {
	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	if p.color {
		buf = append(buf, p.theme.Input...)
	}

	p.out.Write(buf)

	res := make([]byte, max)

	n, err := in.Read(res)

	if cap(buf) < 4096 {
		*bp = buf

		pool.Put(bp)
	}

	if err != nil {
		return "", err
	}

	p.out.Write([]byte(Reset))

	return strings.TrimSpace(string(res[:n])), nil
}

// Select displays a cyclic selector aligned with the logger's format.
func (p *Plain) Select(prompt string, options []string) (int, error) {
	var (
		index  int
		length int
	)

	bp := pool.Get().(*[]byte)
	defer pool.Put(bp)

	for {
		buf := *bp
		buf = buf[:0]

		label := options[index]

		buf = append(buf, '\r')

		buf = p.appendPadding(buf)

		if p.color {
			buf = append(buf, Reset...)
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

		p.out.Write(buf)

		i, err := readArrow()
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
