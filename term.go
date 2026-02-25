package plain

import (
	"context"
	"io"
	"os"
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

func (t *terminal) Close() {
	if !atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
		return
	}

	t.ShowCursor()

	if t.restore != nil {
		t.restore()
	}
}
