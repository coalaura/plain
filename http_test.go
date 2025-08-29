package plain

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	p := New(WithDate(RFC3339Local))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(100 + rand.Intn(420))
	})

	wrapped := p.Middleware()(handler)

	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "http://localhost/test/path", nil)
			req.RemoteAddr = fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(65535))

			wrapped.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}
