package clairic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Result struct {
	Answers   map[string]json.RawMessage `json:"answers"`
	Truncated bool                       `json:"truncated"`
	StateText string                     `json:"state_text,omitempty"`
	Usage     map[string]any             `json:"usage,omitempty"`
}
type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 120 * time.Second}, BaseURL: "https://api.cloudflare.com/client/v4"}
}
func (c *Client) Extract(ctx context.Context, account, token string, request Request) (Result, error) {
	var result Result
	if !AccountPattern.MatchString(account) {
		return result, errors.New("Enter a 32-character Cloudflare account ID")
	}
	if strings.TrimSpace(token) == "" {
		return result, errors.New("Connect your Cloudflare API token first")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return result, errors.New("Could not encode the request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/accounts/"+account+"/ai/run/@cf/cloudflare/clairic", bytes.NewReader(body))
	if err != nil {
		return result, errors.New("Could not create the Cloudflare request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	response, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return result, errors.New("Extraction cancelled")
		}
		return result, errors.New("Cloudflare request failed; check connectivity and timeout")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return result, fmt.Errorf("Cloudflare HTTP %d; check account, token permissions, quota, and request fields", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 32<<20+1))
	if err != nil || len(data) > 32<<20 {
		return result, errors.New("Cloudflare response could not be read or exceeded 32 MiB")
	}
	var envelope struct {
		Success *bool           `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	if err = json.Unmarshal(data, &envelope); err != nil {
		return result, errors.New("Cloudflare returned invalid JSON")
	}
	if envelope.Success != nil {
		if !*envelope.Success {
			return result, errors.New("Cloudflare reported an unsuccessful request")
		}
		data = envelope.Result
	}
	if err = json.Unmarshal(data, &result); err != nil || len(result.Answers) == 0 {
		return result, errors.New("Cloudflare response is missing typed answers")
	}
	for name := range request.Fields {
		raw, ok := result.Answers[name]
		if !ok {
			return result, fmt.Errorf("Response is missing field %s", name)
		}
		var answer map[string]json.RawMessage
		if json.Unmarshal(raw, &answer) != nil || answer["value"] == nil {
			return result, fmt.Errorf("Answer %s is missing its value", name)
		}
	}
	return result, nil
}
