package routers

import (
	"errors"

	"github.com/dreams-money/failover/config"
	"github.com/dreams-money/failover/routers/omada"
	"github.com/dreams-money/failover/routers/opnsense"
)

type Router interface {
	SetAuthorization(config.Config)
	SimpleCall(config.Config) error
	Failover(config.Config) error
}

func Make(cfg config.Config) (Router, error) {
	switch cfg.RouterType {
	case "opnsense":
		return opnsense.Make(cfg)
	case "omada":
		return omada.Make(cfg)
	default:
		return nil, errors.New("Unknown router type - " + cfg.RouterType)
	}
}
