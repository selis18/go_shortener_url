package main

import (
	"testing"

	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	config := &config.Address{}
	config.Set("http://localhost:8081/")
	assert.NotPanics(t, func() {})
}
