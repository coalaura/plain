package plain

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// Mask constants for ReadMask. Pass any rune to ReadMask; these are convenience
// values for common masking characters. MaskNone disables on-screen echoing
// entirely (behaviour equivalent to ReadHidden, but with backspace handling).
const (
	MaskStar = '*'
	MaskDot  = '-'
	MaskHash = '#'
)

// SelectOption provides the label and description for SelectWithDescription.
type SelectOption interface {
	Label() string
	Description() string
}

// Read displays a prompt aligned with the logger's format and reads max bytes from stdin.
func (p *Plain) Read(prompt string, max int) (string, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	if max < 0 {
		panic("plain: negative maximum input length")
	}

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	defer func() {
		if cap(buf) <= 4096 {
			*bp = buf

			pool.Put(bp)
		}
	}()

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	if p.color {
		buf = append(buf, p.theme.Input...)

		defer io.WriteString(p.out, ansiReset)
	}

	p.out.Write(buf)

	res := p.readBuf
	if cap(res) < max {
		res = make([]byte, max)

		p.readBuf = res
	}

	res = res[:max]

	n, err := os.Stdin.Read(res)
	if err != nil {
		return "", err
	}

	end := n

	for end > 0 {
		b := res[end-1]
		if b != '\n' && b != '\r' {
			break
		}

		end--
	}

	return string(res[:end]), nil
}

// ReadHidden displays a prompt aligned with the logger's format, reads a line from stdin without echoing input and prints a newline.
func (p *Plain) ReadHidden(prompt string) (string, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)
	buf := *bp
	buf = buf[:0]

	defer func() {
		if cap(buf) <= 4096 {
			*bp = buf

			pool.Put(bp)
		}
	}()

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	if p.color {
		buf = append(buf, p.theme.Input...)

		defer io.WriteString(p.out, ansiReset)
	}

	p.out.Write(buf)

	term, err := openTTY(false)
	if err != nil {
		return "", err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	line, err := readCtx(ctx, p, term, (*terminal).ReadLine)
	if err != nil {
		io.WriteString(p.out, "\n")

		return "", err
	}

	io.WriteString(p.out, "\n")

	return string(line), nil
}

// ReadMask displays a prompt aligned with the logger's format, reads a line from stdin masking each typed byte on screen, and prints a newline.
func (p *Plain) ReadMask(prompt string, mask rune) (string, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)
	buf := *bp
	buf = buf[:0]

	defer func() {
		if cap(buf) <= 4096 {
			*bp = buf

			pool.Put(bp)
		}
	}()

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	if p.color {
		buf = append(buf, p.theme.Input...)

		defer io.WriteString(p.out, ansiReset)
	}

	p.out.Write(buf)

	term, err := openTTY(false)
	if err != nil {
		return "", err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	line, err := readCtx(ctx, p, term, func(t *terminal) ([]byte, error) {
		return t.ReadMasked(p.out, mask)
	})

	if err != nil {
		io.WriteString(p.out, "\n")

		return "", err
	}

	io.WriteString(p.out, "\n")

	return string(line), nil
}

// ReadOne displays a prompt aligned with the logger's format and reads 1 byte from stdin.
func (p *Plain) ReadOne(prompt string, echo bool) (rune, error) {
	p.readLock.Lock()
	defer p.readLock.Unlock()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	defer func() {
		*bp = buf

		pool.Put(bp)
	}()

	buf = p.appendPadding(buf)

	buf = append(buf, prompt...)

	p.out.Write(buf)

	defer io.WriteString(p.out, "\n")

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

	buf := *bp
	buf = buf[:0]

	defer func() {
		*bp = buf

		pool.Put(bp)
	}()

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
			io.WriteString(p.out, "\n")

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

	io.WriteString(p.out, "\n")

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
	return p.selectOption(prompt, len(options), false, func(index int) (string, string) {
		return options[index], ""
	})
}

// SelectWithDescription displays a cyclic selector with the selected option's description on the line below it.
func (p *Plain) SelectWithDescription(prompt string, options []SelectOption) (int, error) {
	return p.selectOption(prompt, len(options), true, func(index int) (string, string) {
		option := options[index]

		return option.Label(), option.Description()
	})
}

func (p *Plain) selectOption(prompt string, optionCount int, showDescription bool, optionAt func(int) (string, string)) (int, error) {
	if optionCount == 0 {
		return 0, ErrNoOptions
	}

	p.readLock.Lock()
	defer p.readLock.Unlock()

	index := 0

	bp := pool.Get().(*[]byte)

	buf := *bp

	defer func() {
		*bp = buf

		pool.Put(bp)
	}()

	term, err := openTTY(true)
	if err != nil {
		return 0, err
	}

	defer term.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	input := make(chan result[int], 1)

	go func() {
		for {
			value, err := term.ReadArrow()

			readResult := result[int]{value, err}

			select {
			case input <- readResult:
			case <-ctx.Done():
				return
			}

			if err != nil {
				return
			}
		}
	}()

	resize := time.NewTicker(100 * time.Millisecond)
	defer resize.Stop()

	lastWidth := selectWidth()
	rendered := false

	defer func() {
		if !showDescription || !rendered {
			return
		}

		_, description := optionAt(index)

		buf = buf[:0]
		buf = p.appendSelectDescription(buf, description, false)

		p.out.Write(buf)
	}()

	redraw := true

	for {
		if !redraw {
			select {
			case <-ctx.Done():
				return 0, ErrInterrupted
			case readResult := <-input:
				if readResult.err != nil {
					return 0, readResult.err
				}

				switch readResult.val {
				case arrowRight, arrowDown:
					index++

					if index >= optionCount {
						index = 0
					}
				case arrowLeft, arrowUp:
					index--

					if index < 0 {
						index = optionCount - 1
					}
				case enter:
					if !showDescription {
						p.out.Write([]byte("\n"))
					}

					return index, nil
				case cancel:
					return 0, ErrInterrupted
				}

				redraw = true
			case <-resize.C:
				width := selectWidth()
				if width != lastWidth {
					lastWidth = width
					redraw = true
				}
			}

			continue
		}

		buf = buf[:0]

		label, description := optionAt(index)

		buf = append(buf, "\r\x1b[2K"...)

		lineStart := len(buf)

		buf = p.appendPadding(buf)

		usedWidth := len(buf) - lineStart
		usedWidth += utf8.RuneCountInString(prompt)

		label = truncateSelectText(label, availableSelectWidth(usedWidth))

		if p.color {
			buf = append(buf, ansiReset...)
		}

		buf = append(buf, prompt...)

		if p.color {
			buf = append(buf, p.theme.Input...)
		}

		buf = append(buf, label...)

		if p.color {
			buf = append(buf, ansiReset...)
		}

		if showDescription {
			buf = p.appendSelectDescription(buf, description, true)
		}

		p.out.Write(buf)

		rendered = true
		redraw = false
	}
}

func (p *Plain) appendSelectDescription(dst []byte, description string, returnToSelect bool) []byte {
	dst = append(dst, "\n\r\x1b[J"...)

	lineStart := len(dst)

	dst = p.appendPadding(dst)

	usedWidth := len(dst) - lineStart

	description = truncateSelectText(description, availableSelectWidth(usedWidth))

	if p.color {
		dst = append(dst, p.theme.Dimmed...)
	}

	dst = append(dst, description...)

	if p.color {
		dst = append(dst, ansiReset...)
	}

	if returnToSelect {
		dst = append(dst, "\x1b[1A\r"...)

		return dst
	}

	return append(dst, '\n')
}

func (p *Plain) appendPadding(dst []byte) []byte {
	format := p.format.Load()

	if format == "" {
		return dst
	}

	start := len(dst)

	dst = time.Now().AppendFormat(dst, format)

	for i := start; i < len(dst); i++ {
		dst[i] = ' '
	}

	dst = append(dst, ' ')

	return dst
}

func selectWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return -1
	}

	return width
}

func availableSelectWidth(usedWidth int) int {
	width := selectWidth()
	if width < 0 {
		return -1
	}

	available := width - usedWidth
	if available < 0 {
		return 0
	}

	return available
}

func truncateSelectText(text string, width int) string {
	lineEnd := strings.IndexAny(text, "\r\n")
	if lineEnd >= 0 {
		text = text[:lineEnd]
	}

	if width < 0 {
		return text
	}

	runeCount := utf8.RuneCountInString(text)
	if runeCount <= width {
		return text
	}

	if width <= 3 {
		return strings.Repeat(".", width)
	}

	runes := []rune(text)

	return string(runes[:width-3]) + "..."
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
