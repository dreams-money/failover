package opnsense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/dreams-money/opnsense-failover/config"
)

type Route struct {
	ID          string `json:"uuid"`
	Network     string `json:"network"`
	Gateway     string `json:"gateway"`
	Description string `json:"descr"`
	Disabled    string `json:"disabled"`
}

func getRoute(uuid string, cfg config.Config) (Route, error) {
	url := "https://%v/api/routes/routes/searchroute"
	url = fmt.Sprintf(url, cfg.OpnSenseAddress)
	route := Route{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return route, err
	}

	req.Header.Add("Authorization", Authorization)

	resp, err := httpClient.Do(req)
	if err != nil {
		return route, err
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return route, err
	}

	if resp.StatusCode != 200 {
		e := "opnsense GET routes api request failed, status: %v, msg: %v"
		return route, fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type jsonResponse struct {
		Rows     []Route `json:"rows"`
		RowCount int     `json:"rowCount"`
		Total    int     `json:"total"`
		Current  int     `json:"current"`
	}
	jr := jsonResponse{}

	err = json.Unmarshal(respBodyBytes, &jr)
	if err != nil {
		return route, err
	}

	if jr.RowCount == 0 {
		return route, fmt.Errorf("routes get, api response: no routes found")
	} else if jr.Total > jr.RowCount {
		log.Println("routes get, api response: more rows than expected")
	}

	for _, row := range jr.Rows {
		if row.ID == uuid {
			return row, nil
		}
	}

	return route, fmt.Errorf("routes get, api response: requested route not found")
}

func editRoute(route Route, cfg config.Config) error {
	url := "https://%v/api/routes/routes/setroute/%v"
	url = fmt.Sprintf(url, cfg.OpnSenseAddress, route.ID)

	type Payload struct {
		Route `json:"route"`
	}
	payload := Payload{Route: route}
	payloadJS, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(payloadJS)

	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
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
		e := "opnsense POST route api request failed, status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type jsonResponse struct {
		Result string `json:"result"`
	}
	jr := jsonResponse{}

	err = json.Unmarshal(respBodyBytes, &jr)
	if err != nil {
		return err
	}

	if jr.Result != "saved" {
		return fmt.Errorf("opnsense POST route api response, msg: %v", jr.Result)
	}

	return nil
}

func reconfigureRoutes(cfg config.Config) error {
	url := "https://%v/api/routes/routes/reconfigure"
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
		e := "opnsense POST route service reconfigure request failed, status: %v, msg: %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type OpnSenseResponse struct {
		Status string `json:"status"`
	}
	osr := OpnSenseResponse{}

	err = json.Unmarshal(respBodyBytes, &osr)
	if err != nil {
		return err
	}

	if osr.Status != "ok" {
		e := "opnsense route service reconfigure responded not ok, msg: %v"
		return fmt.Errorf(e, string(respBodyBytes))
	}

	return nil
}
