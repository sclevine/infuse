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

func basicAuth(response http.ResponseWriter, request *http.Request) {
	username, password, ok := request.BasicAuth()
	if !ok || !userValid(username, password) {
		response.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(response, "Permission Denied!")
		return
	}
	if !infuse.Set(response, username) || !infuse.Next(response, request) {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(response, "Server Error!")
	}
}

func userGreeter(response http.ResponseWriter, request *http.Request) {
	username := infuse.Get(response).(string)
	fmt.Fprintf(response, "Hello %s!", username)
}

func userDeleter(response http.ResponseWriter, request *http.Request) {
	username := infuse.Get(response).(string)
	fmt.Fprintf(response, "Deleting %s!", username)
	deleteUser(username)
}

func main() {
	authHandler := infuse.New().HandleFunc(basicAuth)
	router := http.NewServeMux()
	router.Handle("/greet", authHandler.HandleFunc(userGreeter))
	router.Handle("/delete", authHandler.HandleFunc(userDeleter))
	http.ListenAndServe(":8080", router)
}
```

The `mock` package makes it easy to unit test code that depends on an
`infuse.Handler`.