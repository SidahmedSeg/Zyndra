package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, channel string, data any) error
}

type CentrifugoPublisher struct {
	apiURL string
	apiKey string
	client *http.Client
}

func NewCentrifugoPublisher(apiURL, apiKey string) *CentrifugoPublisher {
	return &CentrifugoPublisher{
		apiURL: apiURL,
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type centrifugoPublishRequest struct {
	Method string `json:"method"`
	Params struct {
		Channel string `json:"channel"`
		Data    any    `json:"data"`
	} `json:"params"`
}

type centrifugoError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type centrifugoResponse struct {
	Error *centrifugoError `json:"error,omitempty"`
}

func (p *CentrifugoPublisher) Enabled() bool {
	return p != nil && p.apiURL != "" && p.apiKey != ""
}

func (p *CentrifugoPublisher) Publish(ctx context.Context, channel string, data any) error {
	if !p.Enabled() {
		return nil
	}
	if channel == "" {
		return fmt.Errorf("missing channel")
	}

	reqBody := centrifugoPublishRequest{Method: "publish"}
	reqBody.Params.Channel = channel
	reqBody.Params.Data = data

	b, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal publish request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("centrifugo publish request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("centrifugo publish failed: status=%d", resp.StatusCode)
	}

	var out centrifugoResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.Error != nil {
		return fmt.Errorf("centrifugo publish error: %s", out.Error.Message)
	}

	return nil
}


