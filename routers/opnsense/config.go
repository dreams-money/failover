package opnsense

import (
	"encoding/json"
	"log"

	"github.com/dreams-money/failover/config"
)

type Config struct {
	OpnSenseApiKey    string `json:"opnsense_api_key"`
	OpnSenseApiSecret string `json:"opnsense_api_secret"`
}

func getRouterSpecificConfig(cfg config.Config) Config {
	var specificConfig Config
	err := json.Unmarshal(cfg.RouterConfig, &specificConfig)
	if err != nil {
		log.Panic(err)
	}

	return specificConfig
}
