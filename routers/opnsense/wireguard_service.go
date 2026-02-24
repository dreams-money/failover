package opnsense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dreams-money/failover/config"
)

func setWireguardService(cfg config.Config) error {
	url := "https://%v/api/wireguard/general/set"
	url = fmt.Sprintf(url, cfg.RouterAddress)

	payload := "{general: {enabled: \"1\"}}"
	buffer := bytes.NewBuffer([]byte(payload))

	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", Authorization)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		e := "opnsense SET wireguard service request failed, status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type OpnSenseResponse struct {
		Result string `json:"result"`
	}
	osr := OpnSenseResponse{}

	err = json.Unmarshal(respBodyBytes, &osr)
	if err != nil {
		return err
	}

	if osr.Result != "saved" {
		e := "opnsense wireguard service SET responded not saved, msg: %v"
		return fmt.Errorf(e, string(respBodyBytes))
	}

	return nil
}

func reconfigureWireguardService(cfg config.Config) error {
	url := "https://%v/api/wireguard/service/reconfigure"
	url = fmt.Sprintf(url, cfg.RouterAddress)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", Authorization)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		e := "opnsense POST wireguard service reconfigure request failed, status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type OpnSenseResponse struct {
		Result string `json:"result"`
	}
	osr := OpnSenseResponse{}

	err = json.Unmarshal(respBodyBytes, &osr)
	if err != nil {
		return err
	}

	if osr.Result != "ok" {
		e := "opnsense wireguard service reconfigure responded not ok, msg: %v"
		return fmt.Errorf(e, string(respBodyBytes))
	}

	return nil
}
