package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollOperation_Success(t *testing.T) {
	// Track number of requests
	requestCount := 0
	expectedSiteID := "test-site-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/operations/test-op-123", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		requestCount++

		var response OperationResponse
		if requestCount < 3 {
			// First 2 requests return in-progress (202)
			response = OperationResponse{
				Status:  202,
				Message: "Operation in progress",
			}
		} else {
			// Third request returns success (200)
			response = OperationResponse{
				Status:  200,
				Message: "Operation completed",
				Data: map[string]interface{}{
					"idSite": expectedSiteID,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-123")

	require.NoError(t, err)
	assert.Equal(t, expectedSiteID, siteID)
	assert.Equal(t, 3, requestCount, "Should have polled 3 times before success")
}

func TestPollOperation_ImmediateSuccess(t *testing.T) {
	expectedSiteID := "test-site-456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OperationResponse{
			Status:  200,
			Message: "Operation completed",
			Data: map[string]interface{}{
				"idSite": expectedSiteID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-456")

	require.NoError(t, err)
	assert.Equal(t, expectedSiteID, siteID)
}

func TestPollOperation_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OperationResponse{
			Status:  500,
			Message: "database error",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-789")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed: database error")
	assert.Empty(t, siteID)
}

func TestPollOperation_NoSiteIDInResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OperationResponse{
			Status:  200,
			Message: "Operation completed",
			Data: map[string]interface{}{
				// Missing site_id
				"other_field": "value",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-no-site-id")

	// When operation completes without idSite/idEnv (e.g. environment creation),
	// PollOperation returns ("", nil) — the caller is responsible for resource discovery.
	require.NoError(t, err)
	assert.Empty(t, siteID)
}

func TestPollOperation_404RetrySuccess(t *testing.T) {
	requestCount := 0
	expectedSiteID := "test-site-retry"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount < 3 {
			// First 2 requests return 404 (operation not initialized yet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Operation not found"}`))
			return
		}

		// Subsequent requests succeed
		if requestCount == 3 {
			response := OperationResponse{
				Status:  202,
				Message: "Operation in progress",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			response := OperationResponse{
				Status:  200,
				Message: "Operation completed",
				Data: map[string]interface{}{
					"idSite": expectedSiteID,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-retry")

	require.NoError(t, err)
	assert.Equal(t, expectedSiteID, siteID)
	assert.True(t, requestCount >= 3, "Should retry after 404 errors")
}

func TestPollOperation_404RetryExhausted(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Always return 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Operation not found"}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-404")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Empty(t, siteID)
	// Should retry for first 5 attempts (25 seconds)
	assert.True(t, requestCount >= 5, "Should retry 404 for at least 5 attempts")
}

func TestPollOperation_ContextCancellation(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		response := OperationResponse{
			Status:  202,
			Message: "Operation in progress",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after first request
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	siteID, err := client.PollOperation(ctx, "test-op-cancel")

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, siteID)
}

func TestPollOperation_ContextTimeout(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		response := OperationResponse{
			Status:  202,
			Message: "Operation in progress",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	siteID, err := client.PollOperation(ctx, "test-op-timeout")

	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Empty(t, siteID)
}

func TestPollOperation_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	siteID, err := client.PollOperation(ctx, "test-op-error")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Empty(t, siteID)
}

func TestClient_CompanyID(t *testing.T) {
	client := &Client{
		companyID: "test-company-123",
	}

	assert.Equal(t, "test-company-123", client.CompanyID())
}

func TestClient_New(t *testing.T) {
	client := New("test-api-key", "test-company-id")

	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.apiKey)
	assert.Equal(t, "test-company-id", client.companyID)
	assert.Equal(t, DefaultBaseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestClient_CreateWordPressSite(t *testing.T) {
	expectedRequest := &CreateWordPressSiteRequest{
		Company:       "test-company",
		DisplayName:   "Test Site",
		Region:        "us-central1",
		InstallMode:   "new",
		AdminEmail:    "admin@example.com",
		AdminPassword: "password123",
		AdminUser:     "admin",
		SiteTitle:     "Test Site Title",
		WPLanguage:    "en_US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/sites", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req CreateWordPressSiteRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, expectedRequest.DisplayName, req.DisplayName)
		assert.Equal(t, expectedRequest.Region, req.Region)

		response := CreateWordPressSiteResponse{
			OperationID: "test-op-123",
			Message:     "Site creation started",
			Status:      202,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.CreateWordPressSite(ctx, expectedRequest)

	require.NoError(t, err)
	assert.Equal(t, "test-op-123", resp.OperationID)
	assert.Equal(t, 202, resp.Status)
}

func TestClient_GetWordPressSite(t *testing.T) {
	siteID := "test-site-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, fmt.Sprintf("/sites/%s", siteID), r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := GetWordPressSiteResponse{
			Site: WordPressSite{
				ID:          siteID,
				Name:        "test-site",
				DisplayName: "Test Site",
				Status:      "live",
				Environments: []WordPressEnvironment{
					{
						ID:          "env-123",
						Name:        "live",
						DisplayName: "Live",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.GetWordPressSite(ctx, siteID)

	require.NoError(t, err)
	assert.Equal(t, siteID, resp.Site.ID)
	assert.Equal(t, "Test Site", resp.Site.DisplayName)
	assert.Len(t, resp.Site.Environments, 1)
}

func TestClient_DeleteWordPressSite(t *testing.T) {
	siteID := "test-site-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, fmt.Sprintf("/sites/%s", siteID), r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := DeleteWordPressSiteResponse{
			OperationID: "delete-op-123",
			Message:     "Site deletion started",
			Status:      202,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.DeleteWordPressSite(ctx, siteID)

	require.NoError(t, err)
	assert.Equal(t, "delete-op-123", resp.OperationID)
	assert.Equal(t, 202, resp.Status)
}

func TestClient_CreateWordPressEnvironment_Standard(t *testing.T) {
	siteID := "test-site-123"
	expectedRequest := &CreateWordPressEnvironmentRequest{
		DisplayName:   "staging",
		SiteTitle:     "Test Site - Staging",
		IsPremium:     false,
		AdminEmail:    "admin@example.com",
		AdminPassword: "password123",
		AdminUser:     "admin",
		WPLanguage:    "en_US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, fmt.Sprintf("/sites/%s/environments", siteID), r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req CreateWordPressEnvironmentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, expectedRequest.DisplayName, req.DisplayName)
		assert.Equal(t, expectedRequest.IsPremium, req.IsPremium)
		assert.False(t, req.IsPremium, "Standard staging should have IsPremium=false")

		response := CreateWordPressEnvironmentResponse{
			OperationID: "env-create-op-123",
			Message:     "Environment creation started",
			Status:      202,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.CreateWordPressEnvironment(ctx, siteID, expectedRequest)

	require.NoError(t, err)
	assert.Equal(t, "env-create-op-123", resp.OperationID)
	assert.Equal(t, 202, resp.Status)
}

func TestClient_CreateWordPressEnvironment_Premium(t *testing.T) {
	siteID := "test-site-123"
	expectedRequest := &CreateWordPressEnvironmentRequest{
		DisplayName:   "premium-staging",
		SiteTitle:     "Test Site - Premium Staging",
		IsPremium:     true,
		AdminEmail:    "admin@example.com",
		AdminPassword: "password123",
		AdminUser:     "admin",
		WPLanguage:    "en_US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateWordPressEnvironmentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.IsPremium, "Premium staging should have IsPremium=true")

		response := CreateWordPressEnvironmentResponse{
			OperationID: "env-create-premium-op-123",
			Message:     "Premium environment creation started",
			Status:      202,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.CreateWordPressEnvironment(ctx, siteID, expectedRequest)

	require.NoError(t, err)
	assert.Equal(t, "env-create-premium-op-123", resp.OperationID)
	assert.Equal(t, 202, resp.Status)
}

func TestClient_CreateWordPressEnvironment_Error(t *testing.T) {
	siteID := "test-site-123"
	req := &CreateWordPressEnvironmentRequest{
		DisplayName:   "staging",
		SiteTitle:     "Test Site - Staging",
		IsPremium:     false,
		AdminEmail:    "admin@example.com",
		AdminPassword: "password123",
		AdminUser:     "admin",
		WPLanguage:    "en_US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.CreateWordPressEnvironment(ctx, siteID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_GetWordPressEnvironment(t *testing.T) {
	siteID := "test-site-123"
	envID := "test-env-456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, fmt.Sprintf("/sites/environments/%s", envID), r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := GetWordPressEnvironmentResponse{
			Environment: WordPressEnvironment{
				ID:          envID,
				Name:        "staging",
				DisplayName: "Staging",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.GetWordPressEnvironment(ctx, siteID, envID)

	require.NoError(t, err)
	assert.Equal(t, envID, resp.Environment.ID)
	assert.Equal(t, "Staging", resp.Environment.DisplayName)
}

func TestClient_GetWordPressEnvironment_Error(t *testing.T) {
	siteID := "test-site-123"
	envID := "test-env-456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Environment not found"}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.GetWordPressEnvironment(ctx, siteID, envID)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_DeleteWordPressEnvironment(t *testing.T) {
	envID := "test-env-456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, fmt.Sprintf("/sites/environments/%s", envID), r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := DeleteWordPressEnvironmentResponse{
			OperationID: "delete-env-op-123",
			Message:     "Environment deletion started",
			Status:      202,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.DeleteWordPressEnvironment(ctx, envID)

	require.NoError(t, err)
	assert.Equal(t, "delete-env-op-123", resp.OperationID)
	assert.Equal(t, 202, resp.Status)
}

func TestClient_DeleteWordPressEnvironment_Error(t *testing.T) {
	envID := "test-env-456"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	resp, err := client.DeleteWordPressEnvironment(ctx, envID)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestPollOperation_EnvironmentID(t *testing.T) {
	expectedEnvID := "test-env-789"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/operations/test-env-op-123", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := OperationResponse{
			Status:  200,
			Message: "Operation completed",
			Data: map[string]interface{}{
				"idEnv": expectedEnvID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	envID, err := client.PollOperation(ctx, "test-env-op-123")

	require.NoError(t, err)
	assert.Equal(t, expectedEnvID, envID)
}
