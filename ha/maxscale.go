package ha

import "github.com/dreams-money/failover/config"

type MaxScale struct{}

func (MaxScale) GetClusterStatus(cfg config.Config) (ClusterStatus, error) {
	return ClusterStatus{}, nil
}
