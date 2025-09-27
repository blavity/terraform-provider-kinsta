package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blavity/terraform-provider-kinsta/internal/client"
	"github.com/stretchr/testify/assert"
)

type mockWordPressSiteKinstaClient struct {
	client.KinstaClient
	companyID           string
	createWordPressSite func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error)
	getWordPressSite    func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error)
	deleteWordPressSite func(ctx context.Context, id string) (*client.DeleteWordPressSiteResponse, error)
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

func Test_resourceWordPressSiteCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			companyID: "test-company-id",
			createWordPressSite: func(ctx context.Context, req *client.CreateWordPressSiteRequest) (*client.CreateWordPressSiteResponse, error) {
				return &client.CreateWordPressSiteResponse{
					OperationID: "test-operation-id",
				}, nil
			},
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Region:      "us-central1",
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
		assert.Equal(t, "test-operation-id", d.Id())
	})

	t.Run("failed creation", func(t *testing.T) {
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
}

func Test_resourceWordPressSiteRead(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		mockClient := &mockWordPressSiteKinstaClient{
			getWordPressSite: func(ctx context.Context, id string) (*client.GetWordPressSiteResponse, error) {
				return &client.GetWordPressSiteResponse{
					Site: client.WordPressSite{
						ID:          "test-site-id",
						DisplayName: "Test Site",
						Region:      "us-central1",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceWordPressSite().Schema, map[string]interface{}{})
		d.SetId("test-site-id")

		diags := resourceWordPressSiteRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "Test Site", d.Get("display_name").(string))
		assert.Equal(t, "us-central1", d.Get("region").(string))
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
					Message: "WordPress site 'test-site-id' is being deleted",
				}, nil
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
