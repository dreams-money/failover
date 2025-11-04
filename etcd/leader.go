package etcd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dreams-money/opnsense-failover/config"
)

var (
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

func GetLeaderName(cfg config.Config) (string, error) {
	key := "/service/postgres-ha/leader"
	key = base64.StdEncoding.EncodeToString([]byte(key))

	key = "{\"key\": \"" + key + "\"}"
	body := bytes.NewBuffer([]byte(key))
	url := "http://%v:%v/v3/kv/range"
	url = fmt.Sprintf(url, cfg.ETCDAddress, cfg.ETCDPort)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("failed to call ETCD")
	}

	type Value struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	type ETCDResponse struct {
		KeyValues []Value `json:"kvs"`
		Count     string  `json:"count"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := ETCDResponse{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return "", err
	}

	responseCount, err := strconv.Atoi(response.Count)
	if err != nil {
		return "", err
	}

	if responseCount > 1 {
		log.Println("Weird count on ETCD single key retrival")
	} else if responseCount == 0 {
		return "", errors.New("failed to fetch value from ETCD")
	}

	keyValueEncoded := response.KeyValues[0].Value
	value, err := base64.StdEncoding.DecodeString(keyValueEncoded)

	if len(value) == 0 {
		return "", fmt.Errorf("failed to get leader")
	}

	return string(value), err
}
