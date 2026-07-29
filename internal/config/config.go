package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

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

var FlagAddress = Address{
	Host: "localhost",
	Port: 8080,
}
var FlagHost string

func ParseFlags() {
	flag.Var(&FlagAddress, "a", "host and port to start server")
	flag.StringVar(&FlagHost, "b", "http://localhost:8080/", "base address to result")
	flag.Parse()

	if !strings.HasSuffix(FlagHost, "/") {
		FlagHost = "/"
	}
}
