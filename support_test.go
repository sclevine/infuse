package infuse_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sclevine/infuse"
)

func testHandlerResponse(t *testing.T, handler http.Handler, fixture string) {
	expected := strings.TrimSpace(fixture)
	if body := strings.TrimSpace(serve(handler)); body != expected {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", expected, body)
	}
}

func serve(handler http.Handler) string {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}

func buildHandler(name string, nexts int) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(response, "start %s\n", name)

		for i := 0; i < nexts; i++ {
			fmt.Fprintf(response, "attempting next for %s\n", name)
			if infuse.Next(response, request) {
				fmt.Fprintf(response, "finished next for %s\n", name)
			} else {
				fmt.Fprintf(response, "no next for %s\n", name)
			}
		}

		fmt.Fprintf(response, "end %s\n", name)
	}
}

func panicHandler(response http.ResponseWriter, _ *http.Request) {
	if infuse.Get(response) != nil {
		fmt.Fprintf(response, "already panicked\n")
	} else {
		fmt.Fprintf(response, "panicking\n")
		infuse.Set(response, true)
		panic("some error")
	}
}

func recoverHandler(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(response, "recovered and attempting next")
			infuse.Next(response, request)
			fmt.Fprintln(response, "finished next after recovery")
		}
	}()
	fmt.Fprintln(response, "start recoverable")
	infuse.Next(response, request)
	fmt.Fprintln(response, "end recoverable")
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
