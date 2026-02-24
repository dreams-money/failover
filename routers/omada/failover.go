package omada

import (
	"log"

	"github.com/dreams-money/failover/config"
	"github.com/dreams-money/failover/scripts"
)

func (Router) Failover(cfg config.Config) error {
	leader, err := scripts.GetLeaderName(cfg)
	if err != nil {
		return err
	}

	logFailover(leader, cfg.NodeName)

	if leader == cfg.NodeName {
		return isPrimary(cfg)
	}

	return isReplica(leader, cfg)
}

func isPrimary(cfg config.Config) error {
	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return err
	}

	log.Printf("Removed VIP from %v peers.\n", count)

	return nil
}

func isReplica(leader string, cfg config.Config) error {
	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return err
	}

	log.Printf("Removed VIP from %v peers.\n", count)

	err = addVIPToWireguardPeer(leader, cfg)
	if err != nil {
		return err
	}

	log.Printf("Successfully added VIP to %v peer.\n", leader)

	return nil
}

func logFailover(leader, thisNode string) {
	l := "Failing over. Leader is %v."
	if leader == thisNode {
		l += " I am the leader."
	}
	l += "\n"

	log.Printf(l, leader)
}
