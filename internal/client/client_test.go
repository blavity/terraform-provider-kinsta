package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeJSON writes v as JSON to w. Panics on error — only use in test handlers.
func encodeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(fmt.Sprintf("test server: failed to encode JSON: %v", err))
	}
}

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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
	assert.Contains(t, err.Error(), "operation test-op-789 failed")
	assert.Contains(t, err.Error(), "database error")
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
		encodeJSON(w, response)
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
			_, _ = w.Write([]byte(`{"error": "Operation not found"}`))
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
			encodeJSON(w, response)
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
			encodeJSON(w, response)
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
		_, _ = w.Write([]byte(`{"error": "Operation not found"}`))
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
	assert.Contains(t, err.Error(), "operation test-op-404")
	assert.Empty(t, siteID)
	// Grace period is 6 attempts × 5s = 30s
	assert.True(t, requestCount >= 6, "Should retry 404 for at least 6 grace attempts")
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
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
	assert.Contains(t, err.Error(), "operation test-op-error")
	assert.Contains(t, err.Error(), "API error")
	assert.Empty(t, siteID)
}

func TestPollOperation_ExponentialBackoff(t *testing.T) {
	var requestTimes []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())

		status := 202
		if len(requestTimes) >= 4 {
			status = 200
		}

		response := OperationResponse{
			Status:  status,
			Message: "operation status",
		}
		if status == 200 {
			response.Data = map[string]interface{}{"idSite": "test-site-backoff"}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encodeJSON(w, response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	siteID, err := client.PollOperation(context.Background(), "test-op-backoff")

	require.NoError(t, err)
	assert.Equal(t, "test-site-backoff", siteID)
	require.GreaterOrEqual(t, len(requestTimes), 4, "expected at least 4 requests")

	// Verify intervals grow: first ~2s, second ~4s (50% tolerance for CI)
	interval0 := requestTimes[1].Sub(requestTimes[0]).Seconds()
	interval1 := requestTimes[2].Sub(requestTimes[1]).Seconds()

	assert.InDelta(t, 2.0, interval0, 1.0, "first backoff should be ~2s")
	assert.InDelta(t, 4.0, interval1, 2.0, "second backoff should be ~4s")
	assert.Greater(t, interval1, interval0, "backoff should increase between attempts")
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		encodeJSON(w, response)
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
		_, _ = w.Write([]byte(`{"error": "Invalid request"}`))
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
		encodeJSON(w, response)
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
		_, _ = w.Write([]byte(`{"error": "Environment not found"}`))
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
	assert.True(t, IsNotFound(err), "expected NotFoundError for 404 response")
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
		encodeJSON(w, response)
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
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
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

func TestClient_GetWordPressSites(t *testing.T) {
	companyID := "test-company-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/sites", r.URL.Path)
		assert.Equal(t, companyID, r.URL.Query().Get("company"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		response := GetWordPressSitesResponse{}
		response.Company.Sites = []WordPressSite{
			{ID: "site-1", Name: "site-one", DisplayName: "Site One"},
			{ID: "site-2", Name: "site-two", DisplayName: "Site Two"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encodeJSON(w, response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  companyID,
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.GetWordPressSites(context.Background())

	require.NoError(t, err)
	assert.Len(t, resp.Company.Sites, 2)
	assert.Equal(t, "site-1", resp.Company.Sites[0].ID)
}

func TestClient_GetWordPressSites_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.GetWordPressSites(context.Background())

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_CreateWordPressSite_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.CreateWordPressSite(context.Background(), &CreateWordPressSiteRequest{
		Company:     "test-company",
		DisplayName: "Test Site",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_GetWordPressSite_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.GetWordPressSite(context.Background(), "site-id")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_DeleteWordPressSite_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.DeleteWordPressSite(context.Background(), "site-id")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API error")
}

func TestClient_do_TransportError(t *testing.T) {
	// Start a server, then immediately close it so the transport fails.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := server.URL
	server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    closedURL,
		httpClient: &http.Client{Timeout: 500 * time.Millisecond},
	}

	resp, err := client.GetWordPressSite(context.Background(), "site-id")

	require.Error(t, err)
	assert.Nil(t, resp)
	// Transport failures surface as net errors, not our "API error" string.
	assert.NotContains(t, err.Error(), "API error")
}

func TestClient_do_RequestConstructionError(t *testing.T) {
	// An invalid base URL fails http.NewRequestWithContext.
	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    "://invalid-url",
		httpClient: &http.Client{},
	}

	resp, err := client.GetWordPressSite(context.Background(), "site-id")

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestClient_do_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this is not valid json"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	resp, err := client.GetWordPressSite(context.Background(), "site-id")

	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestPollBackoff(t *testing.T) {
	// Documented schedule: 2, 4, 8, 15, 30, then capped at 30 for any attempt >= 5.
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 2 * time.Second},
		{1, 4 * time.Second},
		{2, 8 * time.Second},
		{3, 15 * time.Second},
		{4, 30 * time.Second},
		{5, 30 * time.Second},
		{6, 30 * time.Second},
		{100, 30 * time.Second},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, pollBackoff(tc.attempt), "attempt %d", tc.attempt)
	}
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
		encodeJSON(w, response)
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

// Principle IV: credentials MUST NOT appear in any returned error string.
// A regression like fmt.Errorf("API error for key %s: %s", c.apiKey, status)
// would silently expose the bearer token in every error a user sees.
func TestClient_NoCredentialLeakInErrors(t *testing.T) {
	const apiKey = "kinsta-api-key-sentinel-xyz"
	const adminPassword = "admin-password-sentinel-abc"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "internal error"}`))
	}))
	defer server.Close()

	c := &Client{
		apiKey:     apiKey,
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	cases := []struct {
		name string
		call func() error
	}{
		{"CreateWordPressSite", func() error {
			_, err := c.CreateWordPressSite(context.Background(), &CreateWordPressSiteRequest{
				DisplayName:   "Test",
				AdminPassword: adminPassword,
			})
			return err
		}},
		{"GetWordPressSite", func() error {
			_, err := c.GetWordPressSite(context.Background(), "site-id")
			return err
		}},
		{"GetWordPressSites", func() error {
			_, err := c.GetWordPressSites(context.Background())
			return err
		}},
		{"DeleteWordPressSite", func() error {
			_, err := c.DeleteWordPressSite(context.Background(), "site-id")
			return err
		}},
		{"CreateWordPressEnvironment", func() error {
			_, err := c.CreateWordPressEnvironment(context.Background(), "site-id", &CreateWordPressEnvironmentRequest{
				DisplayName:   "staging",
				AdminPassword: adminPassword,
			})
			return err
		}},
		{"GetWordPressEnvironment", func() error {
			_, err := c.GetWordPressEnvironment(context.Background(), "site-id", "env-id")
			return err
		}},
		{"DeleteWordPressEnvironment", func() error {
			_, err := c.DeleteWordPressEnvironment(context.Background(), "env-id")
			return err
		}},
		{"PollOperation", func() error {
			// PollOperation will exhaust its 6-attempt 404 grace then return — but
			// since the server returns 500 (not 404), the very first attempt is
			// a non-retryable error path. Cap the call with a context to avoid
			// any unintended waits.
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := c.PollOperation(ctx, "op-id")
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			require.Error(t, err)
			assert.NotContains(t, err.Error(), apiKey, "API key leaked into error: %s", err.Error())
			assert.NotContains(t, err.Error(), adminPassword, "admin password leaked into error: %s", err.Error())
		})
	}
}

// Principle V: tests run under -race because Terraform's resource graph
// invokes provider methods concurrently. This test fires many goroutines
// against the same *Client to give the race detector something to inspect
// on the shared http.Client and any future internal state. It also serves
// as a smoke test that the client itself does not mutate shared state.
func TestClient_ConcurrentCalls(t *testing.T) {
	var requestCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)

		// Respond differently per endpoint so we exercise multiple code paths
		// rather than just hammering one happy path.
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/sites" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				resp := GetWordPressSitesResponse{}
				resp.Company.Sites = []WordPressSite{{ID: "s1"}}
				encodeJSON(w, resp)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			encodeJSON(w, GetWordPressSiteResponse{Site: WordPressSite{ID: "site-id"}})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			encodeJSON(w, DeleteWordPressSiteResponse{OperationID: "op", Status: 202})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	c := &Client{
		apiKey:     "test-api-key",
		companyID:  "test-company",
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	const goroutines = 16
	const callsPerGoroutine = 4

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*callsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ctx := context.Background()
			for j := 0; j < callsPerGoroutine; j++ {
				switch (workerID + j) % 3 {
				case 0:
					if _, err := c.GetWordPressSite(ctx, "site-id"); err != nil {
						errs <- err
					}
				case 1:
					if _, err := c.GetWordPressSites(ctx); err != nil {
						errs <- err
					}
				case 2:
					if _, err := c.DeleteWordPressSite(ctx, "site-id"); err != nil {
						errs <- err
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent call failed: %v", err)
	}

	assert.Equal(t, int64(goroutines*callsPerGoroutine), atomic.LoadInt64(&requestCount),
		"every concurrent call should have reached the server")
	// CompanyID() is a getter exercised here to make sure it remains safe under contention
	// (a future refactor might add lazy initialization or caching).
	assert.Equal(t, "test-company", c.CompanyID())
}
