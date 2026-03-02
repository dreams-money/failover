package config

import "fmt"

type Peer struct {
	DDNSAddress     string `json:"ddns"`
	WireguardPeerID string `json:"wireguard_peer_id"`
	Address         string `json:"address"`
	CheckHealth     bool   `json:"check_health"`
	ReplicaWeight   int64  `json:"replica_weight"`
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
