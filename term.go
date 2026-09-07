package plain

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

const (
	invalidInput = iota
	arrowUp
	arrowDown
	arrowLeft
	arrowRight
	enter
	cancel

	ModeNone = iota - 1
	ModeSome
	Mode8Bit
	ModeFull
)

type terminal struct {
	closed  atomic.Uint32
	file    *os.File
	restore func()
}

type fdGetter interface {
	Fd() uintptr
}

type result[T any] struct {
	val T
	err error
}

func readCtx[T any](ctx context.Context, p *Plain, t *terminal, readFn func(*terminal) (T, error)) (T, error) {
	var zero T

	resCh := make(chan result[T], 1)

	go func() {
		val, err := readFn(t)

		resCh <- result[T]{val, err}
	}()

	select {
	case <-ctx.Done():
		return zero, ErrInterrupted
	case res := <-resCh:
		return res.val, res.err
	}
}

func getWriterFd(writer io.Writer) (int, bool) {
	if f, ok := writer.(*os.File); ok {
		return int(f.Fd()), true
	}

	if f, ok := writer.(fdGetter); ok {
		return int(f.Fd()), true
	}

	return 0, false
}

func (t *terminal) ReadLine() ([]byte, error) {
	t.HideCursor()

	var (
		buf []byte
		one [1]byte
	)

	for {
		n, err := t.file.Read(one[:])
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		b := one[0]

		switch {
		case b == '\n' || b == '\r':
			return buf, nil
		case isWordBackspace(b):
			buf, _ = deleteLastWord(buf)
		case isBackspace(b):
			buf, _ = deleteLastRune(buf)
		case b < 0x20:
			continue
		default:
			buf = append(buf, b)
		}
	}
}

func (t *terminal) ReadVisible(out io.Writer, buf []byte, max int) ([]byte, error) {
	var one [1]byte

	for {
		n, err := t.file.Read(one[:])
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		b := one[0]

		switch {
		case b == '\n' || b == '\r':
			return buf, nil
		case isWordBackspace(b):
			previous := buf

			var removed int

			buf, removed = deleteLastWord(buf)

			eraseVisibleInput(out, previous[len(previous)-removed:])
		case isBackspace(b):
			previous := buf

			var removed int

			buf, removed = deleteLastRune(buf)

			eraseVisibleInput(out, previous[len(previous)-removed:])
		case b < 0x20:
			continue
		case len(buf) < max:
			buf = append(buf, b)
			_, _ = out.Write(one[:])
		}
	}
}

func (t *terminal) ReadMasked(out io.Writer, mask rune) ([]byte, error) {
	t.HideCursor()

	var (
		buf   []byte
		one   [1]byte
		maskB [4]byte
	)

	for {
		n, err := t.file.Read(one[:])
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		b := one[0]

		switch {
		case b == '\n' || b == '\r':
			return buf, nil
		case isWordBackspace(b):
			var removed int

			buf, removed = deleteLastWord(buf)

			eraseMaskedInput(out, mask, removed)
		case isBackspace(b):
			var removed int

			buf, removed = deleteLastRune(buf)

			eraseMaskedInput(out, mask, removed)
		case b < 0x20:
			continue
		default:
			buf = append(buf, b)

			if mask != 0 {
				sz := utf8.EncodeRune(maskB[:], mask)

				out.Write(maskB[:sz])
			}
		}
	}
}

func (t *terminal) Close() {
	if !t.closed.CompareAndSwap(0, 1) {
		return
	}

	t.ShowCursor()

	if t.restore != nil {
		t.restore()
	}
}

func deleteLastRune(buf []byte) ([]byte, int) {
	if len(buf) == 0 {
		return buf, 0
	}

	_, size := utf8.DecodeLastRune(buf)

	return buf[:len(buf)-size], size
}

func deleteLastWord(buf []byte) ([]byte, int) {
	end := len(buf)

	for len(buf) > 0 {
		value, size := utf8.DecodeLastRune(buf)
		if !unicode.IsSpace(value) {
			break
		}

		buf = buf[:len(buf)-size]
	}

	for len(buf) > 0 {
		value, size := utf8.DecodeLastRune(buf)
		if unicode.IsSpace(value) {
			break
		}

		buf = buf[:len(buf)-size]
	}

	return buf, end - len(buf)
}

func eraseMaskedInput(out io.Writer, mask rune, count int) {
	if mask == 0 {
		return
	}

	for range count {
		io.WriteString(out, "\b \b")
	}
}

func eraseVisibleInput(out io.Writer, value []byte) {
	for range utf8.RuneCount(value) {
		_, _ = io.WriteString(out, "\b \b")
	}
}
