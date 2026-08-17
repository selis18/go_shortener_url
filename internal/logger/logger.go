package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	//спарсили уровень логирования
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	//новая конфигурация логера
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	//logger по конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	//синглтон
	Log = zl
	return nil
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nowTime := time.Now()
		h.ServeHTTP(w, r)
		reqTime := time.Since(nowTime)

		Log.Info("HTTP request",
			zap.String("URI", r.Host),
			zap.String("Method", r.Method),
			zap.String("Request time", reqTime.String()),
		)
	})
}
