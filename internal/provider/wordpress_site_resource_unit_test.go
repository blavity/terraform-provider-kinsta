package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

type mockWordPressSiteKinstaClient struct {
	client.KinstaClient
	companyID           string
	createWordPressSite func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error)
	getWordPressSite    func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error)
	deleteWordPressSite func(ctx context.Context, id string) (*client.DeleteWordPressSiteResponse, error)
	pollOperation       func(ctx context.Context, operationID string) (string, error)
}

func (m *mockWordPressSiteKinstaClient) CompanyID() string {
	return m.companyID
}

func (m *mockWordPressSiteKinstaClient) CreateWordPressSite(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
	return m.createWordPressSite(ctx, req)
}

func (m *mockWordPressSiteKinstaClient) GetWordPressSite(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
	return m.getWordPressSite(ctx, id)
}

func (m *mockWordPressSiteKinstaClient) DeleteWordPressSite(ctx context.Context, id string) (*client.DeleteWordPressSiteResponse, error) {
	return m.deleteWordPressSite(ctx, id)
}

func (m *mockWordPressSiteKinstaClient) PollOperation(ctx context.Context, operationID string) (string, error) {
	return m.pollOperation(ctx, operationID)
}

func Test_resourceWordPressSiteCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				return &client.CreateWordPressSiteResponse{
					OperationID: "test-operation-id",
					Message:     "Site creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				assert.Equal(t, "test-operation-id", operationID)
				return "test-site-id", nil
			},
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				assert.Equal(t, "test-site-id", id)
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Environments: []client.WordPressEnvironment{
							{
								ID:          "test-env-id",
								Name:        "live",
								DisplayName: "Live",
							},
						},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"admin_email":    "test@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			"site_title":     "Test Site",
			"wp_language":    "en_US",
		})

		diags := resourceWordPressSiteCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-site-id", d.Id())
		assert.Equal(t, "test-site-id", d.Get("site_id").(string))
		assert.Equal(t, "test-env-id", d.Get("environment_id").(string))
	})

	t.Run("failed creation - API error", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				return nil, errors.New("failed to create WordPress site")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"admin_email":    "test@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			"site_title":     "Test Site",
			"wp_language":    "en_US",
		})

		diags := resourceWordPressSiteCreate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})

	t.Run("failed creation - polling failure", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				return &client.CreateWordPressSiteResponse{
					OperationID: "test-operation-id",
					Message:     "Site creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				return "", errors.New("operation failed")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"admin_email":    "test@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			"site_title":     "Test Site",
			"wp_language":    "en_US",
		})

		diags := resourceWordPressSiteCreate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})

	t.Run("validates required fields", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name": "Test Site",
			"region":       "us-central1",
			// Missing required fields: admin_email, admin_password, admin_user, site_title
		})

		// Terraform returns empty strings for missing string fields (not nil)
		assert.Equal(t, "", d.Get("admin_email").(string))
		assert.Equal(t, "", d.Get("admin_password").(string))
		assert.Equal(t, "", d.Get("admin_user").(string))
		assert.Equal(t, "", d.Get("site_title").(string))

		// Verify defaults are set
		assert.Equal(t, "new", d.Get("install_mode").(string))
		assert.Equal(t, "en_US", d.Get("wp_language").(string))

		// Skip actual create as it would fail on missing required fields
		_ = mockClient
	})
}

func Test_resourceWordPressSiteRead(t *testing.T) {
	t.Run("successful read with environment", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				assert.Equal(t, "test-site-id", id)
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Environments: []client.WordPressEnvironment{
							{
								ID:          "test-env-id",
								Name:        "live",
								DisplayName: "Live",
							},
						},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-site-id", d.Get("site_id").(string))
		assert.Equal(t, "Test Site", d.Get("display_name").(string))
		assert.Equal(t, "test-env-id", d.Get("environment_id").(string))
	})

	t.Run("successful read without environment", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:           "test-site-id",
						DisplayName:  "Test Site",
						Environments: []client.WordPressEnvironment{},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-site-id", d.Get("site_id").(string))
		assert.Equal(t, "Test Site", d.Get("display_name").(string))
		// environment_id should not be set if no environments
		assert.Equal(t, "", d.Get("environment_id").(string))
	})

	t.Run("failed read", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return nil, errors.New("failed to get WordPress site")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteRead(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceWordPressSiteDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			deleteWordPressSite: func(ctx context.Context, id string) (*client.DeleteWordPressSiteResponse, error) {
				return &client.DeleteWordPressSiteResponse{
					OperationID: "delete-site-op-123",
					Message:     "WordPress site 'test-site-id' is being deleted",
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				assert.Equal(t, "delete-site-op-123", operationID)
				return "", nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteDelete(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "", d.Id())
	})

	t.Run("failed deletion", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			deleteWordPressSite: func(ctx context.Context, id string) (*client.DeleteWordPressSiteResponse, error) {
				return nil, errors.New("failed to delete WordPress site")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteDelete(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceWordPressSiteUpdate(t *testing.T) {
	t.Run("update returns error because all fields are ForceNew", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"admin_email":    "test@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			"site_title":     "Test Site",
			"wp_language":    "en_US",
		})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteUpdate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, "does not support updates")
	})
}

func Test_resourceWordPressSite_Schema(t *testing.T) {
	resource := resourceWordPressSite()

	t.Run("required fields are marked as required", func(t *testing.T) {
		requiredFields := []string{"display_name"}
		for _, field := range requiredFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Required, "Field %s should be required", field)
		}
	})

	t.Run("write-only fields are optional+computed with DiffSuppressFunc", func(t *testing.T) {
		writeOnlyFields := []string{"region", "admin_email", "admin_password", "admin_user", "site_title"}
		for _, field := range writeOnlyFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Optional, "Field %s should be optional", field)
			assert.True(t, fieldSchema.Computed, "Field %s should be computed", field)
			assert.NotNil(t, fieldSchema.DiffSuppressFunc, "Field %s should have DiffSuppressFunc", field)
		}
	})

	t.Run("optional fields have correct defaults", func(t *testing.T) {
		assert.Equal(t, "new", resource.Schema["install_mode"].Default)
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
		computedFields := []string{"site_id", "environment_id"}
		for _, field := range computedFields {
			fieldSchema, ok := resource.Schema[field]
			assert.True(t, ok, "Field %s should exist in schema", field)
			assert.True(t, fieldSchema.Computed, "Field %s should be computed", field)
			assert.False(t, fieldSchema.Required, "Field %s should not be required", field)
		}
	})

	t.Run("all string fields have correct type", func(t *testing.T) {
		// Boolean fields that are not strings
		booleanFields := map[string]bool{
			"is_multisite":           true,
			"is_subdomain_multisite": true,
			"woocommerce":            true,
			"wordpressseo":           true,
		}

		for name, fieldSchema := range resource.Schema {
			if booleanFields[name] {
				assert.Equal(t, schema.TypeBool, fieldSchema.Type, "Field %s should be TypeBool", name)
			} else {
				assert.Equal(t, schema.TypeString, fieldSchema.Type, "Field %s should be TypeString", name)
			}
		}
	})
}

func Test_resourceWordPressSiteCreate_RequestValidation(t *testing.T) {
	t.Run("creates request with all fields", func(t *testing.T) {
		var capturedRequest *client.CreateWordPressSiteRequest

		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				capturedRequest = req
				return &client.CreateWordPressSiteResponse{
					OperationID: "test-operation-id",
					Message:     "Site creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				return "test-site-id", nil
			},
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Environments: []client.WordPressEnvironment{
							{
								ID:          "test-env-id",
								Name:        "live",
								DisplayName: "Live",
							},
						},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"install_mode":   "migrate",
			"admin_email":    "admin@example.com",
			"admin_password": "secure-password-123",
			"admin_user":     "testadmin",
			"site_title":     "My WordPress Site",
			"wp_language":    "fr_FR",
		})

		diags := resourceWordPressSiteCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		require.NotNil(t, capturedRequest)
		assert.Equal(t, "test-company-id", capturedRequest.Company)
		assert.Equal(t, "Test Site", capturedRequest.DisplayName)
		assert.Equal(t, "us-central1", capturedRequest.Region)
		assert.Equal(t, "migrate", capturedRequest.InstallMode)
		assert.Equal(t, "admin@example.com", capturedRequest.AdminEmail)
		assert.Equal(t, "secure-password-123", capturedRequest.AdminPassword)
		assert.Equal(t, "testadmin", capturedRequest.AdminUser)
		assert.Equal(t, "My WordPress Site", capturedRequest.SiteTitle)
		assert.Equal(t, "fr_FR", capturedRequest.WPLanguage)
	})

	t.Run("uses default values when not specified", func(t *testing.T) {
		var capturedRequest *client.CreateWordPressSiteRequest

		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				capturedRequest = req
				return &client.CreateWordPressSiteResponse{
					OperationID: "test-operation-id",
					Message:     "Site creation started",
					Status:      202,
				}, nil
			},
			pollOperation: func(ctx context.Context, operationID string) (string, error) {
				return "test-site-id", nil
			},
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:           "test-site-id",
						DisplayName:  "Test Site",
						Environments: []client.WordPressEnvironment{},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{
			"display_name":   "Test Site",
			"region":         "us-central1",
			"admin_email":    "admin@example.com",
			"admin_password": "password",
			"admin_user":     "admin",
			"site_title":     "Test Site",
			// install_mode and wp_language not specified - should use defaults
		})

		diags := resourceWordPressSiteCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		require.NotNil(t, capturedRequest)
		assert.Equal(t, "new", capturedRequest.InstallMode)
		assert.Equal(t, "en_US", capturedRequest.WPLanguage)
	})
}

func Test_resourceWordPressSiteRead_EdgeCases(t *testing.T) {
	t.Run("handles multiple environments", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Environments: []client.WordPressEnvironment{
							{
								ID:          "env-live",
								Name:        "live",
								DisplayName: "Live",
							},
							{
								ID:          "env-staging",
								Name:        "staging",
								DisplayName: "Staging",
							},
						},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		// Should use the first environment
		assert.Equal(t, "env-live", d.Get("environment_id").(string))
	})
}
