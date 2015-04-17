// Package mock makes it easy to unit test code that depends on an
// infuse.Handler.
package mock

import (
	"net/http"

	"github.com/sclevine/infuse"
)

// Handler is a mock handler that will call a StubFunc when it is served.
type Handler struct {
	handlers []http.Handler
	stub     StubFunc
}

// A StubFunc is called when a mock.Handler is served. The third argument
// consists of all of the http.Handlers in the middleware chain that a real
// infuse.Handler would call in the order that they would be called.
type StubFunc func(http.ResponseWriter, *http.Request, []http.Handler)

// Stub provides a StubFunc to a mock.Handler. If a mock.Handler is served
// without a StubFunc, it will panic. The provided StubFunc will be inherited
// by any derived handlers
func (h *Handler) Stub(stub StubFunc) {
	h.stub = stub
}

func (h *Handler) Handle(handler http.Handler) infuse.Handler {
	return &Handler{handlers: append(h.handlers, handler), stub: h.stub}
}

func (h *Handler) HandleFunc(handler func(http.ResponseWriter, *http.Request)) infuse.Handler {
	return h.Handle(http.HandlerFunc(handler))
}

func (h *Handler) Stack(handler http.Handler) infuse.Handler {
	return h.HandleFunc(func(response http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(response, request)
		infuse.Next(response, request)
	})
}

func (h *Handler) StackFunc(handler func(http.ResponseWriter, *http.Request)) infuse.Handler {
	return h.Stack(http.HandlerFunc(handler))
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if h.stub == nil {
		panic("Mock infuse.Handler missing stub.")
	}
	h.stub(response, request, h.handlers)
}
