package xolog

import (
	"errors"
	"os"
	"sync"
)

//
// XOLWriter - io.Writer interface
//
type XOLWriter struct {
	mu       sync.Mutex
	writer   *os.File
	ondemand bool
	buf      []byte
}

// NewXOLWriter
func NewXOLWriter(out interface{}, ondemand bool) (wrt *XOLWriter, err error) {
	var writer *os.File

	switch o := out.(type) {
	case string:
		writer, err = os.OpenFile(o, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	case *os.File:
		writer = o
	default:
		err = errors.New("Unknown writer")
	}

	if err == nil {
		wrt = &XOLWriter{writer: writer, ondemand: ondemand}
	}

	return
}

// Write - interface method
func (lw *XOLWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	if lw.ondemand {
		lw.buf = append(lw.buf, p...)
		n = len(p)
	} else {
		n, err = lw.writer.Write(p)
	}
	lw.mu.Unlock()

	return
}

// Flush
func (lw *XOLWriter) Flush() (n int, err error) {
	lw.mu.Lock()
	if len(lw.buf) > 0 {
		n, err = lw.writer.Write(lw.buf)
		lw.buf = lw.buf[:0]
	}
	lw.mu.Unlock()

	return
}

// Reopen
func (lw *XOLWriter) Reopen() (err error) {
	var writer *os.File

	lw.mu.Lock()
	var fname string = lw.writer.Name()
	if fname != "" && fname != "/dev/stdin" && fname != "/dev/stdout" && fname != "/dev/stderr" {
		lw.writer.Close()
		writer, err = os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		lw.writer = writer
	}
	lw.mu.Unlock()

	return
}

// Close
func (lw *XOLWriter) Close() (err error) {
	lw.mu.Lock()
	var fname string = lw.writer.Name()
	if fname != "" && fname != "/dev/stdin" && fname != "/dev/stdout" && fname != "/dev/stderr" {
		err = lw.writer.Close()
	}
	lw.mu.Unlock()

	return
}
