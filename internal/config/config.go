package config

import (
	"errors"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

var ErrFormatNotCorrect = errors.New("need address in a form host:port")

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
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
		return ErrFormatNotCorrect
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
var filePath string = "./storage.json"
var databaseDSN string

func ParseConfig() {
	flag.Var(&address, "a", "host and port to start server")
	flag.StringVar(&host, "b", "http://localhost:8080/", "base address to result")
	flag.StringVar(&filePath, "f", "./storage.txt", "file storage path")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL connection string")
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

	if cfg.FileStoragePath != "" {
		filePath = cfg.FileStoragePath
	}
	if cfg.DatabaseDSN != "" {
		databaseDSN = cfg.DatabaseDSN
	}

	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
}

func GetFlagAddress() string {
	return address.String()
}

func GetFlagHost() string {
	return host
}

func GetLogLevel() string {
	return level
}

func GetFileStoragePath() string {
	return filePath
}

func GetDatabaseDSN() string {
	return databaseDSN
}
