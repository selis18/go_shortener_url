package config

import (
	"errors"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	LogLevel      string `env:"LOG_LEVEL"`
}

type Address struct {
	Host string
	Port int
}

func (a *Address) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *Address) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

var address = Address{
	Host: "localhost",
	Port: 8080,
}
var host string = "http://localhost:8080/"
var level string = "INFO"

func ParseConfig() {
	flag.Var(&address, "a", "host and port to start server")
	flag.StringVar(&host, "b", "http://localhost:8080/", "base address to result")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerAddress != "" {
		address.Set(cfg.ServerAddress)
	}

	if cfg.BaseURL != "" {
		host = cfg.BaseURL
	}

	if cfg.LogLevel != "" {
		level = cfg.LogLevel
	}

	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
}

func GetFlagAdress() string {
	return address.String()
}

func GetFlagHost() string {
	return host
}

func GetLogLevel() string {
	return level
}
