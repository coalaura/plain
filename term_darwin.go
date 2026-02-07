//go:build darwin || freebsd || netbsd || openbsd
// +build darwin freebsd netbsd openbsd

package plain

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
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

func openTTY(_ bool) (*terminal, error) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	fd := int(f.Fd())

	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		f.Close()

		return nil, err
	}

	oldState := *termios

	termios.Lflag &^= unix.ICANON | unix.ECHO
	termios.Lflag |= unix.ISIG

	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		f.Close()

		return nil, err
	}

	return &terminal{
		file: f,
		restore: func() {
			_ = unix.IoctlSetTermios(fd, unix.TIOCSETA, &oldState)
		},
	}, nil
}

func (t *terminal) ReadKey() (rune, error) {
	t.HideCursor()

	var buf [1]byte

	for {
		n, err := t.file.Read(buf[:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			continue
		}

		b := buf[0]

		switch {
		case b == '\r' || b == '\n':
			return '\n', nil
		case (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9'):
			return rune(b), nil
		}
	}
}

func (t *terminal) ReadArrow() (int, error) {
	t.HideCursor()

	var buf [256]byte

	num, err := t.file.Read(buf[:])
	if err != nil {
		return invalidInput, err
	}

	text := string(buf[:num])

	switch text {
	case "\x1b[A", "w":
		return arrowUp, nil
	case "\x1b[B", "s":
		return arrowDown, nil
	case "\x1b[C", "d":
		return arrowRight, nil
	case "\x1b[D", "a":
		return arrowLeft, nil
	case "\r", "\n":
		return enter, nil
	}

	return invalidInput, nil
}

func (t *terminal) HideCursor() {
	fmt.Fprint(os.Stdout, "\x1b[?25l")
}

func (t *terminal) ShowCursor() {
	fmt.Fprint(os.Stdout, "\x1b[?25h")
}
