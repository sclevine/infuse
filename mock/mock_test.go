package mock_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sclevine/infuse"
	"github.com/sclevine/infuse/mock"
)

var mockHandlerFixture = `
start stub
start first
end first
start second
end second
start third
end third
start fourth
end fourth
end stub`

func TestMock(t *testing.T) {
	mockHandler := &mock.Handler{}
	mockHandler.Stub(func(response http.ResponseWriter, request *http.Request, handlers []http.Handler) {
		fmt.Fprintln(response, "start stub")
		for _, handler := range handlers {
			handler.ServeHTTP(response, request)
		}
		fmt.Fprintln(response, "end stub")
	})
	handler := setupHandlers(mockHandler)
	testHandlerResponse(t, serve(handler), mockHandlerFixture)
}

func testHandlerResponse(t *testing.T, body string, fixture string) {
	expected := strings.TrimSpace(fixture)
	if trimmedBody := strings.TrimSpace(body); trimmedBody != expected {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", expected, trimmedBody)
	}
}

func setupHandlers(handler infuse.Handler) http.Handler {
	handler = handler.Handle(http.HandlerFunc(buildHandler("first")))
	handler = handler.HandleFunc(buildHandler("second"))
	handler = handler.Stack(http.HandlerFunc(buildHandler("third")))
	handler = handler.StackFunc(buildHandler("fourth"))
	return handler
}

func serve(handler http.Handler) string {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

func buildHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(response, "start %s\n", name)
		infuse.Next(response, request)
		fmt.Fprintf(response, "end %s\n", name)
	}
}
