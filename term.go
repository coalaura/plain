package plain

import (
	"io"
	"os"

	"github.com/containerd/console"
	"golang.org/x/term"
)

const (
	invalidInput = iota
	arrowUp
	arrowDown
	arrowLeft
	arrowRight
	enter
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

func isWriterTerminal(writer io.Writer) bool {
	if f, ok := writer.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}

	if f, ok := writer.(fdGetter); ok {
		return term.IsTerminal(int(f.Fd()))
	}

	return false
}

func (t *terminal) Close() error {
	t.ShowCursor()

	err := t.Console.Reset()
	if err != nil {
		return err
	}

	return t.File.Close()
}
