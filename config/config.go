package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Config struct {
	Http struct {
		Host            string `json:"host"`
		UserContentHost string `json:"user_content_host"`
		Port            string `json:"port"`
		TlsCertFile     string `json:"tls_cert_file"`
		TlsKeyFile      string `json:"tls_key_file"`
		FrontendDir     string `json:"frontend_dir"`
	} `json:"http"`

	Storage struct {
		UserContent struct {
			Dir       string `json:"dir"`
			MaxSizeMb int    `json:"max_size_mb"`
		} `json:"user_content"`

		BuntDb *struct {
			File string `json:"file"`
		} `json:"buntdb"`

		MongoDb *struct {
			Host     string `json:"host"`
			Port     uint16 `json:"port"`
			Rs       string `json:"rs"`
			Db       string `json:"db"`
			User     string `json:"user"`
			PassFile string `json:"pass_file"`
		} `json:"mongodb"`
	} `json:"storage"`
}

func New(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json: %v", err)
	}

	return &cfg, err
}

func ReadFile(filename string) string {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("read file from config: %v\n", err)
	}
	return string(bytes.TrimRight(b, "\r\n"))
}
