package userservice

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func GetEnv(key string) string {
	env := os.Getenv(key)
	if env == "" {
		log.Fatalf("%s is not well-set", key)
	}
	return env
}
