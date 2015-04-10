package mock

import (
	"net/http"

	"github.com/sclevine/infuse"
)

type Serve struct {
	Response http.ResponseWriter
	Request  *http.Request
}

type Handler struct {
	Handlers []http.Handler
	Serves   []Serve
}

func (h *Handler) Handle(handler http.Handler) infuse.Handler {
	h.Handlers = append(h.Handlers, handler)
	return h
}

func (h *Handler) HandleFunc(handler func(http.ResponseWriter, *http.Request)) infuse.Handler {
	h.Handlers = append(h.Handlers, http.HandlerFunc(handler))
	return h
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	h.Serves = append(h.Serves, Serve{response, request})
}
