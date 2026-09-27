package config

import (
	"github.com/joho/godotenv"
	"os"
	"sync"
)

var once sync.Once

func initEnv() {
	// Load local configuration when present. Environment variables remain the
	// source of truth, so the project can build and run tests without secrets.
	_ = godotenv.Load("config.env")
}

func LoadStrFromEnv(key string) string {
	once.Do(func() {
		initEnv()
	})
	return os.Getenv(key)
}
