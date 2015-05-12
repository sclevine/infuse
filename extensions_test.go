package infuse_test

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sclevine/infuse"
)

var extendedResponseMethodsFixture = `
attempting extensions
CloseNotify called
Flush called
Hijack called
ReadFrom called with some data
WriteString called with some string`

var flushableResponseMethodsFixture = `
attempting extensions
Flush called`

func TestResponseMethods(t *testing.T) {
	handler := infuse.New().HandleFunc(responseMethodsHandler)
	testHandlerResponse(t, serveExtended(handler), extendedResponseMethodsFixture)
	testHandlerResponse(t, serveFlushable(handler), flushableResponseMethodsFixture)
	testHandlerResponse(t, serveLimited(handler), "attempting extensions")
}

func responseMethodsHandler(response http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(response, "attempting extensions")

	if _, ok := response.(http.CloseNotifier); ok {
		response.(http.CloseNotifier).CloseNotify()
	}
	if _, ok := response.(http.Flusher); ok {
		response.(http.Flusher).Flush()
	}
	if _, ok := response.(http.Hijacker); ok {
		response.(http.Hijacker).Hijack()
	}
	if _, ok := response.(io.ReaderFrom); ok {
		response.(io.ReaderFrom).ReadFrom(strings.NewReader("some data"))
	}
	if _, ok := response.(stringWriter); ok {
		response.(stringWriter).WriteString("some string")
	}
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

func serveExtended(handler http.Handler) string {
	response := &extendedResponse{httptest.NewRecorder()}
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

func serveFlushable(handler http.Handler) string {
	response := &flushableResponse{httptest.NewRecorder()}
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

func serveLimited(handler http.Handler) string {
	response := &limitedResponse{httptest.NewRecorder()}
	handler.ServeHTTP(response, &http.Request{})
	return response.ResponseWriter.(*httptest.ResponseRecorder).Body.String()
}

type extendedResponse struct {
	*httptest.ResponseRecorder
}

func (e *extendedResponse) CloseNotify() <-chan bool {
	fmt.Fprintln(e, "CloseNotify called")
	return nil
}

func (e *extendedResponse) Flush() {
	fmt.Fprintln(e, "Flush called")
}

func (e *extendedResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	fmt.Fprintln(e, "Hijack called")
	return nil, nil, nil
}

func (e *extendedResponse) ReadFrom(src io.Reader) (n int64, err error) {
	srcStr, err := ioutil.ReadAll(src)
	if err == nil {
		fmt.Fprintf(e, "ReadFrom called with %s\n", srcStr)
	}
	return 0, nil
}

func (e *extendedResponse) WriteString(s string) (n int, err error) {
	fmt.Fprintf(e, "WriteString called with %s\n", s)
	return 0, nil
}

type flushableResponse struct {
	*httptest.ResponseRecorder
}

func (f *flushableResponse) Flush() {
	fmt.Fprintln(f, "Flush called")
}

type limitedResponse struct {
	http.ResponseWriter
}
