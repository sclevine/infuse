package mock_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sclevine/infuse"
	"github.com/sclevine/infuse/mock"
)

func TestMock(t *testing.T) {
	mockHandler := &mock.Handler{}
	mockHandler.Stub(func(response http.ResponseWriter, request *http.Request, handlers []http.Handler) {
		fmt.Fprintln(response, "start stub")
		for _, handler := range handlers {
			handler.ServeHTTP(response, request)
		}
		fmt.Fprintln(response, "end stub")
	})
	responseBody := serve(setupHandlers(mockHandler))
	if responseBody != mockHandlerFixture {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", mockHandlerFixture, body)
	}
}

func setupHandlers(handler infuse.Handler) http.Handler {
	handler = handler.Handle(http.HandlerFunc(buildHandler("first")))
	handler = handler.HandleFunc(buildHandler("second"))
	handler = handler.Stack(http.HandlerFunc(buildHandler("third")))
	handler = handler.StackFunc(buildHandler("third"))
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
