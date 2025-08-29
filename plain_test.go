package plain

import (
	"testing"
)

func TestPlain(t *testing.T) {
	p := New(WithDate(RFC3339Local))

	p.Println("Hello from Println")
	p.Warnln("Hello from Warnln")
	p.Errorln("Hello from Errorln")
}
