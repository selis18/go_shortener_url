package config

import (
	"errors"
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

var ErrFormatNotCorrect = errors.New("need address in a form host:port")

type Config struct {
	ServerAddress       string        `env:"SERVER_ADDRESS"`
	BaseURL             string        `env:"BASE_URL"`
	LogLevel            string        `env:"LOG_LEVEL"`
	FileStoragePath     string        `env:"FILE_STORAGE_PATH"`
	DatabaseDSN         string        `env:"DATABASE_DSN"`
	SecretKey           string        `env:"SECRET_KEY"`
	DeleteBatchSize     int           `env:"DELETE_BATCH_SIZE"`
	DeleteFlushInterval time.Duration `env:"DELETE_FLUSH_INTERVAL"`
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
var filePath string
var databaseDSN string
var secretKey string
var deleteBatchSize = 1000
var deleteFlushInterval = 25 * time.Millisecond

func ParseConfig() {
	flag.Var(&address, "a", "host and port to start server")
	flag.StringVar(&host, "b", "http://localhost:8080/", "base address to result")
	flag.StringVar(&filePath, "f", "", "file storage path")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL connection string")
	flag.IntVar(&deleteBatchSize, "delete-batch-size", 1000, "maximum number of URLs in a deletion batch (positive)")
	flag.DurationVar(&deleteFlushInterval, "delete-flush-interval", 25*time.Millisecond, "deletion buffer flush interval (positive duration, e.g. 25ms or 1s)")
	flag.Parse()

	cfg := Config{
		DeleteBatchSize:     deleteBatchSize,
		DeleteFlushInterval: deleteFlushInterval,
	}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DeleteBatchSize <= 0 || cfg.DeleteFlushInterval <= 0 {
		log.Fatal("delete batch size and flush interval must be positive")
	}
	deleteBatchSize = cfg.DeleteBatchSize
	deleteFlushInterval = cfg.DeleteFlushInterval

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

	if cfg.SecretKey != "" {
		secretKey = cfg.SecretKey
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

func GetSecretKey() string {
	return secretKey
}

func GetDeleteBatchSize() int {
	return deleteBatchSize
}

func GetDeleteFlushInterval() time.Duration {
	return deleteFlushInterval
}
