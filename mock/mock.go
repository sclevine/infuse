package mock

import (
	"net/http"

	"github.com/sclevine/infuse"
)

type Handler struct {
	handlers []http.Handler
	stub     StubFunc
}

type StubFunc func(http.ResponseWriter, *http.Request, []http.Handler)

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

func (h *Handler) Stub(stub StubFunc) {
	h.stub = stub
}
