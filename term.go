package plain

import (
	"io"
	"os"

	"github.com/containerd/console"
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
	console.Console
	*os.File
}

type fdGetter interface {
	Fd() uintptr
}

func readArrow() (int, error) {
	t, err := openTTY()
	if err != nil {
		return 0, err
	}

	i, err := t.ReadArrow()

	t.Close()

	return i, err
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

	err := t.Console.Reset()
	if err != nil {
		return err
	}

	return t.File.Close()
}
