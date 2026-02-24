package omada

import (
	"encoding/json"
	"log"

	"github.com/dreams-money/failover/config"
)

type Config struct {
	OmadaAddress           string
	OmadaPort              string `json:"omada_port"`
	OmadaClientID          string `json:"omada_client_id"`
	OmadaOAuthClientID     string `json:"omada_oauth_client_id"`
	OmadaOAuthClientSecret string `json:"omada_oauth_client_secret"`
	OmadaSiteID            string `json:"omada_site_id"`
}

func getRouterSpecificConfig(cfg config.Config) Config {
	var specificConfig Config
	err := json.Unmarshal(cfg.RouterConfig, &specificConfig)
	if err != nil {
		log.Panic(err)
	}

	specificConfig.OmadaAddress = cfg.RouterAddress

	return specificConfig
}
