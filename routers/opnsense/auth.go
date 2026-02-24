package opnsense

import (
	"encoding/base64"

	"github.com/dreams-money/failover/config"
)

var Authorization string

func (Router) SetAuthorization(cfg config.Config) {
	opnsenseAuth := getRouterSpecificConfig(cfg)

	auth := opnsenseAuth.OpnSenseApiKey + ":" + opnsenseAuth.OpnSenseApiSecret
	auth = base64.StdEncoding.EncodeToString([]byte(auth))

	Authorization = "Basic " + auth
}
