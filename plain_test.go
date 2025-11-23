package plain

import (
	"testing"
)

func TestPlain(t *testing.T) {
	p := New(WithDate(RFC3339Local))

	p.Debugln("Hello from Debugln")
	p.Println("Hello from Println")
	p.Warnln("Hello from Warnln")
	p.Errorln("Hello from Errorln")
}
