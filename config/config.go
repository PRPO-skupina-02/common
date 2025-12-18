package config

import (
	"fmt"
	"os"
)

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("Env var %s must be set", key))
	}
	return val
}

func GetEnvDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return def
}
