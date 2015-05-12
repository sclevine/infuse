package infuse

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

type fullResponse struct {
	*layeredResponse
}

func (f *fullResponse) CloseNotify() <-chan bool {
	return f.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (f *fullResponse) Flush() {
	f.ResponseWriter.(http.Flusher).Flush()
}

func (f *fullResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return f.ResponseWriter.(http.Hijacker).Hijack()
}

func (f *fullResponse) ReadFrom(src io.Reader) (n int64, err error) {
	return f.ResponseWriter.(io.ReaderFrom).ReadFrom(src)
}

func (f *fullResponse) WriteString(s string) (n int, err error) {
	return f.ResponseWriter.(stringWriter).WriteString(s)
}

type flushableResponse struct {
	*layeredResponse
}

func (f *flushableResponse) Flush() {
	f.ResponseWriter.(http.Flusher).Flush()
}
