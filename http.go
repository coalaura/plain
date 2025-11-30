package plain

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
)

func (p *Plain) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, writer, request)

			p.LogRequest(request, &metrics)
		})
	}
}

func (p *Plain) LogRequest(request *http.Request, metrics *httpsnoop.Metrics) {
	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)

	buf.Reset()

	p.writeHeader(buf, Reset)

	if p.color {
		buf.WriteString(p.theme.Highlight)
	}

	method := request.Method

	buf.WriteString(method)

	if p.color {
		buf.WriteString(Reset)
	}

	for i := len(method); i < 6; i++ {
		buf.WriteByte(' ')
	}

	buf.WriteByte(' ')

	if path := request.URL.EscapedPath(); path != "" {
		buf.WriteString(path)
	} else {
		buf.WriteByte('/')
	}

	buf.WriteByte(' ')

	status := metrics.Code

	if p.color {
		switch {
		case status >= 200 && status <= 299:
			buf.WriteString(p.theme.Success)
		case status >= 300 && status <= 399:
			buf.WriteString(p.theme.Highlight)
		case status >= 400 && status <= 499:
			buf.WriteString(p.theme.Warn)
		case status >= 500 && status <= 599:
			buf.WriteString(p.theme.Error)
		}
	}

	if status < 100 || status > 999 {
		buf.WriteString("???")
	} else {
		buf.WriteByte('0' + byte(status/100))
		buf.WriteByte('0' + byte((status/10)%10))
		buf.WriteByte('0' + byte(status%10))
	}

	if p.color {
		buf.WriteString(Reset)
	}

	buf.WriteByte(' ')

	writeDuration(buf, metrics.Duration)

	buf.WriteByte(' ')

	if p.color {
		buf.WriteString(p.theme.Dimmed)
	}

	addr, _, _ := net.SplitHostPort(request.RemoteAddr)

	buf.WriteString(addr)

	if p.color {
		buf.WriteString(Reset)
	}

	buf.WriteByte('\n')

	p.out.Write(buf.Bytes())
}

func writeDuration(buf *bytes.Buffer, dur time.Duration) {
	if dur < time.Microsecond {
		buf.WriteString(strconv.FormatInt(dur.Nanoseconds(), 10))
		buf.WriteString("ns")
	} else if dur < time.Millisecond {
		buf.WriteString(strconv.FormatInt(dur.Microseconds(), 10))
		buf.WriteString("µs")
	} else if dur < time.Second {
		buf.WriteString(strconv.FormatInt(dur.Milliseconds(), 10))
		buf.WriteString("ms")
	} else if dur < time.Minute {
		buf.WriteString(strconv.FormatInt(dur.Milliseconds()/1000, 10))
		buf.WriteByte('s')
	} else {
		buf.WriteString(strconv.FormatInt(dur.Milliseconds()/1000/60, 10))
		buf.WriteByte('m')
	}
}
