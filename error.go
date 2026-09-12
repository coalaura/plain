package plain

import (
	"errors"

	"github.com/coalaura/plain/internal"
)

var (
	// ErrInterrupted indicates that an input operation was cancelled.
	ErrInterrupted = internal.ErrInterrupted

	// ErrNoOptions indicates that a selector was opened without any options.
	ErrNoOptions = errors.New("no options")
)

// IsInterrupted checks if the given err is ErrInterrupted
func IsInterrupted(err error) bool {
	return errors.Is(err, ErrInterrupted)
}
