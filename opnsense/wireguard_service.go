package opnsense

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dreams-money/opnsense-failover/config"
)

func reconfigureWireguardService(cfg config.Config) error {
	url := "https://%v/api/wireguard/service/reconfigure"
	url = fmt.Sprintf(url, cfg.OpnSenseAddress)

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
