package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	NodeName                string `json:"node_name"`
	OpnSenseAddress         string `json:"opnsense_address"`
	OpnSenseApiKey          string `json:"opnsense_api_key"`
	OpnSenseApiSecret       string `json:"opnsense_api_secret"`
	AppPort                 string `json:"app_port"`
	Peers                   `json:"peers"`
	HeartBeatIntervalString string `json:"heartbeat_interval"`
	HeartBeatInterval       time.Duration
	ETCDAddress             string `json:"etcd_address"`
	ETCDPort                string `json:"etcd_port"`
	VIPAddress              string `json:"vip_address"`
}

func LoadProgramConfiguration() (Config, error) {
	config := Config{}
	fileBytes, err := os.ReadFile("config.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return config, err
	}

	config.HeartBeatInterval, err = time.ParseDuration(config.HeartBeatIntervalString)
	if err != nil {
		return config, err
	}

	return config, err
}
