package orderservice

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

// GetEnv to get config from env file
func GetEnv(key string) string {
	env := os.Getenv(key)
	if env == "" {
		log.Fatalf("%s is not well-set", key)
	}
	return env
}

// CopyStructValue copy identical struct value from source to dest
// will return error if source and dest is not identical
func CopyStructValue(source, dest interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, dest); err != nil {
		return err
	}

	return nil
}
