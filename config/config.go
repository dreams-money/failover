package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	NodeName                string          `json:"node_name"`
	RouterType              string          `json:"router_type"`
	RouterAddress           string          `json:"router_address"`
	RouterConfig            json.RawMessage `json:"router_config"`
	AppPort                 string          `json:"app_port"`
	Peers                   `json:"peers"`
	HeartBeatIntervalString string `json:"heartbeat_interval"`
	HeartBeatInterval       time.Duration
	VIPAddress              string `json:"vip_address"`
	VIPRouteID              string `json:"vip_route_id"`
	LeaderScript            string `json:"leader_script"`
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
