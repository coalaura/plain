package plain

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
)

const (
	invalidInput = iota
	arrowUp
	arrowDown
	arrowLeft
	arrowRight
	enter

	ModeNone = iota - 1
	ModeSome
	Mode8Bit
	ModeFull
)

type terminal struct {
	closed  uint32
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

var ErrInterrupted = errors.New("interrupted")

func readWithInterrupt[T any](p *Plain, virtual bool, readFn func(*terminal) (T, error)) (T, error) {
	var zero T

	t, err := openTTY(virtual)
	if err != nil {
		return zero, err
	}

	defer t.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	resCh := make(chan result[T], 1)

	go func() {
		val, err := readFn(t)

		resCh <- result[T]{val, err}
	}()

	select {
	case <-ctx.Done():
		if p.color {
			p.out.Write([]byte(Reset))
		}

		return zero, ErrInterrupted
	case res := <-resCh:
		return res.val, res.err
	}
}

func readArrow(p *Plain) (int, error) {
	return readWithInterrupt(p, true, func(t *terminal) (int, error) {
		return t.ReadArrow()
	})
}

func readKey(p *Plain) (rune, error) {
	return readWithInterrupt(p, false, func(t *terminal) (rune, error) {
		return t.ReadKey()
	})
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

func (t *terminal) Close() {
	if !atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
		return
	}

	t.ShowCursor()

	if t.restore != nil {
		t.restore()
	}
}
