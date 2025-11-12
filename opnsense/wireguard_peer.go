package opnsense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/dreams-money/opnsense-failover/config"
	"github.com/dreams-money/opnsense-failover/notifications"
)

type Peer struct {
	ID            string `json:"uuid"`
	Enabled       string `json:"enabled"`
	Name          string `json:"name"`
	PublicKey     string `json:"pubkey"`
	PresharedKey  string `json:"psk"`
	TunnelAddress string `json:"tunneladdress"`
	ServerAddress string `json:"serveraddress"`
	ServerPort    string `json:"serverport"`
	Servers       string `json:"servers"`
	KeepAlive     string `json:"keepalive"`
}

func getWireguardPeer(peerID string, cfg config.Config) (Peer, error) {
	url := "https://%v/api/wireguard/client/search_client"
	url = fmt.Sprintf(url, cfg.OpnSenseAddress)
	peer := Peer{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return peer, err
	}

	req.Header.Add("Authorization", Authorization)

	resp, err := httpClient.Do(req)
	if err != nil {
		return peer, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return peer, err
	}

	respBody := string(bodyBytes)
	if resp.StatusCode != 200 {
		e := "opnsense GET wireguard peer api request failed, status: %v, msg: %v"
		return peer, fmt.Errorf(e, resp.StatusCode, respBody)
	}

	type jsonResponse struct {
		Rows     []Peer `json:"rows"`
		RowCount int    `json:"rowCount"`
		Total    int    `json:"total"`
		Current  int    `json:"current"`
	}

	jr := jsonResponse{}
	err = json.Unmarshal(bodyBytes, &jr)
	if err != nil {
		return peer, err
	}

	// I don't ever expect >10 peers
	if jr.Total > jr.RowCount {
		log.Printf("Large row count from opnsense: %v", jr.Total)
		notifications.PushMessage("Large row count from opnsense. Check logs.")
	}

	for _, row := range jr.Rows {
		if row.ID == peerID {
			return row, nil
		}
	}

	return peer, fmt.Errorf("opnsense GET wireguard peer not found - %v", peerID)
}

func editWireguardPeer(peer Peer, cfg config.Config) error {
	url := "https://%v/api/wireguard/client/set_client/%v"
	url = fmt.Sprintf(url, cfg.OpnSenseAddress, peer.ID)

	type jsonRequest struct {
		Peer `json:"client"`
	}
	jsRes := jsonRequest{Peer: peer}

	reqJsonBytes, err := json.Marshal(jsRes)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJsonBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", Authorization)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBody := string(bodyBytes)
	if resp.StatusCode != 200 {
		e := "opnsense POST wireguard peer api request failed, status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, respBody)
	}

	type jsonResponse struct {
		Result string `json:"result"`
	}
	jr := jsonResponse{}

	err = json.Unmarshal(bodyBytes, &jr)
	if err != nil {
		return err
	}

	if jr.Result != "saved" {
		return fmt.Errorf("opnsense POST wireguard peer api response, msg: %v", jr.Result)
	}

	return nil
}

func (p *Peer) hasVIP(vip string) bool {
	addresses := strings.Split(p.TunnelAddress, ",")
	return slices.Contains(addresses, vip)
}
