package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	DefaultBaseURL = "https://api.kinsta.com/v2"
)

// NotFoundError is returned by client methods when the API responds with 404.
type NotFoundError struct {
	Path string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.Path)
}

// IsNotFound returns true if err is a *NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

type Client struct {
	apiKey     string
	companyID  string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey, companyID string) *Client {
	return &Client{
		apiKey:     apiKey,
		companyID:  companyID,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.baseURL, path), body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are not actionable

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Path: path}
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

// CreateWordPressSiteRequest is the body for `addWPSite-Body` in the
// Kinsta v2 OpenAPI spec — provisions a full WordPress install with
// admin user, site title, language, etc.
type CreateWordPressSiteRequest struct {
	Company              string `json:"company"`
	DisplayName          string `json:"display_name"`
	Region               string `json:"region"`
	InstallMode          string `json:"install_mode"`
	AdminEmail           string `json:"admin_email"`
	AdminPassword        string `json:"admin_password"`
	AdminUser            string `json:"admin_user"`
	SiteTitle            string `json:"site_title"`
	WPLanguage           string `json:"wp_language"`
	IsMultisite          bool   `json:"is_multisite"`
	IsSubdomainMultisite bool   `json:"is_subdomain_multisite"`
	WooCommerce          bool   `json:"woocommerce"`
	WordPressSEO         bool   `json:"wordpressseo"`
}

// CreatePlainWordPressSiteRequest is the body for `addPlainWPSite-Body`
// in the Kinsta v2 OpenAPI spec — provisions an empty WordPress
// container (no install template, no admin user). Per the spec only
// company/display_name/region are accepted on the wire; admin_*,
// site_title, wp_language, install_mode, multisite/etc. fields are NOT
// part of the addPlainWPSite shape.
type CreatePlainWordPressSiteRequest struct {
	Company     string `json:"company"`
	DisplayName string `json:"display_name"`
	Region      string `json:"region"`
}

type CreateWordPressSiteResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type WordPressSite struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Status       string                 `json:"status"`
	Environments []WordPressEnvironment `json:"environments"`
}

type WordPressEnvironment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type GetWordPressSiteResponse struct {
	Site WordPressSite `json:"site"`
}

type GetWordPressSitesResponse struct {
	Company struct {
		Sites []WordPressSite `json:"sites"`
	} `json:"company"`
}

type DeleteWordPressSiteResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type OperationResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type CreateWordPressEnvironmentRequest struct {
	DisplayName   string `json:"display_name"`
	SiteTitle     string `json:"site_title"`
	IsPremium     bool   `json:"is_premium"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AdminUser     string `json:"admin_user"`
	WPLanguage    string `json:"wp_language"`
}

type CreateWordPressEnvironmentResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type DeleteWordPressEnvironmentResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type KinstaClient interface {
	CompanyID() string
	CreateWordPressSite(ctx context.Context, req *CreateWordPressSiteRequest) (*CreateWordPressSiteResponse, error)
	CreatePlainWordPressSite(ctx context.Context, req *CreatePlainWordPressSiteRequest) (*CreateWordPressSiteResponse, error)
	GetWordPressSite(ctx context.Context, id string) (*GetWordPressSiteResponse, error)
	GetWordPressSites(ctx context.Context) (*GetWordPressSitesResponse, error)
	DeleteWordPressSite(ctx context.Context, id string) (*DeleteWordPressSiteResponse, error)
	CreateWordPressEnvironment(ctx context.Context, siteID string, req *CreateWordPressEnvironmentRequest) (*CreateWordPressEnvironmentResponse, error)
	DeleteWordPressEnvironment(ctx context.Context, siteID, envID string) (*DeleteWordPressEnvironmentResponse, error)
	PollOperation(ctx context.Context, operationID string) (string, error)
}

func (c *Client) CompanyID() string {
	return c.companyID
}

func (c *Client) CreateWordPressSite(ctx context.Context, req *CreateWordPressSiteRequest) (*CreateWordPressSiteResponse, error) {
	var createResponse CreateWordPressSiteResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.do(ctx, http.MethodPost, "/sites", bytes.NewBuffer(body), &createResponse)
	if err != nil {
		return nil, err
	}

	return &createResponse, nil
}

// CreatePlainWordPressSite posts the `addPlainWPSite-Body` shape to
// POST /sites — the spec's documented path for empty-container sites.
// Same response envelope as CreateWordPressSite.
func (c *Client) CreatePlainWordPressSite(ctx context.Context, req *CreatePlainWordPressSiteRequest) (*CreateWordPressSiteResponse, error) {
	var createResponse CreateWordPressSiteResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.do(ctx, http.MethodPost, "/sites", bytes.NewBuffer(body), &createResponse)
	if err != nil {
		return nil, err
	}

	return &createResponse, nil
}

func (c *Client) GetWordPressSite(ctx context.Context, id string) (*GetWordPressSiteResponse, error) {
	var getResponse GetWordPressSiteResponse

	path := fmt.Sprintf("/sites/%s", id)
	err := c.do(ctx, http.MethodGet, path, nil, &getResponse)
	if err != nil {
		return nil, err
	}

	return &getResponse, nil
}

func (c *Client) GetWordPressSites(ctx context.Context) (*GetWordPressSitesResponse, error) {
	var getResponse GetWordPressSitesResponse

	path := fmt.Sprintf("/sites?company=%s", c.companyID)
	err := c.do(ctx, http.MethodGet, path, nil, &getResponse)
	if err != nil {
		return nil, err
	}

	return &getResponse, nil
}

func (c *Client) DeleteWordPressSite(ctx context.Context, id string) (*DeleteWordPressSiteResponse, error) {
	var deleteResponse DeleteWordPressSiteResponse

	path := fmt.Sprintf("/sites/%s", id)
	err := c.do(ctx, http.MethodDelete, path, nil, &deleteResponse)
	if err != nil {
		return nil, err
	}

	return &deleteResponse, nil
}

func (c *Client) CreateWordPressEnvironment(ctx context.Context, siteID string, req *CreateWordPressEnvironmentRequest) (*CreateWordPressEnvironmentResponse, error) {
	var createResponse CreateWordPressEnvironmentResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/sites/%s/environments", siteID)
	err = c.do(ctx, http.MethodPost, path, bytes.NewBuffer(body), &createResponse)
	if err != nil {
		return nil, err
	}

	return &createResponse, nil
}

// DeleteWordPressEnvironment matches the spec path
// `DELETE /sites/{site_id}/environments/{env_id}` (per
// deleteWpEnvironment-Params). The pre-spec-audit version called
// `/sites/environments/{env_id}` — an undocumented route that may
// silently fail against the real API.
func (c *Client) DeleteWordPressEnvironment(ctx context.Context, siteID, envID string) (*DeleteWordPressEnvironmentResponse, error) {
	var deleteResponse DeleteWordPressEnvironmentResponse

	path := fmt.Sprintf("/sites/%s/environments/%s", siteID, envID)
	err := c.do(ctx, http.MethodDelete, path, nil, &deleteResponse)
	if err != nil {
		return nil, err
	}

	return &deleteResponse, nil
}

// pollBackoff returns the wait duration for a given in-progress attempt using
// exponential backoff capped at 30s: 2s, 4s, 8s, 15s, 30s, 30s, ...
func pollBackoff(attempt int) time.Duration {
	intervals := []time.Duration{2, 4, 8, 15, 30}
	if attempt < len(intervals) {
		return intervals[attempt] * time.Second
	}
	return 30 * time.Second
}

func (c *Client) PollOperation(ctx context.Context, operationID string) (string, error) {
	path := fmt.Sprintf("/operations/%s", operationID)

	// 404 grace period: operation may not be initialized for up to 30s after creation.
	const grace404Max = 6
	const grace404Wait = 5 * time.Second
	grace404Count := 0

	startTime := time.Now()

	for attempt := 0; ; attempt++ {
		tflog.Info(ctx, "polling operation", map[string]interface{}{
			"operation_id": operationID,
			"attempt":      attempt + 1,
			"elapsed":      time.Since(startTime).String(),
		})

		var opResp OperationResponse
		err := c.do(ctx, http.MethodGet, path, nil, &opResp)
		if err != nil {
			// Retry any error (typically 404) within the grace period.
			if grace404Count < grace404Max {
				grace404Count++
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(grace404Wait):
				}
				continue
			}
			return "", fmt.Errorf("operation %s: %w", operationID, err)
		}

		// Kinsta operations API uses HTTP 200 for all terminal states;
		// the inner status field indicates the actual outcome.
		switch opResp.Status {
		case 200:
			tflog.Info(ctx, "operation completed successfully", map[string]interface{}{
				"operation_id": operationID,
				"elapsed":      time.Since(startTime).String(),
			})
			// idSite is present for site creation; idEnv may be absent (use before/after lookup).
			if siteID, ok := opResp.Data["idSite"].(string); ok {
				return siteID, nil
			}
			if envID, ok := opResp.Data["idEnv"].(string); ok {
				return envID, nil
			}
			return "", nil

		case 500:
			dataJSON, _ := json.Marshal(opResp.Data)
			return "", fmt.Errorf("operation %s failed: %s, data: %s", operationID, opResp.Message, string(dataJSON))
		}

		// 202 in progress — exponential backoff.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollBackoff(attempt)):
		}
	}
}
