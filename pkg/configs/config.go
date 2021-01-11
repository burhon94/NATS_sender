package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Prefix string `json:"prefix"`
	Nats struct {
		URL       string `json:"url"`
		ClusterID string `json:"clusterID"`
		ClientID  string `json:"clientID"`
	} `json:"nats"`
}

func (receiver *Config) Init() (err error) {
	conf, err := os.Open("config.json")
	defer func() {
		err = conf.Close()
	}()
	if err != nil {
		return
	}
	all, err := ioutil.ReadAll(conf)
	if err != nil {
		return
	}
	err = json.Unmarshal(all, receiver)
	return
}
