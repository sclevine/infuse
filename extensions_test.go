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
CloseNotify called
Flush called
Hijack called
ReadFrom called with some data
WriteString called with some string`

func TestExtendedResponseMethods(t *testing.T) {
	handler := infuse.New().HandleFunc(extendedMethodsPresentHandler)
	testHandlerResponse(t, serveMock(handler), extendedResponseMethodsFixture)

	handler = infuse.New().HandleFunc(extendedMethodsMissingHandler)
	testHandlerResponse(t, serve(handler), "")
}

func extendedMethodsPresentHandler(response http.ResponseWriter, _ *http.Request) {
	response.(http.CloseNotifier).CloseNotify()
	response.(http.Flusher).Flush()
	response.(http.Hijacker).Hijack()
	response.(io.ReaderFrom).ReadFrom(strings.NewReader("some data"))
	response.(stringWriter).WriteString("some string")
}

func extendedMethodsMissingHandler(response http.ResponseWriter, _ *http.Request) {
	if _, ok := response.(http.CloseNotifier); ok {
		panic("Response should not implement http.CloseNotifier.")
	}
	if _, ok := response.(http.Flusher); ok {
		panic("Response should not implement http.Flusher.")
	}
	if _, ok := response.(http.Hijacker); ok {
		panic("Response should not implement http.Hijacker.")
	}
	if _, ok := response.(io.ReaderFrom); ok {
		panic("Response should not implement io.ReaderFrom.")
	}
	if _, ok := response.(stringWriter); ok {
		panic("Response should not implement WriteString method.")
	}
}

func serveMock(handler http.Handler) string {
	response := &mockResponse{httptest.NewRecorder()}
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

type mockResponse struct {
	*httptest.ResponseRecorder
}

func (m *mockResponse) CloseNotify() <-chan bool {
	fmt.Fprintf(m, "CloseNotify called\n")
	return nil
}

func (m *mockResponse) Flush() {
	fmt.Fprintf(m, "Flush called\n")
}

func (m *mockResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	fmt.Fprintf(m, "Hijack called\n")
	return nil, nil, nil
}

func (m *mockResponse) ReadFrom(src io.Reader) (n int64, err error) {
	srcStr, err := ioutil.ReadAll(src)
	if err == nil {
		fmt.Fprintf(m, "ReadFrom called with %s\n", srcStr)
	}
	return 0, nil
}

func (m *mockResponse) WriteString(s string) (n int, err error) {
	fmt.Fprintf(m, "WriteString called with %s\n", s)
	return 0, nil
}
