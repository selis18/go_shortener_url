package main

import "net/http"

func InitServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", postUrl)
	mux.HandleFunc("/", getUrl)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
