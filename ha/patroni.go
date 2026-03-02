package ha

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dreams-money/failover/config"
)

type Patroni struct{}

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}
)

// Really we should have used the patroni API. smh.
func (Patroni) GetClusterStatus(cfg config.Config) (ClusterStatus, error) {
	encoder := base64.URLEncoding

	type etcdRangeRequest struct {
		Key      string `json:"key"`
		RangeEnd string `json:"range_end"`
	}
	rangeRequest := etcdRangeRequest{
		Key:      "/service/postgres-ha/members",
		RangeEnd: "/service/postgres-ha/memberst",
	}
	rangeRequest.Key = encoder.EncodeToString([]byte(rangeRequest.Key))
	rangeRequest.RangeEnd = encoder.EncodeToString([]byte(rangeRequest.RangeEnd))

	status := ClusterStatus{}

	rangeRequestPayload, err := json.Marshal(rangeRequest)
	if err != nil {
		return status, err
	}
	buffer := bytes.NewBuffer(rangeRequestPayload)

	url := fmt.Sprintf("%v/v3/kv/range", cfg.HighAvailabilityAPIAddress)
	request, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return status, err
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		return status, err
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return status, err
	}

	if resp.StatusCode != 200 {
		e := "patroni: etcd range keys request failed. (%v) %v"
		return status, fmt.Errorf(e, resp.StatusCode, string(respBodyBytes))
	}

	type rangeResult struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	type rangeResponse struct {
		Count string        `json:"count"`
		KVS   []rangeResult `json:"kvs"`
	}
	rangeRes := rangeResponse{}

	err = json.Unmarshal(respBodyBytes, &rangeRes)
	if err != nil {
		return status, err
	}

	rangeCount, err := strconv.Atoi(rangeRes.Count)
	if err != nil {
		return status, err
	} else if rangeCount < 1 {
		return status, errors.New("patroni - no nodes")
	}

	type nodeStatus struct {
		Role  string `json:"role"`
		State string `json:"State"`
	}

	for _, result := range rangeRes.KVS {
		var err error
		var buffer []byte
		node := Node{}

		buffer, err = encoder.DecodeString(result.Key)
		if err != nil {
			return status, err
		}
		result.Key = string(buffer)

		buffer, err = encoder.DecodeString(result.Value)
		if err != nil {
			return status, err
		}
		result.Value = string(buffer)

		node.Name = strings.ReplaceAll(result.Key, "/service/postgres-ha/members/", "")

		ns := nodeStatus{}
		err = json.Unmarshal([]byte(result.Value), &ns)
		if err != nil {
			return status, err
		}

		node.Role = ns.Role
		node.State = ns.State

		status = append(status, node)
	}

	return status, nil
}
