package plain

import (
	"io"
	"os"
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
	file    *os.File
	restore func()
}

type fdGetter interface {
	Fd() uintptr
}

type closer interface {
	Close() error
}

func readArrow(p *Plain) (int, error) {
	t, err := openTTY(true)
	if err != nil {
		return 0, err
	}

	p.closer.Store(func() {
		t.Close()
	})

	defer p.closer.RunAndClear()

	return t.ReadArrow()
}

func readKey(p *Plain) (rune, error) {
	t, err := openTTY(false)
	if err != nil {
		return 0, err
	}

	p.closer.Store(func() {
		t.Close()
	})

	defer p.closer.RunAndClear()

	return t.ReadKey()
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

func (t *terminal) Close() error {
	t.ShowCursor()

	if t.restore != nil {
		t.restore()
	}

	return t.file.Close()
}
