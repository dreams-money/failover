package opnsense

import (
	"log"

	"github.com/dreams-money/failover/config"
)

func (Router) SimpleCall(cfg config.Config) error {
	var peer config.Peer
	for _, peer = range cfg.Peers {
		break
	}

	_, err := getWireguardPeer(peer.WireguardPeerID, cfg)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
