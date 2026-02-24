package opnsense

import (
	"log"

	"github.com/dreams-money/failover/config"
	"github.com/dreams-money/failover/scripts"
)

func (Router) Failover(cfg config.Config) error {
	newLeader, err := scripts.GetLeaderName(cfg)
	if err != nil {
		return err
	}

	logFailover(newLeader, cfg.NodeName)

	toggledInternalRoute := false
	if newLeader == cfg.NodeName {
		toggledInternalRoute, err = makePrimary(cfg)
	} else {
		toggledInternalRoute, err = makeReplica(newLeader, cfg)
	}
	if err != nil {
		return err
	}

	if toggledInternalRoute {
		err = reconfigureRoutes(cfg)
		if err != nil {
			return err
		}
		log.Println("Successfully reconfigured routes")
	}

	err = reconfigureWireguardService(cfg)
	if err != nil {
		return err
	}
	log.Println("Successfully reconfigured wireguard services")

	return nil
}

func makePrimary(cfg config.Config) (bool, error) {
	var err error

	toggledInternalRoute, err := enableVIPRoute(cfg.VIPRouteID, cfg)
	if err != nil {
		return toggledInternalRoute, err
	}
	log.Printf("Enabled VIP route.")

	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return toggledInternalRoute, err
	}
	log.Printf("Removed VIP from %v wireguard peers.\n", count)

	return toggledInternalRoute, nil
}

func makeReplica(leader string, cfg config.Config) (bool, error) {
	toggledInternalRoute, err := disableVIPRoute(cfg.VIPRouteID, cfg)
	if err != nil {
		return toggledInternalRoute, err
	}
	log.Printf("Disabled VIP route.")

	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return toggledInternalRoute, err
	}
	log.Printf("Removed VIP from %v wireguard peers.\n", count)

	err = addVIPToWireguardPeer(leader, cfg)
	if err != nil {
		return toggledInternalRoute, err
	}
	log.Printf("Added VIP to leader.\n")

	return toggledInternalRoute, nil
}

func logFailover(leader, thisNode string) {
	l := "Failing over. Leader is %v."
	if leader == thisNode {
		l += " I am the leader."
	}
	l += "\n"

	log.Printf(l, leader)
}
