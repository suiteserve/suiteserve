package config

import (
	"io/ioutil"
	"log"
	"os"
)

type Key string

const (
	Host            Key = "HOST"
	Port                = "PORT"
	TlsCertFile         = "TLS_CERT_FILE"
	TlsKeyFile          = "TLS_KEY_FILE"
	MongoHost           = "MONGO_HOST"
	MongoPort           = "MONGO_PORT"
	MongoReplicaSet     = "MONGO_REPLICA_SET"
	MongoUser           = "MONGO_USER"
	MongoPassFile       = "MONGO_PASS_FILE"
)

func Env(key Key, defVal string) string {
	val, ok := os.LookupEnv(string(key))
	if !ok {
		return defVal
	}
	return val
}

func EnvFile(key Key, defVal string) string {
	val, ok := os.LookupEnv(string(key))
	if !ok {
		return defVal
	}
	b, err := ioutil.ReadFile(val)
	if err != nil {
		log.Printf("read env file: %v\n", err)
		return defVal
	}
	return string(b)
}
