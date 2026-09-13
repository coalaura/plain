//go:build windows

package internal

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type _consoleCursorInfo struct {
	size    uint32
	visible int32
}

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleCursorInfo = kernel32.NewProc("GetConsoleCursorInfo")
	procSetConsoleCursorInfo = kernel32.NewProc("SetConsoleCursorInfo")
)

const (
	enableVirtualTerminalProcessing = 0x0004
)

func DetectColorLevel(fd int) int {
	if os.Getenv("NO_COLOR") != "" {
		return ModeNone
	}

	if os.Getenv("WT_SESSION") != "" || os.Getenv("COLORTERM") == "truecolor" {
		return ModeFull
	}

	termVal := os.Getenv("TERM")
	if strings.Contains(termVal, "256") {
		return Mode8Bit
	}

	if termVal == "xterm" || termVal == "cygwin" {
		return ModeSome
	}

	handle := syscall.Handle(fd)

	var mode uint32

	err := syscall.GetConsoleMode(handle, &mode)
	if err != nil {
		return ModeNone
	}

	if mode&enableVirtualTerminalProcessing == 0 {
		return ModeNone
	}

	return ModeSome
}

func OpenTTY(virtual bool) (*Terminal, error) {
	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	handle := windows.Handle(f.Fd())

	var mode uint32

	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		f.Close()

		return nil, err
	}

	oldMode := mode

	mode &^= windows.ENABLE_LINE_INPUT | windows.ENABLE_ECHO_INPUT

	if virtual {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
		mode |= windows.ENABLE_PROCESSED_INPUT
	}

	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		f.Close()

		return nil, err
	}

	return &Terminal{
		file: f,
		restore: func() {
			windows.SetConsoleMode(handle, oldMode)
		},
	}, nil
}

func (t *Terminal) ReadKey() (rune, error) {
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

func (t *Terminal) ReadArrow() (int, error) {
	t.HideCursor()

	var buf [256]byte

	num, err := t.file.Read(buf[:])
	if err != nil {
		return InvalidInput, err
	}

	if num == 1 {
		switch buf[0] {
		case 'w':
			return ArrowUp, nil
		case 's':
			return ArrowDown, nil
		case 'd':
			return ArrowRight, nil
		case 'a':
			return ArrowLeft, nil
		case '\r', '\n':
			return Enter, nil
		case '\x1b':
			return Cancel, nil
		}
	}

	if num >= 3 && buf[0] == '\x1b' {
		switch buf[1] {
		case '[':
			switch buf[2] {
			case 'A':
				return ArrowUp, nil
			case 'B':
				return ArrowDown, nil
			case 'C':
				return ArrowRight, nil
			case 'D':
				return ArrowLeft, nil
			}
		case 'O':
			switch buf[2] {
			case 'A':
				return ArrowUp, nil
			case 'B':
				return ArrowDown, nil
			case 'C':
				return ArrowRight, nil
			case 'D':
				return ArrowLeft, nil
			}
		}
	}

	if num == 2 && buf[0] == '\x1b' && buf[1] == '\r' {
		return Enter, nil
	}

	return InvalidInput, nil
}

func (t *Terminal) HideCursor() {
	handle := syscall.Handle(os.Stdout.Fd())

	var cci _consoleCursorInfo

	procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))

	cci.visible = 0

	procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
}

func (t *Terminal) ShowCursor() {
	handle := syscall.Handle(os.Stdout.Fd())

	var cci _consoleCursorInfo

	procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))

	cci.visible = 1

	procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
}

func isBackspace(value byte) bool {
	return value == '\b'
}

func isWordBackspace(value byte) bool {
	return value == '\x17' || value == '\x7f'
}
