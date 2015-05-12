Infuse
======

[![Build Status](https://api.travis-ci.org/sclevine/infuse.png?branch=master)](http://travis-ci.org/sclevine/infuse)
[![GoDoc](https://godoc.org/github.com/sclevine/infuse?status.svg)](https://godoc.org/github.com/sclevine/infuse)

Infuse provides an immutable, concurrency-safe middleware handler
that conforms to `http.Handler`. An `infuse.Handler` is fully compatible with
the Go standard library, supports flexible chaining, and provides a shared
context between middleware handlers without relying on global state, locks,
or shared closures.

BasicAuth example:
```go

func Example() {
	authHandler := infuse.New().HandleFunc(basicAuth)
	router := http.NewServeMux()
	router.Handle("/hello", authHandler.HandleFunc(userGreeting))
	router.Handle("/goodbye", authHandler.HandleFunc(userFarewell))
	server := httptest.NewServer(router)
	defer server.Close()

	doRequest(server.URL+"/hello", "bob", "1234")
	doRequest(server.URL+"/goodbye", "alice", "5678")
	doRequest(server.URL+"/goodbye", "intruder", "guess")

	// Output:
	// Hello bob!
	// Goodbye alice!
	// Permission Denied
}

func userGreeting(response http.ResponseWriter, request *http.Request) {
	username := infuse.Get(response).(string)
	fmt.Fprintf(response, "Hello %s!", username)
}

func userFarewell(response http.ResponseWriter, request *http.Request) {
	username := infuse.Get(response).(string)
	fmt.Fprintf(response, "Goodbye %s!", username)
}

func basicAuth(response http.ResponseWriter, request *http.Request) {
	username, password, ok := request.BasicAuth()
	if !ok || !userValid(username, password) {
		response.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(response, "Permission Denied")
		return
	}

	if ok := infuse.Set(response, username); !ok {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(response, "Server Error")
	}

	if ok := infuse.Next(response, request); !ok {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(response, "Server Error")
	}
}
```

The `mock` package makes it easy to unit test code that depends on an
`infuse.Handler`.