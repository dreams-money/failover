package omada

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dreams-money/failover/console"
	"github.com/dreams-money/failover/persistence"
)

type setupContext struct {
	cfg               Config
	user              string
	password          string
	crsfToken         string
	sessionID         string
	authorizationCode string
}

func OAuth(cfg Config) error {
	var err error
	ctx := setupContext{
		cfg: cfg,
	}

	ctx.user, ctx.password, err = console.Credentials()
	if err != nil {
		return err
	}

	err = login(&ctx)
	if err != nil {
		return err
	}
	log.Println("Logged in successfully.")

	err = authCode(&ctx)
	if err != nil {
		return err
	}
	log.Println("Recieved authorization code.")

	err = getTokens(&ctx)
	if err != nil {
		return err
	}
	log.Println("Successfully set tokens.")

	return nil
}

func login(ctx *setupContext) error {
	url := "https://%v:%v/openapi/authorize/login?"
	url = fmt.Sprintf(url, ctx.cfg.OmadaAddress, ctx.cfg.OmadaPort)
	url += "client_id="
	url += ctx.cfg.OmadaOAuthClientID
	url += "&omadac_id="
	url += ctx.cfg.OmadaClientID

	type LoginPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	payload := LoginPayload{Username: ctx.user, Password: ctx.password}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(payloadBytes)

	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

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
		e := "response failed (%v) %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type Result struct {
		CSRFToken string `json:"csrfToken"`
		SessionID string `json:"sessionId"`
	}

	type OmadaResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
		Result    `json:"result"`
	}

	or := OmadaResponse{}
	err = json.Unmarshal(respBodyBytes, &or)
	if err != nil {
		return err
	}

	if or.ErrorCode < 0 {
		e := "omada error (%v): %v"
		return fmt.Errorf(e, or.ErrorCode, or.Message)
	}

	ctx.crsfToken = or.Result.CSRFToken
	ctx.sessionID = or.Result.SessionID

	return nil
}

func authCode(ctx *setupContext) error {
	url := "https://%v:%v/openapi/authorize/code?"
	url = fmt.Sprintf(url, ctx.cfg.OmadaAddress, ctx.cfg.OmadaPort)
	url += "client_id="
	url += ctx.cfg.OmadaOAuthClientID
	url += "&omadac_id="
	url += ctx.cfg.OmadaClientID
	url += "&response_type=code"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Csrf-Token", ctx.crsfToken)
	req.Header.Add("Cookie", "TPOMADA_SESSIONID="+ctx.sessionID)

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
		e := "response failed (%v) %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type OmadaResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
		Result    string `json:"result"`
	}
	or := OmadaResponse{}

	err = json.Unmarshal(respBodyBytes, &or)
	if err != nil {
		return err
	}

	if or.ErrorCode < 0 {
		e := "omada error (%v): %v"
		return fmt.Errorf(e, or.ErrorCode, or.Message)
	}

	ctx.authorizationCode = or.Result

	return nil
}

func getTokens(ctx *setupContext) error {
	url := "https://%v:%v/openapi/authorize/token?"
	url = fmt.Sprintf(url, ctx.cfg.OmadaAddress, ctx.cfg.OmadaPort)
	url += "grant_type=authorization_code"
	url += "&code=" + ctx.authorizationCode

	payload := "{\"client_id\":\"%v\",\"client_secret\":\"%v\"}"
	payload = fmt.Sprintf(payload, ctx.cfg.OmadaOAuthClientID, ctx.cfg.OmadaOAuthClientSecret)
	buffer := bytes.NewBuffer([]byte(payload))

	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

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
		e := "response failed (%v) %v"
		return fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type Result struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		TokenType    string `json:"tokenType"`
		ExpiresIn    int    `json:"expiresIn"`
	}

	type OmadaResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
		Result    `json:"result"`
	}

	or := OmadaResponse{}
	err = json.Unmarshal(respBodyBytes, &or)
	if err != nil {
		return err
	}

	if or.ErrorCode < 0 {
		e := "omada error (%v): %v"
		return fmt.Errorf(e, or.ErrorCode, or.Message)
	}

	err = persistence.SetAccessToken(or.AccessToken)
	if err != nil {
		return err
	}
	err = persistence.SetRefreshToken(or.RefreshToken)
	if err != nil {
		return err
	}
	if or.ExpiresIn > 0 {
		accessTokenTime = time.Duration(or.ExpiresIn)
	}

	return nil
}
