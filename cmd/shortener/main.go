package main

import (
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/logger"
)

func main() {
	config.ParseConfig()
	if err := logger.Initialize(config.GetLogLevel()); err != nil {
		return
	}
	InitServer()
}
