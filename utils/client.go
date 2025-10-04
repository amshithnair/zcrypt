// utils/client.go
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LogClient struct {
	BaseURL string
	Client  *http.Client
}

type LogSubmission struct {
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
	PubKey    string                 `json:"pubkey"`
	AgentID   string                 `json:"agent_id"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ServerResponse struct {
	Success     bool                   `json:"success"`
	Entry       interface{}            `json:"entry,omitempty"`
	ChainLength int                    `json:"chain_length,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// NewLogClient creates a new client for the Zcrypt server
func NewLogClient(baseURL string) *LogClient {
	return &LogClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SubmitLog sends a log entry to the server
func (lc *LogClient) SubmitLog(submission LogSubmission) (*ServerResponse, error) {
	url := fmt.Sprintf("%s/api/v1/logs", lc.BaseURL)

	jsonData, err := json.Marshal(submission)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := lc.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var serverResp ServerResponse
	if err := json.Unmarshal(body, &serverResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return &serverResp, fmt.Errorf("server error: %s", serverResp.Error)
	}

	return &serverResp, nil
}

// VerifyChain asks the server to verify its chain
func (lc *LogClient) VerifyChain() (*ServerResponse, error) {
	url := fmt.Sprintf("%s/api/v1/verify/chain", lc.BaseURL)

	resp, err := lc.Client.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var serverResp ServerResponse
	if err := json.Unmarshal(body, &serverResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &serverResp, nil
}

// GetStats retrieves server statistics
func (lc *LogClient) GetStats() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/stats", lc.BaseURL)

	resp, err := lc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return stats, nil
}

// RegisterAgent registers an agent with the server
func (lc *LogClient) RegisterAgent(agentID, pubKey, name string) error {
	url := fmt.Sprintf("%s/api/v1/agents/register", lc.BaseURL)

	data := map[string]string{
		"agent_id": agentID,
		"pubkey":   pubKey,
		"name":     name,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := lc.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}

	return nil
}

// HealthCheck checks if the server is running
func (lc *LogClient) HealthCheck() (bool, error) {
	url := fmt.Sprintf("%s/api/v1/health", lc.BaseURL)

	resp, err := lc.Client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}