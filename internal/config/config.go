package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Http struct {
		Host            string `json:"host"`
		Port            uint16 `json:"port"`
		TlsCertFile     string `json:"tls_cert_file"`
		TlsKeyFile      string `json:"tls_key_file"`
		PublicDir       string `json:"public_dir"`
		UserContentHost string `json:"user_content_host"`
	} `json:"http"`
	Storage struct {
		UserContent struct {
			Dir       string `json:"dir"`
			MaxSizeMb int    `json:"max_size_mb"`
		} `json:"user_content"`
		Rethinkdb struct {
			Host     string `json:"host"`
			Port     uint16 `json:"port"`
			User     string `json:"user"`
			PassFile string `json:"pass_file"`
			Db       string `json:"db"`
		} `json:"rethinkdb"`
	} `json:"storage"`
}

func Load(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
