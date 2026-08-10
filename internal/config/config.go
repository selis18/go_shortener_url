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
var host string

func ParseConfig() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerAddress != "" && cfg.BaseURL == "" {
		addressArr := strings.Split(cfg.ServerAddress, ":")
		address.Host = addressArr[0]
		address.Port, err = strconv.Atoi(addressArr[1])
	} else {
		flag.Var(&address, "a", "host and port to start server")
		flag.Parse()
	}

	if cfg.BaseURL != "" && cfg.ServerAddress == "" {
		host = cfg.BaseURL
	} else {
		flag.StringVar(&host, "b", "http://localhost:8080/", "base address to result")
		if !strings.HasSuffix(host, "/") {
			host += "/"
		}
		flag.Parse()
	}

	if cfg.BaseURL != "" && cfg.ServerAddress != "" {
		addressArr := strings.Split(cfg.ServerAddress, ":")
		address.Host = addressArr[0]
		address.Port, err = strconv.Atoi(addressArr[1])
		host = cfg.BaseURL
	} else {
		flag.Var(&address, "a", "host and port to start server")
		flag.StringVar(&host, "b", "http://localhost:8080/", "base address to result")
		if !strings.HasSuffix(host, "/") {
			host += "/"
		}
		flag.Parse()
	}

}

func GetFlagAdress() string {
	return address.String()
}

func GetFlagHost() string {
	return host
}
