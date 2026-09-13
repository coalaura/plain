package minimal

import (
	"io"
	"testing"
)

func BenchmarkInfoln(b *testing.B) {
	m := New(WithTarget(io.Discard))

	for b.Loop() {
		m.Infoln("hello", "world", 123)
	}
}

func BenchmarkInfof(b *testing.B) {
	m := New(WithTarget(io.Discard))

	for b.Loop() {
		m.Infof("%s %s %d", "hello", "world", 123)
	}
}
