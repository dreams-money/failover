package omada

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dreams-money/failover/persistence"
)

var (
	accessTokenTime   = time.Duration(7200 * time.Second)
	ticker            = time.NewTicker(accessTokenTime)
	NeedsTokenRefresh = errors.New("Refresh Token")
)

func CheckAccessTokens(c Config) error {
	// Check if tokens exist
	_, err := persistence.GetAccessToken()
	if err != nil {
		return err
	}
	_, err = persistence.GetRefreshToken()
	if err != nil {
		return err
	}

	// Check if we need to refresh tokens
	accessTokenAge, err := persistence.GetAccessTokenTime()
	if err != nil {
		return err
	}
	if accessTokenAge.Before(time.Now().Add(-1 * accessTokenTime)) {
		return NeedsTokenRefresh
	}

	return nil
}

func RefreshTokens(cfg Config) error {
	newTime, err := refreshToken(cfg)
	if err != nil {
		return err
	}

	log.Println("Successfully refreshed tokens.")

	if newTime != accessTokenTime {
		accessTokenTime = newTime
		ticker.Reset(newTime)
	}

	return nil
}

func RefreshTokensJob(cfg Config) {
	var err error
	for range ticker.C {
		err = RefreshTokens(cfg)
		if err != nil {
			log.Println(err)
		}
	}
}

func refreshToken(c Config) (time.Duration, error) {
	log.Println("Refreshing access tokens")

	refreshToken, err := persistence.GetRefreshToken()
	if err != nil {
		return accessTokenTime, err
	}

	url := "https://%v:%v"
	url = fmt.Sprintf(url, c.OmadaAddress, c.OmadaPort)
	url += "/openapi/authorize/token"
	url += "?client_id=%v"
	url += "&client_secret=%v"
	url += "&refresh_token=%v"
	url += "&grant_type=refresh_token"

	url = fmt.Sprintf(url, c.OmadaOAuthClientID, c.OmadaOAuthClientSecret, refreshToken)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return accessTokenTime, err
	}

	req.Header.Add("content-type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return accessTokenTime, err
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if resp.StatusCode != 200 {
		e := "response failed (%v) %v"
		return accessTokenTime, fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type Result struct {
		AccessToken  string `json:"accessToken"`
		TokenType    string `json:"tokenType"`
		ExpiresIn    int64  `json:"expiresIn"`
		RefreshToken string `json:"refreshToken"`
	}

	type RefreshResponse struct {
		ErrorCode int    `json:"errorCode"`
		Message   string `json:"msg"`
		Result    `json:"result"`
	}

	rr := RefreshResponse{}
	err = json.Unmarshal(respBodyBytes, &rr)
	if err != nil {
		return accessTokenTime, err
	}

	if rr.ErrorCode < 0 {
		e := "omada error (%v): %v"
		return accessTokenTime, fmt.Errorf(e, rr.ErrorCode, rr.Message)
	}

	expireDuration := time.Duration(rr.Result.ExpiresIn)
	expireDuration *= time.Second

	err = persistence.SetAccessToken(rr.Result.AccessToken)
	if err != nil {
		return expireDuration, err
	}
	err = persistence.SetRefreshToken(rr.Result.RefreshToken)
	if err != nil {
		return expireDuration, err
	}

	return expireDuration, nil
}
