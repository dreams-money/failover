package opnsense

import (
	"errors"
	"strings"

	"github.com/dreams-money/failover/config"
)

func disableVIPRoute(routeID string, cfg config.Config) (bool, error) {
	route, err := getRoute(routeID, cfg)
	if err != nil {
		return false, err
	}

	if route.IsDisabled() {
		return false, nil
	}

	route.Disabled = "1"

	return true, editRoute(route, cfg)
}

func enableVIPRoute(routeID string, cfg config.Config) (bool, error) {
	route, err := getRoute(routeID, cfg)
	if err != nil {
		return false, err
	}

	if route.IsEnabled() {
		return false, nil
	}

	route.Disabled = "0"

	return true, editRoute(route, cfg)
}

func removeVIPFromWireguardPeer(peerID string, cfg config.Config) error {
	peer, err := getWireguardPeer(peerID, cfg)
	if err != nil {
		return err
	}

	if !peer.hasVIP(cfg.VIPAddress) {
		return nil
	}

	nonVIPAddresses := []string{}
	addresses := strings.Split(peer.TunnelAddress, ",")
	for _, address := range addresses {
		if address != cfg.VIPAddress {
			nonVIPAddresses = append(nonVIPAddresses, address)
		}
	}
	peer.TunnelAddress = strings.Join(nonVIPAddresses, ",")

	return editWireguardPeer(peer, cfg)
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

	peer, err := getWireguardPeer(peerCfg.WireguardPeerID, cfg)
	if err != nil {
		return err
	}

	if peer.hasVIP(cfg.VIPAddress) {
		return nil
	}

	peer.TunnelAddress += "," + cfg.VIPAddress

	return editWireguardPeer(peer, cfg)
}
