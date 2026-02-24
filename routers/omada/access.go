package omada

import (
	"errors"
	"log"
	"os"

	"github.com/dreams-money/failover/config"
	"github.com/dreams-money/failover/notifications"
)

func (Router) SetAuthorization(cfg config.Config) {
	err := RefreshAPIAccessOrFail(cfg)
	if err != nil {
		log.Println(err)
		notifications.PushMessage("No omada access! - " + cfg.NodeName)
		os.Exit(1)
		return
	}
}

func RefreshAPIAccessOrFail(cfg config.Config) error {
	err := (Router{}).SimpleCall(cfg)
	if err != nil {
		log.Println("No Omada:", err)
		return getTokensOrFail(getRouterSpecificConfig(cfg))
	}

	// Even if the API succeeded, we should refresh tokens to update timers on program start
	return RefreshTokens(getRouterSpecificConfig(cfg))
}

func getTokensOrFail(cfg Config) error {
	// As long as the program hasn't been down for 2 weeks, tokens should refresh.
	err := checkTokensTryRefresh(cfg)

	if err != nil {
		log.Println("Failed refresh:", err)
		return oauthOrFail(cfg)
	}

	return nil
}

func checkTokensTryRefresh(cfg Config) error {
	err := CheckAccessTokens(cfg)

	if errors.Is(err, NeedsTokenRefresh) {
		return RefreshTokens(cfg)
	}

	return err
}

func oauthOrFail(cfg Config) error {
	// Either the program has been down for 2 weeks or this is program intialization
	err := OAuth(cfg)
	if err != nil {
		log.Println("Failed to setup tokens.")
		return err
	}

	return nil
}
