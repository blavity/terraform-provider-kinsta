package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blavity/terraform-provider-kinsta/internal/client"
	"github.com/stretchr/testify/assert"
)

type mockApplicationKinstaClient struct {
	client.KinstaClient
	companyID         string
	createApplication func(ctx context.Context, req *client.CreateApplicationRequest) (*client.CreateApplicationResponse, error)
	getApplication    func(ctx context.Context, id string) (*client.GetApplicationResponse, error)
	updateApplication func(ctx context.Context, id string, req *client.UpdateApplicationRequest) (*client.UpdateApplicationResponse, error)
	deleteApplication func(ctx context.Context, id string) (*client.DeleteApplicationResponse, error)
}

func (m *mockApplicationKinstaClient) CompanyID() string {
	return m.companyID
}

func (m *mockApplicationKinstaClient) CreateApplication(ctx context.Context, req *client.CreateApplicationRequest) (*client.CreateApplicationResponse, error) {
	return m.createApplication(ctx, req)
}

func (m *mockApplicationKinstaClient) GetApplication(ctx context.Context, id string) (*client.GetApplicationResponse, error) {
	return m.getApplication(ctx, id)
}

func (m *mockApplicationKinstaClient) UpdateApplication(ctx context.Context, id string, req *client.UpdateApplicationRequest) (*client.UpdateApplicationResponse, error) {
	return m.updateApplication(ctx, id, req)
}

func (m *mockApplicationKinstaClient) DeleteApplication(ctx context.Context, id string) (*client.DeleteApplicationResponse, error) {
	return m.deleteApplication(ctx, id)
}

func Test_resourceApplicationCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			companyID: "test-company-id",
			createApplication: func(ctx context.Context, req *client.CreateApplicationRequest) (*client.CreateApplicationResponse, error) {
				return &client.CreateApplicationResponse{
					Application: struct {
						ID string `json:"id"`
					}{
						ID: "test-application-id",
					},
				}, nil
			},
			getApplication: func(ctx context.Context, id string) (*client.GetApplicationResponse, error) {
				return &client.GetApplicationResponse{
					Application: client.Application{
						ID:          "test-application-id",
						Name:        "test-application",
						DisplayName: "Test Application",
						Status:      "ready",
						Region:      "us-central1",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{
			"name":         "test-application",
			"display_name": "Test Application",
			"region":       "us-central1",
		})

		diags := resourceApplicationCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-application-id", d.Id())
	})

	t.Run("failed creation", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			companyID: "test-company-id",
			createApplication: func(ctx context.Context, req *client.CreateApplicationRequest) (*client.CreateApplicationResponse, error) {
				return nil, errors.New("failed to create application")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{
			"name":         "test-application",
			"display_name": "Test Application",
			"region":       "us-central1",
		})

		diags := resourceApplicationCreate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceApplicationRead(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			getApplication: func(ctx context.Context, id string) (*client.GetApplicationResponse, error) {
				return &client.GetApplicationResponse{
					Application: client.Application{
						ID:          "test-application-id",
						Name:        "test-application",
						DisplayName: "Test Application",
						Status:      "ready",
						Region:      "us-central1",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})
		d.SetId("test-application-id")

		diags := resourceApplicationRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-application", d.Get("name").(string))
		assert.Equal(t, "Test Application", d.Get("display_name").(string))
		assert.Equal(t, "us-central1", d.Get("region").(string))
	})

	t.Run("failed read", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			getApplication: func(ctx context.Context, id string) (*client.GetApplicationResponse, error) {
				return nil, errors.New("failed to get application")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})
		d.SetId("test-application-id")

		diags := resourceApplicationRead(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceApplicationUpdate(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		var updatedReq *client.UpdateApplicationRequest
		mockClient := &mockApplicationKinstaClient{
			updateApplication: func(ctx context.Context, id string, req *client.UpdateApplicationRequest) (*client.UpdateApplicationResponse, error) {
				updatedReq = req
				return &client.UpdateApplicationResponse{
					Application: struct {
						ID          string `json:"id"`
						DisplayName string `json:"display_name"`
						Status      string `json:"status"`
					}{
						ID:          "test-application-id",
						DisplayName: "New Test Application",
						Status:      "updating",
					},
				}, nil
			},
			getApplication: func(ctx context.Context, id string) (*client.GetApplicationResponse, error) {
				return &client.GetApplicationResponse{
					Application: client.Application{
						ID:          "test-application-id",
						Name:        "test-application",
						DisplayName: "New Test Application",
						Status:      "ready",
						Region:      "us-central1",
					},
				}, nil
			},
		}

		d := resourceApplication().TestResourceData()
		d.SetId("test-application-id")
		d.Set("display_name", "New Test Application")

		diags := resourceApplicationUpdate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.NotNil(t, updatedReq)
		assert.Equal(t, "New Test Application", updatedReq.DisplayName)
	})

	t.Run("failed update", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			updateApplication: func(ctx context.Context, id string, req *client.UpdateApplicationRequest) (*client.UpdateApplicationResponse, error) {
				return nil, errors.New("failed to update application")
			},
		}

		d := resourceApplication().TestResourceData()
		d.SetId("test-application-id")
		d.Set("display_name", "New Test Application")

		diags := resourceApplicationUpdate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceApplicationDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			deleteApplication: func(ctx context.Context, id string) (*client.DeleteApplicationResponse, error) {
				return &client.DeleteApplicationResponse{
					Message: "Application 'test-application-id' is being deleted",
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})
		d.SetId("test-application-id")

		diags := resourceApplicationDelete(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "", d.Id())
	})

	t.Run("failed deletion", func(t *testing.T) {
		mockClient := &mockApplicationKinstaClient{
			deleteApplication: func(ctx context.Context, id string) (*client.DeleteApplicationResponse, error) {
				return nil, errors.New("failed to delete application")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})
		d.SetId("test-application-id")

		diags := resourceApplicationDelete(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}