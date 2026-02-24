package config

import "fmt"

type Peer struct {
	WireguardPeerID string `json:"wireguard_peer_id"`
	Address         string `json:"address"`
	CheckHealth     bool   `json:"check_health"`
}

type Peers map[string]Peer

func (p *Peers) GetPeer(key string) (*Peer, error) {
	for peer, config := range *p {
		if peer == key {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("peer not found: %v", key)
}
