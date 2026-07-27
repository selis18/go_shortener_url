package main

import (
	"github.com/selis18/go_shortener_url/internal/config"
)

func main() {
	config.ParseFlags()
	InitServer()
}
