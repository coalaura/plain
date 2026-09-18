package plain

import (
	"net/http"
	"testing"
)

type middlewareResponseWriter struct{}

func (middlewareResponseWriter) Header() http.Header {
	return nil
}

func (middlewareResponseWriter) Write(body []byte) (int, error) {
	return len(body), nil
}

func (middlewareResponseWriter) WriteHeader(int) {}

func (middlewareResponseWriter) ServeHTTP(http.ResponseWriter, *http.Request) {}

func BenchmarkRealIPMiddlewareNoHeader(b *testing.B) {
	request := &http.Request{
		Header:     make(http.Header),
		RemoteAddr: "192.0.2.1:8080",
	}

	handler := RealIPMiddleware(WithTrustedProxies("192.0.2.0/24"))(middlewareResponseWriter{})

	writer := middlewareResponseWriter{}

	b.ReportAllocs()

	for b.Loop() {
		handler.ServeHTTP(writer, request)
	}
}

func BenchmarkRealIPMiddlewareReplaceAddress(b *testing.B) {
	request := &http.Request{
		Header:     http.Header{HeaderXForwardedFor: {"198.51.100.1, 203.0.113.1"}},
		RemoteAddr: "192.0.2.1:8080",
	}

	handler := RealIPMiddleware(WithTrustedProxies("192.0.2.0/24"))(middlewareResponseWriter{})

	writer := middlewareResponseWriter{}

	b.ReportAllocs()

	for b.Loop() {
		request.RemoteAddr = "192.0.2.1:8080"

		handler.ServeHTTP(writer, request)
	}
}
