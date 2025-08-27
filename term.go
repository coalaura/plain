package plain

import (
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
)

type terminal struct {
	console.Console
	*os.File
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

func (t *terminal) Close() error {
	t.ShowCursor()

	err := t.Console.Reset()
	if err != nil {
		return err
	}

	return t.File.Close()
}
