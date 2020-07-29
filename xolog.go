package xolog

import (
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC
	XOLQueryString
	XOLflushondemand
	LstdFlags = Ldate | Ltime
)

//
// XOLogger
//
type XOLogger struct {
	writer *XOLWriter
	logger *log.Logger
	flag   int
}

// NewXOLogger
func NewXOLogger(out interface{}, prefix string, flag int) (xlog *XOLogger, err error) {
	var writer *XOLWriter

	var ondemand bool
	if flag&XOLflushondemand != 0 {
		ondemand = true
	}
	writer, err = NewXOLWriter(out, ondemand)

	if err == nil {
		xlog = &XOLogger{
			writer: writer,
			logger: log.New(writer, prefix, flag&^XOLQueryString&^XOLflushondemand),
			flag:   flag,
		}
	}

	return
}

// GetLogger
func (l *XOLogger) GetLogger() *log.Logger {
	return l.logger
}

// Flush
func (l *XOLogger) Flush() (n int, err error) {
	n, err = l.writer.Flush()
	return
}

// Close
func (l *XOLogger) Close() (err error) {
	err = l.writer.Close()
	return
}

// LogWrapper
func (l *XOLogger) LogWrapper(lw http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := NewXOLResponseWriter(w)
		lw.ServeHTTP(ww, r)
		l.logHttp(time.Now().Sub(start).Nanoseconds(), ww, r)
	})
}

// logHttp
func (l *XOLogger) logHttp(duration int64, lrw *XOLResponseWriter, r *http.Request) {
	var login, mediatype, scheme, ua, referer, raddr, rport, host, qs string

	login, _, _ = r.BasicAuth()
	if len(login) == 0 {
		login = "-"
	}

	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	if r.UserAgent() == "" {
		ua = "-"
	} else {
		ua = r.UserAgent()
	}

	if r.Referer() == "" {
		referer = "-"
	} else {
		referer = r.Referer()
	}

	raddr, rport, _ = net.SplitHostPort(r.RemoteAddr)
	if raddr == "" {
		raddr = "-"
	}
	if rport == "" {
		rport = "-"
	}

	host = r.Host
	for i := len(host) - 1; i > 0; i-- {
		if host[i] == ']' {
			break
		}
		if host[i] == ':' {
			host = host[:i]
			break
		}
	}

	if r.URL.RawQuery != "" {
		qs = "?" + r.URL.RawQuery
	} else if r.URL.ForceQuery {
		qs = "?"
	}

	h := lrw.Header()
	if ct, ok := h["Content-Type"]; ok {
		if len(ct) > 0 {
			mediatype, _, _ = mime.ParseMediaType(ct[0])
		}
	}
	if len(mediatype) == 0 {
		mediatype = "-"
	}

	l.Printf("%s %s %s %s %s %s %d %d %s \"%s://%s%s%s\" \"%s\" \"%s\"\n",
		NSecondsToSeconds(duration), raddr, rport, url.PathEscape(login), r.Proto, r.Method, lrw.statusCode, lrw.contentSize, mediatype,
		scheme, host, r.URL.Path, qs, ua, referer)
}

// LogHttpRequest
func (l *XOLogger) LogHttpRequest(r *http.Request) {
	var scheme, ua, referer, addr, host, qs string
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	if r.UserAgent() == "" {
		ua = "-"
	} else {
		ua = r.UserAgent()
	}

	if r.Referer() == "" {
		referer = "-"
	} else {
		referer = r.Referer()
	}

	addr = r.RemoteAddr
	for i := len(addr) - 1; i > 0; i-- {
		if addr[i] == ']' {
			break
		}
		if addr[i] == ':' {
			addr = addr[:i]
			break
		}
	}

	host = r.Host
	for i := len(host) - 1; i > 0; i-- {
		if host[i] == ']' {
			break
		}
		if host[i] == ':' {
			host = host[:i]
			break
		}
	}

	if l.flag&XOLQueryString != 0 {
		if r.URL.RawQuery != "" {
			qs = "?" + r.URL.RawQuery
		} else if r.URL.ForceQuery {
			qs = "?"
		}
	}

	l.logger.Output(2, fmt.Sprintf("%s %s %s %s://%s%s%s \"%s\" \"%s\"\n", addr, r.Proto, r.Method, scheme, host, r.URL.Path, qs, ua, referer))
}

//
// NSecondsToSeconds
//
func NSecondsToSeconds(n int64) (s string) {
	s = strconv.FormatInt(n, 10)
	slen := len(s)
	if slen < 10 {
		s = strings.Repeat("0", 10-slen) + s
		slen = 10
	}
	s = s[:slen-9] + "." + s[slen-9:]
	return
}

// Printf
func (l *XOLogger) Printf(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
}

// Print
func (l *XOLogger) Print(v ...interface{}) {
	l.logger.Output(2, fmt.Sprint(v...))
}

// Println
func (l *XOLogger) Println(v ...interface{}) {
	l.logger.Output(2, fmt.Sprintln(v...))
}

// Fatal
func (l *XOLogger) Fatal(v ...interface{}) {
	l.logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf
func (l *XOLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln
func (l *XOLogger) Fatalln(v ...interface{}) {
	l.logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic
func (l *XOLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.logger.Output(2, s)
	panic(s)
}

// Panicf
func (l *XOLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.logger.Output(2, s)
	panic(s)
}

// Panicln
func (l *XOLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.logger.Output(2, s)
	panic(s)
}
