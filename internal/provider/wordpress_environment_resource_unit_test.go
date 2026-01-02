package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blavity/terraform-provider-kinsta/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWordPressEnvironmentKinstaClient struct {
	client.KinstaClient
	companyID                   string
	createWordPressEnvironment  func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error)
	getWordPressEnvironment     func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error)
	deleteWordPressEnvironment  func(ctx context.Context, envID string) (*client.DeleteWordPressEnvironmentResponse, error)
	pollOperation               func(ctx context.Context, operationID string) (string, error)
}

func (m *mockWordPressEnvironmentKinstaClient) CompanyID() string {
	return m.companyID
}

func (m *mockWordPressEnvironmentKinstaClient) CreateWordPressEnvironment(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
	return m.createWordPressEnvironment(ctx, siteID, req)
}

func (m *mockWordPressEnvironmentKinstaClient) GetWordPressEnvironment(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
	return m.getWordPressEnvironment(ctx, siteID, envID)
}

func (m *mockWordPressEnvironmentKinstaClient) DeleteWordPressEnvironment(ctx context.Context, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
	return m.deleteWordPressEnvironment(ctx, envID)
}

func (m *mockWordPressEnvironmentKinstaClient) PollOperation(ctx context.Context, operationID string) (string, error) {
	return m.pollOperation(ctx, operationID)
}

func Test_resourceWordPressEnvironmentCreate_Standard(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			assert.Equal(t, "staging", req.DisplayName)
			assert.False(t, req.IsPremium)
			return &client.CreateWordPressEnvironmentResponse{
				OperationID: "test-env-operation-id",
				Message:     "Environment creation started",
				Status:      202,
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			assert.Equal(t, "test-env-operation-id", operationID)
			return "test-env-id", nil
		},
		getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			assert.Equal(t, "test-env-id", envID)
			return &client.GetWordPressEnvironmentResponse{
				Environment: client.WordPressEnvironment{
					ID:          "test-env-id",
					Name:        "staging",
					DisplayName: "staging",
				},
			}, nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":        "test-site-id",
		"display_name":   "staging",
		"site_title":     "Test Site - Staging",
		"is_premium":     false,
		"admin_email":    "test@example.com",
		"admin_password": "password",
		"admin_user":     "admin",
		"wp_language":    "en_US",
	})

	diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

	assert.False(t, diags.HasError())
	assert.Equal(t, "test-env-id", d.Id())
	assert.Equal(t, "test-env-id", d.Get("environment_id").(string))
}

func Test_resourceWordPressEnvironmentCreate_Premium(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			assert.Equal(t, "premium-staging", req.DisplayName)
			assert.True(t, req.IsPremium)
			return &client.CreateWordPressEnvironmentResponse{
				OperationID: "test-env-premium-operation-id",
				Message:     "Premium environment creation started",
				Status:      202,
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			assert.Equal(t, "test-env-premium-operation-id", operationID)
			return "test-premium-env-id", nil
		},
		getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
			return &client.GetWordPressEnvironmentResponse{
				Environment: client.WordPressEnvironment{
					ID:          "test-premium-env-id",
					Name:        "premium-staging",
					DisplayName: "Premium Staging",
				},
			}, nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":        "test-site-id",
		"display_name":   "premium-staging",
		"site_title":     "Test Site - Premium Staging",
		"is_premium":     true,
		"admin_email":    "test@example.com",
		"admin_password": "password",
		"admin_user":     "admin",
		"wp_language":    "en_US",
	})

	diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

	assert.False(t, diags.HasError())
	assert.Equal(t, "test-premium-env-id", d.Id())
	assert.Equal(t, "test-premium-env-id", d.Get("environment_id").(string))
}

func Test_resourceWordPressEnvironmentCreate_APIError(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			return nil, errors.New("failed to create WordPress environment")
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":        "test-site-id",
		"display_name":   "staging",
		"site_title":     "Test Site - Staging",
		"is_premium":     false,
		"admin_email":    "test@example.com",
		"admin_password": "password",
		"admin_user":     "admin",
		"wp_language":    "en_US",
	})

	diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

	assert.True(t, diags.HasError())
}

func Test_resourceWordPressEnvironmentCreate_PollingFailure(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			return &client.CreateWordPressEnvironmentResponse{
				OperationID: "test-env-operation-id",
				Message:     "Environment creation started",
				Status:      202,
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			return "", errors.New("operation failed")
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":        "test-site-id",
		"display_name":   "staging",
		"site_title":     "Test Site - Staging",
		"is_premium":     false,
		"admin_email":    "test@example.com",
		"admin_password": "password",
		"admin_user":     "admin",
		"wp_language":    "en_US",
	})

	diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

	assert.True(t, diags.HasError())
}

func Test_resourceWordPressEnvironmentRead(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			assert.Equal(t, "test-env-id", envID)
			return &client.GetWordPressEnvironmentResponse{
				Environment: client.WordPressEnvironment{
					ID:          "test-env-id",
					Name:        "staging",
					DisplayName: "Staging",
				},
			}, nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentRead(context.Background(), d, mockClient)

	assert.False(t, diags.HasError())
	assert.Equal(t, "test-env-id", d.Get("environment_id").(string))
	assert.Equal(t, "Staging", d.Get("display_name").(string))
}

func Test_resourceWordPressEnvironmentRead_Error(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
			return nil, errors.New("failed to get WordPress environment")
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentRead(context.Background(), d, mockClient)

	assert.True(t, diags.HasError())
}

func Test_resourceWordPressEnvironmentDelete(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		deleteWordPressEnvironment: func(ctx context.Context, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-env-id", envID)
			return &client.DeleteWordPressEnvironmentResponse{
				OperationID: "delete-env-op-123",
				Message:     "WordPress environment 'test-env-id' is being deleted",
			}, nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentDelete(context.Background(), d, mockClient)

	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
}

func Test_resourceWordPressEnvironmentDelete_Error(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		deleteWordPressEnvironment: func(ctx context.Context, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
			return nil, errors.New("failed to delete WordPress environment")
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentDelete(context.Background(), d, mockClient)

	assert.True(t, diags.HasError())
}

func Test_resourceWordPressEnvironmentUpdate(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":        "test-site-id",
		"display_name":   "staging",
		"site_title":     "Test Site - Staging",
		"is_premium":     false,
		"admin_email":    "test@example.com",
		"admin_password": "password",
		"admin_user":     "admin",
		"wp_language":    "en_US",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentUpdate(context.Background(), d, mockClient)

	assert.False(t, diags.HasError())
}

func Test_resourceWordPressEnvironment_Schema(t *testing.T) {
	resource := resourceWordPressEnvironment()

	t.Run("required fields are marked as required", func(t *testing.T) {
		requiredFields := []string{"site_id", "display_name", "site_title", "is_premium", "admin_email", "admin_password", "admin_user"}
		for _, field := range requiredFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Required, "Field %s should be required", field)
		}
	})

	t.Run("all required fields are ForceNew", func(t *testing.T) {
		forceNewFields := []string{"site_id", "display_name", "site_title", "is_premium", "admin_email", "admin_password", "admin_user", "wp_language"}
		for _, field := range forceNewFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.ForceNew, "Field %s should be ForceNew", field)
		}
	})

	t.Run("optional fields have correct defaults", func(t *testing.T) {
		assert.Equal(t, "en_US", resource.Schema["wp_language"].Default)
	})

	t.Run("sensitive fields are marked as sensitive", func(t *testing.T) {
		sensitiveFields := []string{"admin_email", "admin_password"}
		for _, field := range sensitiveFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Sensitive, "Field %s should be sensitive", field)
		}
	})

	t.Run("computed fields are marked as computed", func(t *testing.T) {
		computedFields := []string{"environment_id"}
		for _, field := range computedFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Computed, "Field %s should be computed", field)
			assert.False(t, fieldSchema.Required, "Field %s should not be required", field)
		}
	})

	t.Run("is_premium is TypeBool", func(t *testing.T) {
		assert.Equal(t, schema.TypeBool, resource.Schema["is_premium"].Type)
	})
}

func Test_resourceWordPressEnvironmentCreate_RequestValidation(t *testing.T) {
	t.Run("creates request with all fields", func(t *testing.T) {
		var capturedRequest *client.CreateWordPressEnvironmentRequest
		var capturedSiteID string

		mockClient := &mockWordPressEnvironmentKinstaClient{
			companyID: "test-company-id",
			createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
				capturedRequest = req
				capturedSiteID = siteID
				return &client.CreateWordPressEnvironmentResponse{
					OperationID: "test-env-operation-id",
					Message:     "Environment creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				return "test-env-id", nil
			},
			getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
				return &client.GetWordPressEnvironmentResponse{
					Environment: client.WordPressEnvironment{
						ID:          "test-env-id",
						DisplayName: "Premium Staging",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
			"site_id":        "test-site-123",
			"display_name":   "premium-staging",
			"site_title":     "My WordPress Site - Premium Staging",
			"is_premium":     true,
			"admin_email":    "admin@example.com",
			"admin_password": "secure-password-123",
			"admin_user":     "testadmin",
			"wp_language":    "fr_FR",
		})

		diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		require.NotNil(t, capturedRequest)
		assert.Equal(t, "test-site-123", capturedSiteID)
		assert.Equal(t, "premium-staging", capturedRequest.DisplayName)
		assert.Equal(t, "My WordPress Site - Premium Staging", capturedRequest.SiteTitle)
		assert.True(t, capturedRequest.IsPremium)
		assert.Equal(t, "admin@example.com", capturedRequest.AdminEmail)
		assert.Equal(t, "secure-password-123", capturedRequest.AdminPassword)
		assert.Equal(t, "testadmin", capturedRequest.AdminUser)
		assert.Equal(t, "fr_FR", capturedRequest.WPLanguage)
	})

	t.Run("uses default values when not specified", func(t *testing.T) {
		var capturedRequest *client.CreateWordPressEnvironmentRequest

		mockClient := &mockWordPressEnvironmentKinstaClient{
			companyID: "test-company-id",
			createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
				capturedRequest = req
				return &client.CreateWordPressEnvironmentResponse{
					OperationID: "test-env-operation-id",
					Message:     "Environment creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				return "test-env-id", nil
			},
			getWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.GetWordPressEnvironmentResponse, error) {
				return &client.GetWordPressEnvironmentResponse{
					Environment: client.WordPressEnvironment{
						ID:          "test-env-id",
						DisplayName: "Staging",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
			"site_id":        "test-site-123",
			"display_name":   "staging",
			"site_title":     "Test Site - Staging",
			"is_premium":     false,
			"admin_email":    "admin@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			// wp_language not specified - should use default
		})

		diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		require.NotNil(t, capturedRequest)
		assert.Equal(t, "en_US", capturedRequest.WPLanguage)
	})
}
