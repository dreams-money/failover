package omada

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"

	"github.com/dreams-money/failover/persistence"
)

type Peer struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        bool     `json:"status"`
	InterfaceID   string   `json:"interfaceId"`
	InterfaceName string   `json:"interfaceName"`
	PublicKey     string   `json:"publicKey"`
	EndPoint      string   `json:"endPoint"`
	EndPointPort  int      `json:"endPointPort"`
	ExistDomain   bool     `json:"existDomain"`
	AllowAddress  []string `json:"allowAddress"`
	PresharedKey  string   `json:"presharedKey"`
	KeepAlive     int      `json:"keepAlive"`
}

func (p *Peer) hasVIP(vip string) bool {
	return slices.Contains(p.AllowAddress, vip)
}

func getWireguardPeer(peerID string, cfg Config) (Peer, error) {
	url := "https://%v:%v/openapi/v1/"
	url += cfg.OmadaClientID
	url += "/sites/" + cfg.OmadaSiteID
	url += "/vpn/wireguard-peers?pageSize=10&page=1"

	url = fmt.Sprintf(url, cfg.OmadaAddress, cfg.OmadaPort)
	peer := Peer{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return peer, err
	}

	accessToken, err := persistence.GetAccessToken()
	if err != nil {
		return peer, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "AccessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return peer, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return peer, err
	}

	if resp.StatusCode != 200 {
		e := "request failed to remove VIP from peer, http status: %v, msg: %v"
		return peer, fmt.Errorf(e, resp.StatusCode, string(bodyBytes))
	}

	type Result struct {
		TotalRows     int    `json:"totalRows"`
		CurrentSize   int    `json:"currentSize"`
		Data          []Peer `json:"data"`
		SupportDomain bool   `json:"supportDomain"`
	}

	type getWireguardPeerResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
		Result    `json:"result"`
	}

	jsonResponse := getWireguardPeerResponse{} // jsonResponse
	err = json.Unmarshal(bodyBytes, &jsonResponse)
	if err != nil {
		return peer, err
	}

	if jsonResponse.ErrorCode < 0 {
		return peer, fmt.Errorf("from omada - (%v) %v", jsonResponse.ErrorCode, jsonResponse.Message)
	}

	if jsonResponse.Result.TotalRows > jsonResponse.Result.CurrentSize {
		log.Printf("Much more wireguard peers than we expected: %v", jsonResponse.Result.TotalRows)
	}

	for _, peer := range jsonResponse.Data {
		if peer.ID == peerID {
			return peer, nil
		}
	}

	return Peer{}, fmt.Errorf("peer not found! - %v", peerID)
}

func editWireguardPeer(peer Peer, cfg Config) error {
	url := "https://%v:%v/openapi/v1/"
	url += cfg.OmadaClientID
	url += "/sites/" + cfg.OmadaSiteID
	url += "/vpn/wireguard-peers/" + peer.ID

	url = fmt.Sprintf(url, cfg.OmadaAddress, cfg.OmadaPort)

	peerBytes, err := json.Marshal(peer)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(peerBytes))
	if err != nil {
		return err
	}

	accessToken, err := persistence.GetAccessToken()
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "AccessToken="+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		e := "request failed to remove VIP from peer, http status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, string(bodyBytes))
	}

	type peerEditResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
	}

	jr := peerEditResponse{} // Json Response
	err = json.Unmarshal(bodyBytes, &jr)
	if err != nil {
		return err
	}

	if jr.ErrorCode != 0 {
		return fmt.Errorf("openAPI Error %v: %v", jr.ErrorCode, jr.Message)
	}

	return nil
}
