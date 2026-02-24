package opnsense

import "github.com/dreams-money/failover/config"

type Router struct {
}

func Make(cfg config.Config) (Router, error) {
	return Router{}, nil
}
