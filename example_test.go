// +build go1.4

package infuse_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/sclevine/infuse"
)

func Example() {
	authHandler := infuse.New().HandleFunc(basicAuth)
	router := http.NewServeMux()
	router.Handle("/hello", authHandler.HandleFunc(userGreeting))
	router.Handle("/goodbye", authHandler.HandleFunc(userFarewell))
	port := freePort()
	go http.ListenAndServe(":"+port, router)

	doRequest(fmt.Sprintf("http://bob:1234@localhost:%s/hello", port))
	doRequest(fmt.Sprintf("http://alice:5678@localhost:%s/goodbye", port))
	doRequest(fmt.Sprintf("http://intruder:guess@localhost:%s/goodbye", port))

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

func userValid(username, password string) bool {
	if username == "bob" && password == "1234" {
		return true
	}
	if username == "alice" && password == "5678" {
		return true
	}
	return false
}

func doRequest(url string) {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", body)
}

func freePort() string {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	return strings.SplitN(listener.Addr().String(), ":", 2)[1]
}
