package plain

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMiddleware(b *testing.B) {
	benchmarkMiddleware(b)
}

func BenchmarkMiddlewareHostAsPeer(b *testing.B) {
	benchmarkMiddleware(b, WithHostAsPeer())
}

func benchmarkMiddleware(b *testing.B, opts ...middlewareOption) {
	logger := New(WithTarget(io.Discard))

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	wrapped := logger.Middleware(opts...)(handler)

	request := httptest.NewRequest(http.MethodGet, "http://example.com/test/path", nil)

	request.RemoteAddr = "192.0.2.1:8080"

	writer := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		wrapped.ServeHTTP(writer, request)
	}
}
