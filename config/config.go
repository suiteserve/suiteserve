package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Http struct {
		Host            string `json:"host"`
		UserContentHost string `json:"user_content_host"`
		Port            string `json:"port"`
		TlsCertFile     string `json:"tls_cert_file"`
		TlsKeyFile      string `json:"tls_key_file"`
		PublicDir       string `json:"public_dir"`
	} `json:"http"`
	Storage struct {
		Db          string `json:"db"`
		UserContent struct {
			Dir       string `json:"dir"`
			MaxSizeMb int    `json:"max_size_mb"`
		} `json:"user_content"`
	} `json:"storage"`
}

func Load(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var c Config
	return &c, json.Unmarshal(b, &c)
}
