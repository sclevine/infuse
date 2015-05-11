package infuse_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sclevine/infuse"
)

func TestGetAndSet(t *testing.T) {
	handler := infuse.New().HandleFunc(createMapHandler)
	handler = handler.HandleFunc(buildSetMapHandler("first key", "first value"))
	handler = handler.HandleFunc(buildSetMapHandler("second key", "second value"))
	handler = handler.HandleFunc(buildSetMapHandler("first key", "new first value"))
	handler = handler.HandleFunc(buildOutputMapHandler("first key"))
	handler = handler.HandleFunc(buildOutputMapHandler("second key"))
	testHandlerResponse(t, serve(handler), "first key: new first value\nsecond key: second value\n")
}

func TestGetAndSetForNestedHandlers(t *testing.T) {
	firstGroup := infuse.New().HandleFunc(createMapHandler)
	firstGroup = firstGroup.HandleFunc(buildSetMapHandler("key", "first value"))

	secondGroup := infuse.New().HandleFunc(createMapHandler)
	secondGroup = secondGroup.HandleFunc(buildSetMapHandler("key", "second value"))
	secondGroup = secondGroup.HandleFunc(buildOutputMapHandler("key"))

	handler := firstGroup.Stack(secondGroup)
	handler = handler.HandleFunc(buildOutputMapHandler("key"))

	testHandlerResponse(t, serve(handler), "key: second value\nkey: first value\n")
}

func TestInvalidResponseForGetAndSet(t *testing.T) {
	if context := infuse.Get(nil); context != nil {
		t.Fatalf("Expected nil context from invalid response, got %s.", context)
	}
	if ok := infuse.Set(nil, "value"); ok {
		t.Fatal("Expected failure to set context on invalid response.")
	}
}

func createMapHandler(response http.ResponseWriter, request *http.Request) {
	if ok := infuse.Set(response, make(map[string]string)); !ok {
		panic("Failed to set map.")
	}
	infuse.Next(response, request)
}

func buildOutputMapHandler(key string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		shared := infuse.Get(response).(map[string]string)
		fmt.Fprintf(response, "%s: %s\n", key, shared[key])
		infuse.Next(response, request)
	}
}

func buildSetMapHandler(key, value string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		shared := infuse.Get(response).(map[string]string)
		shared[key] = value
		infuse.Next(response, request)
	}
}
