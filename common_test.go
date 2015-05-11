package infuse_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHandlerResponse(t *testing.T, body string, fixture string) {
	expected := strings.TrimSpace(fixture)
	if trimmedBody := strings.TrimSpace(body); trimmedBody != expected {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", expected, trimmedBody)
	}
}

func serve(handler http.Handler) string {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, &http.Request{})
	return response.Body.String()
}
