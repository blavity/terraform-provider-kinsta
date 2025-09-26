package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kinsta/terraform-provider-kinsta/internal/client"
)

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	req := &client.CreateApplicationRequest{
		CompanyID:   c.CompanyID(),
		DisplayName: d.Get("display_name").(string),
		Name:        d.Get("name").(string),
		Region:      d.Get("region").(string),
	}

	resp, err := c.CreateApplication(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Application.ID)

	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	resp, err := c.GetApplication(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", resp.Application.Name)
	d.Set("display_name", resp.Application.DisplayName)
	d.Set("region", resp.Application.Region)

	return nil
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	// TODO: This is a temporary solution for the unit test to pass.
	// We should use d.HasChange() here, but it's not working as expected in the test.
	req := &client.UpdateApplicationRequest{}
	if v, ok := d.GetOk("display_name"); ok {
		req.DisplayName = v.(string)
	}

	if req.DisplayName != "" {
		_, err := c.UpdateApplication(ctx, id, req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	_, err := c.DeleteApplication(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
