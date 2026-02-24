package omada

import (
	"errors"

	"github.com/dreams-money/failover/config"
)

func (Router) SimpleCall(cfg config.Config) error {
	var peer config.Peer
	for _, peer = range cfg.Peers {
		break
	}
	_, err := getWireguardPeer(peer.WireguardPeerID, getRouterSpecificConfig(cfg))

	if err != nil {
		return errors.Join(errors.New("unable to access API"), err)
	}

	return nil
}
