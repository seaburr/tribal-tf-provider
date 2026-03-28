package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TribalClient is the HTTP client for the Tribal API.
type TribalClient struct {
	Host       string
	APIKey     string
	HTTPClient *http.Client
}

func NewTribalClient(host, apiKey string) *TribalClient {
	return &TribalClient{
		Host:   host,
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *TribalClient) doRequest(method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.Host+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// --- Resource types ---

type ResourceRequest struct {
	Name                   string  `json:"name"`
	DRI                    string  `json:"dri"`
	Type                   string  `json:"type"`
	ExpirationDate         *string `json:"expiration_date"`
	DoesNotExpire          bool    `json:"does_not_expire"`
	Purpose                string  `json:"purpose"`
	GenerationInstructions string  `json:"generation_instructions"`
	SecretManagerLink      string  `json:"secret_manager_link,omitempty"`
	SlackWebhook           string  `json:"slack_webhook"`
	TeamID                 *int    `json:"team_id,omitempty"`
	CertificateURL         string  `json:"certificate_url,omitempty"`
	AutoRefreshExpiry      bool    `json:"auto_refresh_expiry"`
}

type ResourceResponse struct {
	ID                     int     `json:"id"`
	Name                   string  `json:"name"`
	DRI                    string  `json:"dri"`
	Type                   string  `json:"type"`
	ExpirationDate         *string `json:"expiration_date"`
	DoesNotExpire          bool    `json:"does_not_expire"`
	Purpose                string  `json:"purpose"`
	GenerationInstructions string  `json:"generation_instructions"`
	SecretManagerLink      *string `json:"secret_manager_link"`
	SlackWebhook           string  `json:"slack_webhook"`
	TeamID                 *int    `json:"team_id"`
	PublicKeyPEM           *string `json:"public_key_pem"`
	CertificateURL         *string `json:"certificate_url"`
	AutoRefreshExpiry      bool    `json:"auto_refresh_expiry"`
	LastReviewedAt         *string `json:"last_reviewed_at"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

func (c *TribalClient) CreateResource(req ResourceRequest) (*ResourceResponse, error) {
	var resp ResourceResponse
	if err := c.doRequest("POST", "/api/resources/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) GetResource(id int) (*ResourceResponse, error) {
	var resp ResourceResponse
	if err := c.doRequest("GET", fmt.Sprintf("/api/resources/%d", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) UpdateResource(id int, req ResourceRequest) (*ResourceResponse, error) {
	var resp ResourceResponse
	if err := c.doRequest("PUT", fmt.Sprintf("/api/resources/%d", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) DeleteResource(id int) error {
	return c.doRequest("DELETE", fmt.Sprintf("/api/resources/%d", id), nil, nil)
}

func (c *TribalClient) ListResources() ([]ResourceResponse, error) {
	var resp []ResourceResponse
	if err := c.doRequest("GET", "/api/resources/", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// --- Admin Settings types ---

type AdminSettingsRequest struct {
	ReminderDays         []int   `json:"reminder_days"`
	NotifyHour           int     `json:"notify_hour"`
	SlackWebhook         *string `json:"slack_webhook"`
	AlertOnOverdue       bool    `json:"alert_on_overdue"`
	AlertOnDelete        bool    `json:"alert_on_delete"`
	AlertOnReviewOverdue bool    `json:"alert_on_review_overdue"`
	ReviewCadenceMonths  *int    `json:"review_cadence_months"`
}

type AdminSettingsResponse struct {
	ReminderDays         []int   `json:"reminder_days"`
	NotifyHour           int     `json:"notify_hour"`
	SlackWebhook         *string `json:"slack_webhook"`
	AlertOnOverdue       bool    `json:"alert_on_overdue"`
	AlertOnDelete        bool    `json:"alert_on_delete"`
	AlertOnReviewOverdue bool    `json:"alert_on_review_overdue"`
	ReviewCadenceMonths  *int    `json:"review_cadence_months"`
}

// --- Team types ---

type TeamRequest struct {
	Name string `json:"name"`
}

type TeamResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (c *TribalClient) ListTeams() ([]TeamResponse, error) {
	var resp []TeamResponse
	if err := c.doRequest("GET", "/admin/teams", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *TribalClient) GetTeam(id int) (*TeamResponse, error) {
	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}
	for _, t := range teams {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("API error 404: team %d not found", id)
}

func (c *TribalClient) CreateTeam(req TeamRequest) (*TeamResponse, error) {
	var resp TeamResponse
	if err := c.doRequest("POST", "/admin/teams", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) UpdateTeam(id int, req TeamRequest) (*TeamResponse, error) {
	var resp TeamResponse
	if err := c.doRequest("PUT", fmt.Sprintf("/admin/teams/%d", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) GetAdminSettings() (*AdminSettingsResponse, error) {
	var resp AdminSettingsResponse
	if err := c.doRequest("GET", "/admin/settings", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *TribalClient) UpdateAdminSettings(req AdminSettingsRequest) (*AdminSettingsResponse, error) {
	var resp AdminSettingsResponse
	if err := c.doRequest("PUT", "/admin/settings", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
