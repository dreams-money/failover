package opnsense

import (
	"log"
	"time"

	"github.com/dreams-money/opnsense-failover/config"
	"github.com/dreams-money/opnsense-failover/etcd"
)

var lastLeader string

func Initialize(cfg config.Config) error {
	var err error
	waitTime := time.Duration(3 * time.Second)

	// We need to wait for ETCD to load
	time.Sleep(waitTime)
	for {
		lastLeader, err = etcd.GetLeaderName(cfg)
		if err == nil {
			break
		}
		log.Println("Waiting for ETCD to load", err)

		time.Sleep(waitTime)
	}

	log.Print("Loaded leader", lastLeader)

	return err
}

func Failover(cfg config.Config) error {
	newLeader, err := etcd.GetLeaderName(cfg)
	if err != nil {
		return err
	}

	logFailover(newLeader, cfg.NodeName)

	needsRouteChange := needsRouteChange(newLeader, cfg.NodeName)

	lastLeader = newLeader

	if newLeader == cfg.NodeName {
		err = makePrimary(cfg)
	} else {
		err = makeReplica(newLeader, cfg, needsRouteChange)
	}
	if err != nil {
		return err
	}

	if needsRouteChange {
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

func makeReplica(leader string, cfg config.Config, needsRouteChange bool) error {
	var err error

	if needsRouteChange {
		err = disableVIPRoute(cfg.VIPRouteID, cfg)
		if err != nil {
			return err
		}
		log.Printf("Disabled VIP route.")
	}

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

func needsRouteChange(newLeader, thisNode string) bool {
	return newLeader == thisNode || lastLeader == thisNode
}
