//go:build linux
// +build linux

package plain

import (
	"fmt"
	"os"
	"strings"

	"github.com/containerd/console"
)

func detectColorLevel(_ int) int {
	if os.Getenv("NO_COLOR") != "" {
		return ModeNone
	}

	termVal := os.Getenv("TERM")
	if termVal == "dumb" {
		return ModeNone
	}

	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return ModeFull
	}

	if strings.Contains(termVal, "256") {
		return Mode8Bit
	}

	return ModeSome
}

func openTTY() (*terminal, error) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	c := console.Current()

	err = c.SetRaw()
	if err != nil {
		return nil, err
	}

	return &terminal{
		Console: c,
		File:    f,
	}, nil
}

func (t *terminal) ReadArrow() (int, error) {
	t.HideCursor()

	var buf [256]byte

	num, err := t.File.Read(buf[:])
	if err != nil {
		return invalidInput, err
	}

	text := string(buf[:num])

	switch text {
	case "\x1b[A":
		return arrowUp, nil
	case "\x1b[B":
		return arrowDown, nil
	case "\x1b[C":
		return arrowRight, nil
	case "\x1b[D":
		return arrowLeft, nil
	case "\r", "\n":
		return enter, nil
	}

	if num == 1 && buf[0] == 3 {
		t.Close()

		os.Exit(0)
	}

	return invalidInput, nil
}

func (t *terminal) HideCursor() {
	fmt.Fprint(os.Stdout, "\x1b[?25l")
}

func (t *terminal) ShowCursor() {
	fmt.Fprint(os.Stdout, "\x1b[?25h")
}
