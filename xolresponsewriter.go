package xolog

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

//
// XOLResponseWriter - http.ResponseWriter and Hijacker interface
//
type XOLResponseWriter struct {
	ResponseWriter http.ResponseWriter
	statusCode     int
	contentSize    uint64
}

// NewXOLResponseWriter
func NewXOLResponseWriter(w http.ResponseWriter) *XOLResponseWriter {
	return &XOLResponseWriter{w, http.StatusOK, 0}
}

// WriteHeader - interface method
func (lrw *XOLResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

// Write - interface method
func (lrw *XOLResponseWriter) Write(p []byte) (n int, err error) {
	n, err = lrw.ResponseWriter.Write(p)
	lrw.contentSize += uint64(n)
	return
}

// Header - interface method
func (lrw *XOLResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

// Hijack - Hijacker interface method
func (lrw *XOLResponseWriter) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	wrt, ok := lrw.ResponseWriter.(http.Hijacker)
	if ok {
		rwc, buf, err = wrt.Hijack()
	} else {
		err = errors.New("xolog: underlying response writer does not implement http.Hijacker")
	}
	return
}

// Flush - Flusher interface method
func (lrw *XOLResponseWriter) Flush() {
	wrt, ok := lrw.ResponseWriter.(http.Flusher)
	if ok {
		wrt.Flush()
	}
}

// Push - Pusher interface method
func (lrw *XOLResponseWriter) Push(target string, opts *http.PushOptions) error {
	var err error
	wrt, ok := lrw.ResponseWriter.(http.Pusher)
	if ok {
		err = wrt.Push(target, opts)
	} else {
		err = errors.New("xolog: underlying response writer does not implement http.Pusher")
	}
	return err
}

// CloseNotify - CloseNotifier interface method
//func (lrw *XOLResponseWriter) CloseNotify() <-chan bool {
//	// not released
//}
