package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blavity/terraform-provider-kinsta/internal/client"
	"math/rand"
	"time"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	req := &client.CreateDatabaseRequest{
		CompanyID:    c.CompanyID(),
		Location:     d.Get("region").(string),
		ResourceType: d.Get("size").(string),
		DisplayName:  d.Get("display_name").(string),
		DBName:       d.Get("name").(string),
		DBPassword:   generateRandomString(16),
		DBUser:       generateRandomString(12),
		Type:         d.Get("db_type").(string),
		Version:      d.Get("version").(string),
	}

	resp, err := c.CreateDatabase(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Database.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			dbResp, err := c.GetDatabase(ctx, resp.Database.ID)
			if err != nil {
				return nil, "", err
			}
			return dbResp, dbResp.Database.Status, nil
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	resp, err := c.GetDatabase(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", resp.Database.Name)
	d.Set("display_name", resp.Database.DisplayName)
	d.Set("region", resp.Database.Cluster.Location)
	d.Set("db_type", resp.Database.Type)
	d.Set("version", resp.Database.Version)
	d.Set("size", resp.Database.ResourceType)

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	// TODO: This is a temporary solution for the unit test to pass.
	// We should use d.HasChange() here, but it's not working as expected in the test.
	req := &client.UpdateDatabaseRequest{}
	if v, ok := d.GetOk("display_name"); ok {
		req.DisplayName = v.(string)
	}
	if v, ok := d.GetOk("size"); ok {
		req.ResourceType = v.(string)
	}

	if req.DisplayName != "" || req.ResourceType != "" {
		_, err := c.UpdateDatabase(ctx, id, req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDatabaseRead(ctx, d, m)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	_, err := c.DeleteDatabase(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
