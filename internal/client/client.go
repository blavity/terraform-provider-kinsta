package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.kinsta.com/v2"
)

type Client struct {
	apiKey     string
	companyID  string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey, companyID string) *Client {
	return &Client{
		apiKey:    apiKey,
		companyID: companyID,
		baseURL:   DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
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
	defer resp.Body.Close()

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

type CreateDatabaseRequest struct {
	CompanyID    string `json:"company_id"`
	Location     string `json:"location"`
	ResourceType string `json:"resource_type"`
	DisplayName  string `json:"display_name"`
	DBName       string `json:"db_name"`
	DBPassword   string `json:"db_password"`
	DBUser       string `json:"db_user,omitempty"`
	Type         string `json:"type"`
	Version      string `json:"version"`
}

type CreateDatabaseResponse struct {
	Database struct {
		ID string `json:"id"`
	} `json:"database"`
}

type Database struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
	MemoryLimit    int    `json:"memory_limit"`
	CPULimit       int    `json:"cpu_limit"`
	StorageSize    int    `json:"storage_size"`
	Type           string `json:"type"`
	Version        string `json:"version"`
	Cluster        struct {
		ID          string `json:"id"`
		Location    string `json:"location"`
		DisplayName string `json:"display_name"`
	} `json:"cluster"`
	ResourceType string `json:"resource_type_name"`
}

type GetDatabaseResponse struct {
	Database Database `json:"database"`
}

type UpdateDatabaseRequest struct {
	ResourceType string `json:"resource_type,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
}

type UpdateDatabaseResponse struct {
	Database struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Status      string `json:"status"`
	} `json:"database"`
}

type DeleteDatabaseResponse struct {
	Message string `json:"message"`
}

type CreateApplicationRequest struct {
	CompanyID   string `json:"company_id"`
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Region      string `json:"region"`
}

type CreateApplicationResponse struct {
	Application struct {
		ID string `json:"id"`
	} `json:"application"`
}

type Application struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Region      string `json:"region"`
}

type GetApplicationResponse struct {
	Application Application `json:"app"`
}

type UpdateApplicationRequest struct {
	DisplayName string `json:"display_name,omitempty"`
}

type UpdateApplicationResponse struct {
	Application struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Status      string `json:"status"`
	} `json:"app"`
}

type DeleteApplicationResponse struct {
	Message string `json:"message"`
}

type CreateWordPressSiteRequest struct {
	Company      string `json:"company"`
	DisplayName  string `json:"display_name"`
	Region       string `json:"region"`
	AdminEmail   string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AdminUser    string `json:"admin_user"`
	SiteTitle    string `json:"site_title"`
	WPLanguage   string `json:"wp_language"`
}

type CreateWordPressSiteResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type WordPressSite struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Region      string `json:"region"`
}

type GetWordPressSiteResponse struct {
	Site WordPressSite `json:"site"`
}

type DeleteWordPressSiteResponse struct {
	OperationID string `json:"operation_id"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}

type GetOperationResponse struct {
	Status string `json:"status"`
	SiteID string `json:"site_id"`
}

type KinstaClient interface {
	CompanyID() string
	CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*CreateDatabaseResponse, error)
	GetDatabase(ctx context.Context, id string) (*GetDatabaseResponse, error)
	UpdateDatabase(ctx context.Context, id string, req *UpdateDatabaseRequest) (*UpdateDatabaseResponse, error)
	DeleteDatabase(ctx context.Context, id string) (*DeleteDatabaseResponse, error)
	CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*CreateApplicationResponse, error)
	GetApplication(ctx context.Context, id string) (*GetApplicationResponse, error)
	UpdateApplication(ctx context.Context, id string, req *UpdateApplicationRequest) (*UpdateApplicationResponse, error)
	DeleteApplication(ctx context.Context, id string) (*DeleteApplicationResponse, error)
	CreateWordPressSite(ctx context.Context, req *CreateWordPressSiteRequest) (*CreateWordPressSiteResponse, error)
	GetWordPressSite(ctx context.Context, id string) (*GetWordPressSiteResponse, error)
	DeleteWordPressSite(ctx context.Context, id string) (*DeleteWordPressSiteResponse, error)
	GetOperation(ctx context.Context, id string) (*GetOperationResponse, error)
}

func (c *Client) CompanyID() string {
	return c.companyID
}
func (c *Client) GetOperation(ctx context.Context, id string) (*GetOperationResponse, error) {
	var getResponse GetOperationResponse

	path := fmt.Sprintf("/operations/%s", id)
	err := c.do(ctx, http.MethodGet, path, nil, &getResponse)
	if err != nil {
		return nil, err
	}

	return &getResponse, nil
}

func (c *Client) CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*CreateDatabaseResponse, error) {
	var createResponse CreateDatabaseResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.do(ctx, http.MethodPost, "/databases", bytes.NewBuffer(body), &createResponse)
	if err != nil {
		return nil, err
	}

	return &createResponse, nil
}

func (c *Client) GetDatabase(ctx context.Context, id string) (*GetDatabaseResponse, error) {
	var getResponse GetDatabaseResponse

	path := fmt.Sprintf("/databases/%s", id)
	err := c.do(ctx, http.MethodGet, path, nil, &getResponse)
	if err != nil {
		return nil, err
	}

	return &getResponse, nil
}

func (c *Client) UpdateDatabase(ctx context.Context, id string, req *UpdateDatabaseRequest) (*UpdateDatabaseResponse, error) {
	var updateResponse UpdateDatabaseResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/databases/%s", id)
	err = c.do(ctx, http.MethodPut, path, bytes.NewBuffer(body), &updateResponse)
	if err != nil {
		return nil, err
	}

	return &updateResponse, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, id string) (*DeleteDatabaseResponse, error) {
	var deleteResponse DeleteDatabaseResponse

	path := fmt.Sprintf("/databases/%s", id)
	err := c.do(ctx, http.MethodDelete, path, nil, &deleteResponse)
	if err != nil {
		return nil, err
	}

	return &deleteResponse, nil
}

func (c *Client) CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*CreateApplicationResponse, error) {
	var createResponse CreateApplicationResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.do(ctx, http.MethodPost, "/applications", bytes.NewBuffer(body), &createResponse)
	if err != nil {
		return nil, err
	}

	return &createResponse, nil
}

func (c *Client) GetApplication(ctx context.Context, id string) (*GetApplicationResponse, error) {
	var getResponse GetApplicationResponse

	path := fmt.Sprintf("/applications/%s", id)
	err := c.do(ctx, http.MethodGet, path, nil, &getResponse)
	if err != nil {
		return nil, err
	}

	return &getResponse, nil
}

func (c *Client) UpdateApplication(ctx context.Context, id string, req *UpdateApplicationRequest) (*UpdateApplicationResponse, error) {
	var updateResponse UpdateApplicationResponse

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/applications/%s", id)
	err = c.do(ctx, http.MethodPut, path, bytes.NewBuffer(body), &updateResponse)
	if err != nil {
		return nil, err
	}

	return &updateResponse, nil
}

func (c *Client) DeleteApplication(ctx context.Context, id string) (*DeleteApplicationResponse, error) {
	var deleteResponse DeleteApplicationResponse

	path := fmt.Sprintf("/applications/%s", id)
	err := c.do(ctx, http.MethodDelete, path, nil, &deleteResponse)
	if err != nil {
		return nil, err
	}

	return &deleteResponse, nil
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

func (c *Client) GetWordPressSite(ctx context.Context, id string) (*GetWordPressSiteResponse, error) {
	var getResponse GetWordPressSiteResponse

	path := fmt.Sprintf("/sites/%s", id)
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
