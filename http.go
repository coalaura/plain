package plain

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
)

const padding = "      "

type middlewareConfig struct {
	hostAsPeer bool
}

type middlewareOption func(*middlewareConfig)

// Middleware returns an http middleware that logs each request after it has been handled
func (p *Plain) Middleware(opts ...middlewareOption) func(http.Handler) http.Handler {
	var cfg middlewareConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, writer, request)

			p.logRequest(request, &metrics, &cfg)
		})
	}
}

// LogRequest writes a single formatted access log line for request using the provided metrics
func (p *Plain) LogRequest(request *http.Request, metrics *httpsnoop.Metrics) {
	p.logRequest(request, metrics, nil)
}

func (p *Plain) logRequest(request *http.Request, metrics *httpsnoop.Metrics, cfg *middlewareConfig) {
	state := p.outputState()

	bp := pool.Get().(*[]byte)

	buf := *bp
	buf = buf[:0]

	buf = p.appendHeader(buf, AnsiReset, state.color, state.theme)

	if state.color {
		buf = append(buf, state.theme.Highlight...)
	}

	method := request.Method
	buf = append(buf, method...)

	if state.color {
		buf = append(buf, AnsiReset...)
	}

	l := len(method)
	if l < 6 {
		buf = append(buf, padding[:6-l]...)
	}

	buf = append(buf, ' ')

	path := request.URL.EscapedPath()
	if path != "" {
		buf = append(buf, path...)
	} else {
		buf = append(buf, '/')
	}

	buf = append(buf, ' ')

	status := metrics.Code

	if state.color {
		switch {
		case status >= 200 && status <= 299:
			buf = append(buf, state.theme.Success...)
		case status >= 300 && status <= 399:
			buf = append(buf, state.theme.Highlight...)
		case status >= 400 && status <= 499:
			buf = append(buf, state.theme.Warn...)
		case status >= 500 && status <= 599:
			buf = append(buf, state.theme.Error...)
		}
	}

	if status < 100 || status > 999 {
		buf = append(buf, "??"...)
	} else {
		buf = append(buf, '0'+byte(status/100))
		buf = append(buf, '0'+byte((status/10)%10))
		buf = append(buf, '0'+byte(status%10))
	}

	if state.color {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, ' ')

	buf = appendDuration(buf, metrics.Duration)
	buf = append(buf, ' ')

	if state.color {
		buf = append(buf, state.theme.Dimmed...)
	}

	var addr string

	if cfg != nil && cfg.hostAsPeer {
		addr = request.Host
	} else {
		addr = request.RemoteAddr
	}

	idx := strings.LastIndexByte(addr, ':')
	if idx != -1 {
		addr = addr[:idx]
	}

	buf = append(buf, addr...)

	if state.color {
		buf = append(buf, AnsiReset...)
	}

	buf = append(buf, '\n')

	p.writeBytes(state.out, buf)

	if cap(buf) <= 4096 {
		*bp = buf
		pool.Put(bp)
	}
}

// WithHostAsPeer logs request.Host instead of the remote address, using the same port stripping.
func WithHostAsPeer() middlewareOption {
	return func(cfg *middlewareConfig) {
		cfg.hostAsPeer = true
	}
}

func appendDuration(dst []byte, dur time.Duration) []byte {
	if dur < time.Microsecond {
		dst = strconv.AppendInt(dst, dur.Nanoseconds(), 10)

		return append(dst, "ns"...)
	} else if dur < time.Millisecond {
		dst = strconv.AppendInt(dst, dur.Microseconds(), 10)

		return append(dst, "µs"...)
	} else if dur < time.Second {
		dst = strconv.AppendInt(dst, dur.Milliseconds(), 10)

		return append(dst, "ms"...)
	} else if dur < time.Minute {
		dst = strconv.AppendInt(dst, int64(dur.Seconds()), 10)

		return append(dst, 's')
	}

	dst = strconv.AppendInt(dst, int64(dur.Minutes()), 10)

	return append(dst, 'm')
}
