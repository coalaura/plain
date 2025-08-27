package plain

import (
	"fmt"
	"io"
	"strings"
)

// Read displays a prompt, waits for user input from the provided io.Reader and returns the entered text (trimmed of surrounding whitespace).
func (p *Plain) Read(in io.Reader, prompt string, max int) (string, error) {
	p.Write(Text, prompt, false)
	p.Write(Input, "", false)

	buf := make([]byte, max)

	n, err := in.Read(buf)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(buf[:n])), nil
}

// Select displays a prompt followed by a cyclic selector UI, allowing the user to navigate through a list of options using arrow keys and confirm the selection by pressing Enter.
func (p *Plain) Select(prompt string, options []string) (int, error) {
	var (
		index  int
		length int
	)

	for {
		label := fmt.Sprint(options[index])

		p.Printf("\r%s%s%-*s", prompt, Input, length, label)

		length = len(label)

		i, err := readArrow()
		if err != nil {
			return 0, err
		}

		switch i {
		case arrowRight, arrowDown:
			index++

			if index >= len(options) {
				index = 0
			}
		case arrowLeft, arrowUp:
			index--

			if index < 0 {
				index = len(options) - 1
			}
		case enter:
			p.Println()

			return index, nil
		}
	}
}

// AsStrings converts a slice/array of any type to a slice of strings
func AsStrings[T any](input []T) []string {
	if strs, ok := any(input).([]string); ok {
		return strs
	}

	out := make([]string, len(input))

	for i, value := range input {
		if str, ok := any(value).(string); ok {
			out[i] = str
		} else {
			out[i] = fmt.Sprint(value)
		}
	}

	return out
}
