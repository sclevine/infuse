package infuse

import "net/http"

type infuseResponse interface {
	next(request *http.Request) bool
	get() interface{}
	set(value interface{})
}

type layeredResponse struct {
	*contextualResponse
	layers []*layer
}

func newLayeredResponse(response http.ResponseWriter) *layeredResponse {
	return &layeredResponse{&contextualResponse{response, nil}, nil}
}

func (l *layeredResponse) next(request *http.Request) bool {
	if len(l.layers) == 0 {
		return false
	}

	next := l.layers[len(l.layers)-1]
	remaining := l.layers[:len(l.layers)-1]
	sharedResponse := &layeredResponse{l.contextualResponse, remaining}
	next.handler.ServeHTTP(sharedResponse.extend(), request)
	return true
}

func (l *layeredResponse) extend() http.ResponseWriter {
	if _, ok := l.ResponseWriter.(extendedResponse); !ok {
		return l
	}
	return &httpResponse{l}
}
