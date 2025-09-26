package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kinsta/terraform-provider-kinsta/internal/client"
	"github.com/stretchr/testify/assert"
)

type mockKinstaClient struct {
	client.KinstaClient
	companyID      string
	createDatabase func(ctx context.Context, req *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error)
	getDatabase    func(ctx context.Context, id string) (*client.GetDatabaseResponse, error)
	updateDatabase func(ctx context.Context, id string, req *client.UpdateDatabaseRequest) (*client.UpdateDatabaseResponse, error)
	deleteDatabase func(ctx context.Context, id string) (*client.DeleteDatabaseResponse, error)
}

func (m *mockKinstaClient) CompanyID() string {
	return m.companyID
}

func (m *mockKinstaClient) CreateDatabase(ctx context.Context, req *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error) {
	return m.createDatabase(ctx, req)
}

func (m *mockKinstaClient) GetDatabase(ctx context.Context, id string) (*client.GetDatabaseResponse, error) {
	return m.getDatabase(ctx, id)
}

func (m *mockKinstaClient) UpdateDatabase(ctx context.Context, id string, req *client.UpdateDatabaseRequest) (*client.UpdateDatabaseResponse, error) {
	return m.updateDatabase(ctx, id, req)
}

func (m *mockKinstaClient) DeleteDatabase(ctx context.Context, id string) (*client.DeleteDatabaseResponse, error) {
	return m.deleteDatabase(ctx, id)
}

func Test_resourceDatabaseCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			companyID: "test-company-id",
			createDatabase: func(ctx context.Context, req *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error) {
				return &client.CreateDatabaseResponse{
					Database: struct {
						ID string `json:"id"`
					}{
						ID: "test-database-id",
					},
				}, nil
			},
			getDatabase: func(ctx context.Context, id string) (*client.GetDatabaseResponse, error) {
				return &client.GetDatabaseResponse{
					Database: client.Database{
						ID:          "test-database-id",
						Name:        "test-database",
						DisplayName: "Test Database",
						Status:      "ready",
						Cluster: struct {
							ID          string `json:"id"`
							Location    string `json:"location"`
							DisplayName string `json:"display_name"`
						}{
							Location: "us-central1",
						},
						Type:         "postgresql",
						Version:      "15",
						ResourceType: "db1",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{
			"name":         "test-database",
			"display_name": "Test Database",
			"region":       "us-central1",
			"db_type":      "postgresql",
			"version":      "15",
			"size":         "db1",
		})

		diags := resourceDatabaseCreate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-database-id", d.Id())
	})

	t.Run("failed creation", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			companyID: "test-company-id",
			createDatabase: func(ctx context.Context, req *client.CreateDatabaseRequest) (*client.CreateDatabaseResponse, error) {
				return nil, errors.New("failed to create database")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{
			"name":         "test-database",
			"display_name": "Test Database",
			"region":       "us-central1",
			"db_type":      "postgresql",
			"version":      "15",
			"size":         "db1",
		})

		diags := resourceDatabaseCreate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceDatabaseRead(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			getDatabase: func(ctx context.Context, id string) (*client.GetDatabaseResponse, error) {
				return &client.GetDatabaseResponse{
					Database: client.Database{
						ID:          "test-database-id",
						Name:        "test-database",
						DisplayName: "Test Database",
						Status:      "ready",
						Cluster: struct {
							ID          string `json:"id"`
							Location    string `json:"location"`
							DisplayName string `json:"display_name"`
						}{
							Location: "us-central1",
						},
						Type:         "postgresql",
						Version:      "15",
						ResourceType: "db1",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{})
		d.SetId("test-database-id")

		diags := resourceDatabaseRead(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "test-database", d.Get("name").(string))
		assert.Equal(t, "Test Database", d.Get("display_name").(string))
		assert.Equal(t, "us-central1", d.Get("region").(string))
		assert.Equal(t, "postgresql", d.Get("db_type").(string))
		assert.Equal(t, "15", d.Get("version").(string))
		assert.Equal(t, "db1", d.Get("size").(string))
	})

	t.Run("failed read", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			getDatabase: func(ctx context.Context, id string) (*client.GetDatabaseResponse, error) {
				return nil, errors.New("failed to get database")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{})
		d.SetId("test-database-id")

		diags := resourceDatabaseRead(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceDatabaseUpdate(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		var updatedReq *client.UpdateDatabaseRequest
		mockClient := &mockKinstaClient{
			updateDatabase: func(ctx context.Context, id string, req *client.UpdateDatabaseRequest) (*client.UpdateDatabaseResponse, error) {
				updatedReq = req
				return &client.UpdateDatabaseResponse{
					Database: struct {
						ID          string `json:"id"`
						DisplayName string `json:"display_name"`
						Status      string `json:"status"`
					}{
						ID:          "test-database-id",
						DisplayName: "New Test Database",
						Status:      "updating",
					},
				}, nil
			},
			getDatabase: func(ctx context.Context, id string) (*client.GetDatabaseResponse, error) {
				return &client.GetDatabaseResponse{
					Database: client.Database{
						ID:          "test-database-id",
						Name:        "test-database",
						DisplayName: "New Test Database",
						Status:      "ready",
						Cluster: struct {
							ID          string `json:"id"`
							Location    string `json:"location"`
							DisplayName string `json:"display_name"`
						}{
							Location: "us-central1",
						},
						Type:         "postgresql",
						Version:      "15",
						ResourceType: "db2",
					},
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{
			"name":         "test-database",
			"display_name": "Test Database",
			"region":       "us-central1",
			"db_type":      "postgresql",
			"version":      "15",
			"size":         "db1",
		})
		d.SetId("test-database-id")
		d.Set("display_name", "New Test Database")
		d.Set("size", "db2")

		diags := resourceDatabaseUpdate(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.NotNil(t, updatedReq)
		assert.Equal(t, "New Test Database", updatedReq.DisplayName)
		assert.Equal(t, "db2", updatedReq.ResourceType)
	})

	t.Run("failed update", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			updateDatabase: func(ctx context.Context, id string, req *client.UpdateDatabaseRequest) (*client.UpdateDatabaseResponse, error) {
				return nil, errors.New("failed to update database")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{})
		d.SetId("test-database-id")
		d.Set("display_name", "New Test Database")

		diags := resourceDatabaseUpdate(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}

func Test_resourceDatabaseDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			deleteDatabase: func(ctx context.Context, id string) (*client.DeleteDatabaseResponse, error) {
				return &client.DeleteDatabaseResponse{
					Message: "Database 'test-database-id' is being deleted",
				}, nil
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{})
		d.SetId("test-database-id")

		diags := resourceDatabaseDelete(context.Background(), d, mockClient)

		assert.False(t, diags.HasError())
		assert.Equal(t, "", d.Id())
	})

	t.Run("failed deletion", func(t *testing.T) {
		mockClient := &mockKinstaClient{
			deleteDatabase: func(ctx context.Context, id string) (*client.DeleteDatabaseResponse, error) {
				return nil, errors.New("failed to delete database")
			},
		}

		d := schema.TestResourceDataRaw(t, resourceDatabase().Schema, map[string]interface{}{})
		d.SetId("test-database-id")

		diags := resourceDatabaseDelete(context.Background(), d, mockClient)

		assert.True(t, diags.HasError())
	})
}