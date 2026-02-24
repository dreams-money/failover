package omada

import (
	"errors"

	"github.com/dreams-money/failover/config"
)

func removeVIPFromWireguardPeer(peerID string, cfg config.Config) error {
	peer, err := getWireguardPeer(peerID, getRouterSpecificConfig(cfg))
	if err != nil {
		return err
	}

	if !peer.hasVIP(cfg.VIPAddress) {
		return nil
	}

	nonVIPAddresses := []string{}
	for _, address := range peer.AllowAddress {
		if address != cfg.VIPAddress {
			nonVIPAddresses = append(nonVIPAddresses, address)
		}
	}
	peer.AllowAddress = nonVIPAddresses

	return editWireguardPeer(peer, getRouterSpecificConfig(cfg))
}

func removeVIPFromWireguardPeers(cfg config.Config) (int, error) {
	var err error
	var count int

	for peer, peerConfig := range cfg.Peers {
		err = removeVIPFromWireguardPeer(peerConfig.WireguardPeerID, cfg)
		if err != nil {
			return count, errors.Join(err, errors.New(peer))
		}
		count++
	}

	return count, nil
}

func addVIPToWireguardPeer(leader string, cfg config.Config) error {
	peerCfg, err := cfg.Peers.GetPeer(leader)
	if err != nil {
		return err
	}

	peer, err := getWireguardPeer(peerCfg.WireguardPeerID, getRouterSpecificConfig(cfg))
	if err != nil {
		return err
	}

	if peer.hasVIP(cfg.VIPAddress) {
		return nil
	}

	peer.AllowAddress = append(peer.AllowAddress, cfg.VIPAddress)

	return editWireguardPeer(peer, getRouterSpecificConfig(cfg))
}
