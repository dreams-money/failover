package opnsense

import (
	"log"

	"github.com/dreams-money/opnsense-failover/config"
	"github.com/dreams-money/opnsense-failover/etcd"
)

func Failover(cfg config.Config) error {
	leader, err := etcd.GetLeaderName(cfg)
	if err != nil {
		return err
	}

	logFailover(leader, cfg.NodeName)

	if leader == cfg.NodeName {
		err = makePrimary(cfg)
	} else {
		err = makeReplica(leader, cfg)
	}
	if err != nil {
		return err
	}

	err = reconfigureRoutes(cfg)
	if err != nil {
		return err
	}
	log.Println("Successfully reconfigured routes")

	err = reconfigureWireguardService(cfg)
	if err != nil {
		return err
	}
	log.Println("Successfully reconfigured wireguard services")

	return nil
}

func makePrimary(cfg config.Config) error {
	var err error

	err = enableVIPRoute(cfg.VIPRouteID, cfg)
	if err != nil {
		return err
	}
	log.Printf("Enabled VIP route.")

	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return err
	}
	log.Printf("Removed VIP from %v wireguard peers.\n", count)

	return nil
}

func makeReplica(leader string, cfg config.Config) error {
	var err error

	err = disableVIPRoute(cfg.VIPRouteID, cfg)
	if err != nil {
		return err
	}
	log.Printf("Disabled VIP route.")

	count, err := removeVIPFromWireguardPeers(cfg)
	if err != nil {
		return err
	}
	log.Printf("Removed VIP from %v wireguard peers.\n", count)

	err = addVIPToWireguardPeer(leader, cfg)
	if err != nil {
		return err
	}
	log.Printf("Added VIP to leader.\n")

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
