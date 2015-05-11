package infuse

import "net/http"

// Get will retrieve a context value that is shared by all http.Handlers
// attached to the same infuse.Handler. The context value is associated with
// an http.ResponseWriter, so it has the same life cycle as the provided
// response. Get will return nil if the provided response does not have an
// associated context value.
//
// If the context value is a pointer, map, or slice, then changes to the data
// in one http.Handler will be seen by other http.Handlers that share it.
//
// To retrieve a value of a particular type, consider wrapping Get and Set in
// another function. For example, this function could be used to share a map
// of floating-point numbers between middleware handlers:
//
//   func HandlerMap(response http.ResponseWriter) map[string]float64 {
//      handlerMap, ok := infuse.Get(response).(map[string]float64)
//      if !ok || handlerMap == nil {
//         handlerMap = make(map[string]float64)
//         if ok := infuse.Set(response, handlerMap); !ok {
//            return nil
//         }
//      }
//      return handlerMap
//   }
func Get(response http.ResponseWriter) interface{} {
	sharedResponse, ok := response.(infuseResponse)
	if !ok {
		return nil
	}
	return sharedResponse.get()
}

// Set will store a context value that is shared by all http.Handlers attached
// to the same infuse.Handler. The context value is associated with an
// http.ResponseWriter, so it has the same life cycle as the provided response.
//
// The boolean return value indicates whether the setting the context value
// succeeded. Set will return false if the provided response is invalid.
func Set(response http.ResponseWriter, value interface{}) bool {
	sharedResponse, ok := response.(infuseResponse)
	if !ok {
		return false
	}
	sharedResponse.set(value)
	return true
}

type contextualResponse struct {
	http.ResponseWriter
	context interface{}
}

func (c *contextualResponse) get() interface{} {
	return c.context
}

func (c *contextualResponse) set(value interface{}) {
	c.context = value
}
