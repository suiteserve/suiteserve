package config

import "os"

type Key string

const (
	Host            Key = "HOST"
	Port                = "PORT"
	MongoHost           = "MONGO_HOST"
	MongoPort           = "MONGO_PORT"
	MongoReplicaSet     = "MONGO_REPLICA_SET"
	MongoUser           = "MONGO_USER"
	MongoPass           = "MONGO_PASS"
)

func Get(key Key, defVal string) string {
	val, ok := os.LookupEnv(string(key))
	if !ok {
		return defVal
	} else {
		return val
	}
}
