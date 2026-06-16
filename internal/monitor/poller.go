// Package monitor provides HTTP health checking for gpt-load instances.
package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HealthResult holds the parsed response from GET /health.
type HealthResult struct {
	Online     bool   `json:"online"`
	Status     string `json:"status"`
	Uptime     string `json:"uptime"`
	Timestamp  string `json:"timestamp"`
	ResponseMs int64  `json:"response_ms"`
	Error      string `json:"error,omitempty"`
}

// StatsResult holds the parsed response from GET /api/dashboard/stats.
type StatsResult struct {
	TotalRequests  int     `json:"total_requests"`
	TotalSuccess   int     `json:"total_success"`
	TotalFailures  int     `json:"total_failures"`
	ErrorRate      float64 `json:"error_rate"`
}

// Poller periodically checks gpt-load instances via HTTP.
type Poller struct {
	client *http.Client
}

// NewPoller creates a health check poller with a 5-second timeout.
func NewPoller() *Poller {
	return &Poller{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// HealthCheck performs GET /health on the target gpt-load instance.
func (p *Poller) HealthCheck(host string, port int) *HealthResult {
	url := fmt.Sprintf("http://%s:%d/health", host, port)

	start := time.Now()
	resp, err := p.client.Get(url)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return &HealthResult{
			Online:     false,
			ResponseMs: elapsed,
			Error:      err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Status    string `json:"status"`
		Uptime    string `json:"uptime"`
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return &HealthResult{
			Online:     resp.StatusCode == 200,
			ResponseMs: elapsed,
			Status:     fmt.Sprintf("http_%d", resp.StatusCode),
		}
	}

	return &HealthResult{
		Online:     resp.StatusCode == 200 && result.Status == "healthy",
		Status:     result.Status,
		Uptime:     result.Uptime,
		Timestamp:  result.Timestamp,
		ResponseMs: elapsed,
	}
}

// FetchStats gets the dashboard stats from a gpt-load master.
func (p *Poller) FetchStats(host string, port int, authKey string) (*StatsResult, error) {
	url := fmt.Sprintf("http://%s:%d/api/dashboard/stats", host, port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+authKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var wrapper struct {
		Data StatsResult `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		// try flat
		var sr StatsResult
		if err := json.Unmarshal(body, &sr); err != nil {
			return nil, err
		}
		return &sr, nil
	}

	// calculate error rate
	result := wrapper.Data
	if result.TotalRequests > 0 {
		result.ErrorRate = float64(result.TotalFailures) / float64(result.TotalRequests) * 100
	}

	return &result, nil
}

// FetchClusterNodes gets the cluster node list from a gpt-load master.
func (p *Poller) FetchClusterNodes(host string, port int, authKey string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d/api/cluster/nodes", host, port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+authKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var wrapper struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Data, nil
}

// GetGPTModeFromHealth tries to detect the gpt-load mode from /health and /api endpoints.
func GetGPTModeFromHealth(health *HealthResult) string {
	if !health.Online {
		return "unknown"
	}
	if strings.Contains(health.Status, "slave") || strings.Contains(health.Uptime, "slave") {
		return "slave"
	}
	return "standalone"
}
