package plain

import "errors"

var ErrInterrupted = errors.New("interrupted")

// IsInterrupted checks if the given err is ErrInterrupted
func IsInterrupted(err error) bool {
	return errors.Is(err, ErrInterrupted)
}
