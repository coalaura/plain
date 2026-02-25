package plain

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/felixge/httpsnoop"
)

func BenchmarkPrintln(b *testing.B) {
	p := New(WithTarget(io.Discard))

	for b.Loop() {
		p.Println("hello", "world", 123)
	}
}

func BenchmarkPrintf(b *testing.B) {
	p := New(WithTarget(io.Discard))

	for b.Loop() {
		p.Printf("%s %s %d", "hello", "world", 123)
	}
}

func BenchmarkLogRequest(b *testing.B) {
	p := New(WithTarget(io.Discard))

	req := httptest.NewRequest("GET", "http://localhost/test/path", nil)

	req.RemoteAddr = "127.0.0.1:1234"

	metrics := httpsnoop.Metrics{
		Code:     200,
		Duration: 1500 * time.Microsecond,
	}

	for b.Loop() {
		p.LogRequest(req, &metrics)
	}
}
