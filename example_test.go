// +build go1.4

package infuse_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/sclevine/infuse"
)

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

func userValid(username, password string) bool {
	if username == "bob" && password == "1234" {
		return true
	}
	if username == "alice" && password == "5678" {
		return true
	}
	return false
}

func doRequest(url, username, password string) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	request.SetBasicAuth(username, password)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", body)
}
