package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/haya14busa/reviewdog"
	"github.com/haya14busa/reviewdog/doghouse"
)

const baseEndpoint = "https://review-dog.appspot.com"

type DogHouseClient struct {
	Client *http.Client
	// Base URL for API requests. Defaults is https://review-dog.appspot.com.
	BaseURL *url.URL
}

func New(client *http.Client) *DogHouseClient {
	dh := &DogHouseClient{Client: client}
	if dh.Client == nil {
		dh.Client = http.DefaultClient
	}
	base := baseEndpoint
	if baseEnvURL := os.Getenv("REVIEWDOG_GITHUB_APP_URL"); baseEnvURL != "" {
		base = baseEnvURL
	}
	var err error
	dh.BaseURL, err = url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	return dh
}

func (c *DogHouseClient) Check(ctx context.Context, req *doghouse.CheckRequest) (*doghouse.CheckResponse, error) {
	url := c.BaseURL.String() + "/check"
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.WithContext(ctx)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", fmt.Sprintf("reviewdog/%s", reviewdog.Version))

	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Check request failed: %v", err)
	}
	defer httpResp.Body.Close()

	respb, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status=%v: %s", httpResp.StatusCode, respb)
	}

	var resp doghouse.CheckResponse
	if err := json.Unmarshal(respb, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: error=%v, resp=%s", err, respb)
	}
	return &resp, nil
}
