package config

import "os"

type key string

const (
	Host      key = "HOST"
	Port          = "PORT"
	MongoHost     = "MONGO_HOST"
	MongoPort     = "MONGO_PORT"
	MongoUser     = "MONGO_USER"
	MongoPass     = "MONGO_PASS"
)

func Get(key key, defVal string) string {
	val, ok := os.LookupEnv(string(key))
	if !ok {
		return defVal
	} else {
		return val
	}
}
