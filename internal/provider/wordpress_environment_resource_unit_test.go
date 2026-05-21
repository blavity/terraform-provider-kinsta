package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

type mockWordPressEnvironmentKinstaClient struct {
	client.KinstaClient
	companyID                  string
	createWordPressEnvironment func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error)
	deleteWordPressEnvironment func(ctx context.Context, siteID, envID string) (*client.DeleteWordPressEnvironmentResponse, error)
	getWordPressSite           func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error)
	pollOperation              func(ctx context.Context, operationID string) (string, error)
}

func (m *mockWordPressEnvironmentKinstaClient) CompanyID() string {
	return m.companyID
}

func (m *mockWordPressEnvironmentKinstaClient) CreateWordPressEnvironment(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
	return m.createWordPressEnvironment(ctx, siteID, req)
}

func (m *mockWordPressEnvironmentKinstaClient) DeleteWordPressEnvironment(ctx context.Context, siteID, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
	return m.deleteWordPressEnvironment(ctx, siteID, envID)
}

func (m *mockWordPressEnvironmentKinstaClient) GetWordPressSite(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
	if m.getWordPressSite != nil {
		return m.getWordPressSite(ctx, siteID)
	}
	return nil, errors.New("GetWordPressSite not implemented in mock")
}

func (m *mockWordPressEnvironmentKinstaClient) PollOperation(ctx context.Context, operationID string) (string, error) {
	return m.pollOperation(ctx, operationID)
}

func Test_resourceWordPressEnvironmentCreate_Standard(t *testing.T) {
	callCount := 0
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			callCount++
			if callCount == 1 {
				// First call - before environment creation (empty list)
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:           "test-site-id",
						Environments: []client.WordPressEnvironment{},
					},
				}, nil
			}
			// Second call - after environment creation (with new environment)
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: "test-site-id",
					Environments: []client.WordPressEnvironment{
						{
							ID:          "test-env-id",
							Name:        "staging",
							DisplayName: "staging",
						},
					},
				},
			}, nil
		},
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
	callCount := 0
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			callCount++
			if callCount == 1 {
				// First call - before environment creation (empty list)
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:           "test-site-id",
						Environments: []client.WordPressEnvironment{},
					},
				}, nil
			}
			// Second call - after environment creation (with new environment)
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: "test-site-id",
					Environments: []client.WordPressEnvironment{
						{
							ID:          "test-premium-env-id",
							Name:        "premium-staging",
							DisplayName: "Premium Staging",
						},
					},
				},
			}, nil
		},
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
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID:           "test-site-id",
					Environments: []client.WordPressEnvironment{},
				},
			}, nil
		},
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
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID:           "test-site-id",
					Environments: []client.WordPressEnvironment{},
				},
			}, nil
		},
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
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			assert.Equal(t, "test-site-id", siteID)
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: "test-site-id",
					Environments: []client.WordPressEnvironment{
						{
							ID:          "test-env-id",
							Name:        "staging",
							DisplayName: "Staging",
						},
					},
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
	// resourceWordPressEnvironmentRead calls GetWordPressSite (not
	// GetWordPressEnvironment) to find the env in the site's environments
	// list — the individual environment GET endpoint returns 404 for env
	// IDs that exist within a site. Failing GetWordPressSite is the only
	// way to exercise the Read error path. A previous version of this
	// test failed GetWordPressEnvironment, which is never called, so the
	// test passed only because of the mock's default fallback — not
	// because it covered the intended code path.
	wantErr := errors.New("failed to get WordPress site")
	mockClient := &mockWordPressEnvironmentKinstaClient{
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			return nil, wantErr
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("test-env-id")

	diags := resourceWordPressEnvironmentRead(context.Background(), d, mockClient)

	require.True(t, diags.HasError(), "Read must surface the underlying GetWordPressSite error")
	assert.Contains(t, diags[0].Summary, wantErr.Error(), "diagnostic must contain the underlying error string")
}

func Test_resourceWordPressEnvironmentDelete(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		deleteWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
			assert.Equal(t, "test-env-id", envID)
			return &client.DeleteWordPressEnvironmentResponse{
				OperationID: "delete-env-op-123",
				Message:     "WordPress environment 'test-env-id' is being deleted",
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			assert.Equal(t, "delete-env-op-123", operationID)
			return "", nil
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
		deleteWordPressEnvironment: func(ctx context.Context, siteID, envID string) (*client.DeleteWordPressEnvironmentResponse, error) {
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
		requiredFields := []string{"site_id", "display_name"}
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
		// wp_language has no default - it's optional with no default value
		wpLangSchema := resource.Schema["wp_language"]
		assert.True(t, wpLangSchema.Optional)
		assert.Nil(t, wpLangSchema.Default)
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

	t.Run("importer is wired", func(t *testing.T) {
		// Schema-level: just confirm an importer is registered. The
		// behavioral check that it's the custom site_id:env_id parser
		// (and not the SDK's passthrough — which would silently accept
		// "no-colon-here") lives in Test_resourceWordPressEnvironmentImport,
		// which exercises both the happy path and rejects malformed input.
		require.NotNil(t, resource.Importer, "Principle III requires terraform import to work")
		require.NotNil(t, resource.Importer.StateContext, "import state context must be set")
	})

	t.Run("timeouts are configured for create and delete", func(t *testing.T) {
		require.NotNil(t, resource.Timeouts)
		require.NotNil(t, resource.Timeouts.Create, "create timeout must be set (Principle III async polling)")
		require.NotNil(t, resource.Timeouts.Delete, "delete timeout must be set (Principle III async polling)")
	})

	// Principle II: pinning the exact set of fields the resource exposes.
	// Adding, removing, or renaming a field changes this list; the next
	// person touching it must update the assertion and explicitly state
	// the semver impact.
	t.Run("schema field set is locked", func(t *testing.T) {
		expected := []string{
			"site_id",
			"display_name",
			"site_title",
			"is_premium",
			"admin_email",
			"admin_password",
			"admin_user",
			"wp_language",
			"environment_id",
		}
		assert.Len(t, resource.Schema, len(expected),
			"unexpected number of schema fields; adding/removing fields is a Principle II event")
		for _, field := range expected {
			_, ok := resource.Schema[field]
			assert.True(t, ok, "expected schema field missing: %s", field)
		}
	})
}

func Test_resourceWordPressEnvironmentCreate_RequestValidation(t *testing.T) {
	t.Run("creates request with all fields", func(t *testing.T) {
		var capturedRequest *client.CreateWordPressEnvironmentRequest
		var capturedSiteID string
		callCount := 0

		mockClient := &mockWordPressEnvironmentKinstaClient{
			companyID: "test-company-id",
			getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
				callCount++
				if callCount == 1 {
					return &client.GetWordPressSiteResponse{
						Site: client.WordPressSite{
							ID:           siteID,
							Environments: []client.WordPressEnvironment{},
						},
					}, nil
				}
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID: siteID,
						Environments: []client.WordPressEnvironment{
							{
								ID:          "test-env-id",
								DisplayName: "premium-staging",
							},
						},
					},
				}, nil
			},
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
		callCount := 0

		mockClient := &mockWordPressEnvironmentKinstaClient{
			companyID: "test-company-id",
			getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
				callCount++
				if callCount == 1 {
					return &client.GetWordPressSiteResponse{
						Site: client.WordPressSite{
							ID:           siteID,
							Environments: []client.WordPressEnvironment{},
						},
					}, nil
				}
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID: siteID,
						Environments: []client.WordPressEnvironment{
							{
								ID:          "test-env-id",
								DisplayName: "staging",
							},
						},
					},
				}, nil
			},
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

func Test_resourceWordPressEnvironmentImport(t *testing.T) {
	t.Run("valid import ID", func(t *testing.T) {
		mockClient := &mockWordPressEnvironmentKinstaClient{
			getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID: siteID,
						Environments: []client.WordPressEnvironment{
							{ID: "test-env-id", Name: "staging", DisplayName: "Staging"},
						},
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
			"site_id": "test-site-id",
		})
		d.SetId("test-site-id:test-env-id")

		results, err := resourceWordPressEnvironmentImport(context.Background(), d, mockClient)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "test-env-id", results[0].Id())
		assert.Equal(t, "test-site-id", results[0].Get("site_id"))
	})

	t.Run("invalid import ID format", func(t *testing.T) {
		mockClient := &mockWordPressEnvironmentKinstaClient{}

		d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{})
		d.SetId("invalid-no-colon")

		_, err := resourceWordPressEnvironmentImport(context.Background(), d, mockClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})
}

// Read drift: when the environment's ID is no longer present in the site's
// environments list, the resource MUST clear its ID (Principle III) so
// Terraform re-creates instead of erroring on a phantom resource.
func Test_resourceWordPressEnvironmentRead_DriftClearsID(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: siteID,
					Environments: []client.WordPressEnvironment{
						// Different env IDs — our target env was deleted out of band.
						{ID: "other-env-id", Name: "live", DisplayName: "Live"},
					},
				},
			}, nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id": "test-site-id",
	})
	d.SetId("deleted-env-id")

	diags := resourceWordPressEnvironmentRead(context.Background(), d, mockClient)

	assert.False(t, diags.HasError(), "drift detection must not surface as an error")
	assert.Equal(t, "", d.Id(), "missing env must clear the resource ID")
}

// Eventual-consistency retry: after deleting an environment, the API may
// reject re-creation with "display name … already used" for ~30s. The
// resource retries with exponential backoff. This test verifies the retry
// path by failing the first attempt and succeeding the second.
func Test_resourceWordPressEnvironmentCreate_DisplayNameRetry(t *testing.T) {
	// Swap the package-level afterFunc with an instant-fire so the retry
	// loop doesn't burn real wall-clock seconds. Restored automatically.
	origAfter := afterFunc
	afterFunc = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { afterFunc = origAfter })

	createCallCount := 0
	siteCallCount := 0

	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			siteCallCount++
			if siteCallCount == 1 {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:           siteID,
						Environments: []client.WordPressEnvironment{},
					},
				}, nil
			}
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: siteID,
					Environments: []client.WordPressEnvironment{
						{ID: "new-env-id", Name: "staging", DisplayName: "staging"},
					},
				},
			}, nil
		},
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			createCallCount++
			if createCallCount == 1 {
				// First call: API still considers the display name reserved.
				return nil, errors.New("display name 'staging' is already used by another environment")
			}
			return &client.CreateWordPressEnvironmentResponse{
				OperationID: "env-op-id",
				Status:      202,
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			return "", nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":      "test-site-id",
		"display_name": "staging",
	})

	// 5s deadline keeps the test fast: first retry waits 1s (1<<0), second wait would be 2s.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	diags := resourceWordPressEnvironmentCreate(ctx, d, mockClient)

	assert.False(t, diags.HasError(), "retry should recover from eventual-consistency conflict")
	assert.Equal(t, 2, createCallCount, "should have retried exactly once")
	assert.Equal(t, "new-env-id", d.Id())
}

// When the after-create site listing returns no new environment ID (e.g., a
// stale read of the API), the resource MUST surface an explicit error rather
// than committing partial state.
func Test_resourceWordPressEnvironmentCreate_NoNewEnvID(t *testing.T) {
	mockClient := &mockWordPressEnvironmentKinstaClient{
		companyID: "test-company-id",
		getWordPressSite: func(ctx context.Context, siteID string) (*client.GetWordPressSiteResponse, error) {
			// Same environment list before and after creation → no new ID found.
			return &client.GetWordPressSiteResponse{
				Site: client.WordPressSite{
					ID: siteID,
					Environments: []client.WordPressEnvironment{
						{ID: "pre-existing-env", Name: "live"},
					},
				},
			}, nil
		},
		createWordPressEnvironment: func(ctx context.Context, siteID string, req *client.CreateWordPressEnvironmentRequest) (*client.CreateWordPressEnvironmentResponse, error) {
			return &client.CreateWordPressEnvironmentResponse{
				OperationID: "env-op-id",
				Status:      202,
			}, nil
		},
		pollOperation: func(ctx context.Context, operationID string) (string, error) {
			return "", nil
		},
	}

	d := schema.TestResourceDataRaw(t, resourceWordPressEnvironment().Schema, map[string]interface{}{
		"site_id":      "test-site-id",
		"display_name": "staging",
	})

	diags := resourceWordPressEnvironmentCreate(context.Background(), d, mockClient)

	require.True(t, diags.HasError(), "missing new env ID must surface an error, not silent success")
	assert.Contains(t, diags[0].Summary, "failed to identify newly created environment")
	assert.Equal(t, "", d.Id(), "no partial state when env ID can't be determined")
}
