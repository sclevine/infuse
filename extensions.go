package infuse

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type extendedResponse interface {
	http.CloseNotifier
	http.Flusher
	http.Hijacker
	io.ReaderFrom
	stringWriter
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

type httpResponse struct {
	*layeredResponse
}

func (h *httpResponse) CloseNotify() <-chan bool {
	return h.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (h *httpResponse) Flush() {
	h.ResponseWriter.(http.Flusher).Flush()
}

func (h *httpResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.ResponseWriter.(http.Hijacker).Hijack()
}

func (h *httpResponse) ReadFrom(src io.Reader) (n int64, err error) {
	return h.ResponseWriter.(io.ReaderFrom).ReadFrom(src)
}

func (h *httpResponse) WriteString(s string) (n int, err error) {
	return h.ResponseWriter.(stringWriter).WriteString(s)
}
