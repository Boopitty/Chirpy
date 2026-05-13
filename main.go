package main

import (
	"fmt"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	err := http.ListenAndServe(server.Addr, server.Handler)
	if err != nil {
		fmt.Printf("Problem running server: %s", err)
	}
}
