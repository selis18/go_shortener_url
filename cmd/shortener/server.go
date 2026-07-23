package main

import "net/http"

func InitServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/text/plain", mainPage)
	mux.HandleFunc("/", getUrl)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
