//go:build windows
// +build windows

package plain

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/containerd/console"
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

func openTTY() (*terminal, error) {
	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0644)
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
	}

	hex := fmt.Sprintf("%x", buf[:num])

	switch hex {
	case "1b4f41":
		return arrowUp, nil
	case "1b4f42":
		return arrowDown, nil
	case "1b4f43":
		return arrowRight, nil
	case "1b4f44":
		return arrowLeft, nil
	case "1b0d":
		return enter, nil
	}

	if num == 1 && buf[0] == 13 {
		return enter, nil
	}

	if num == 1 && buf[0] == 3 {
		t.Close()

		os.Exit(0)
	}

	return invalidInput, nil
}

func (t *terminal) HideCursor() {
	handle := syscall.Handle(os.Stdout.Fd())

	var cci _consoleCursorInfo

	procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))

	cci.visible = 0

	procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
}

func (t *terminal) ShowCursor() {
	handle := syscall.Handle(os.Stdout.Fd())

	var cci _consoleCursorInfo

	procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))

	cci.visible = 1

	procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
}
