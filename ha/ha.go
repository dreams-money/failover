package ha

import "github.com/dreams-money/failover/config"

type HighAvailability interface {
	GetClusterStatus(config.Config) (ClusterStatus, error)
}
