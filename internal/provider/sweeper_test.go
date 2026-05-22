package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

// Sweepers are registered for every resource type that creates real upstream
// state during acceptance tests. They run when the suite is invoked with
// `go test -sweep=<region>` and reap any orphan whose display_name starts
// with testAccNamePrefix — produced by acctest.RandomWithPrefix in #64.
//
// The Kinsta API isn't sharded by region, so the SDK-required region argument
// is ignored. Run with any non-empty value, e.g. `-sweep=global`.

func init() {
	resource.AddTestSweepers("kinsta_wordpress_environment", &resource.Sweeper{
		Name: "kinsta_wordpress_environment",
		F:    sweepWordPressEnvironments,
	})
	resource.AddTestSweepers("kinsta_wordpress_site", &resource.Sweeper{
		Name: "kinsta_wordpress_site",
		// Environments must be reaped before their parent sites so this
		// sweeper sees consistent state. Deleting a site does cascade to its
		// envs server-side, but listing all eligible orphans first means we
		// don't depend on that behavior.
		Dependencies: []string{"kinsta_wordpress_environment"},
		F:            sweepWordPressSites,
	})
}

// sweepClient builds a Kinsta client from the same env vars the rest of
// the acceptance suite uses. Returns an error (not a fatal) because
// sweepers are invoked outside testing.T.
func sweepClient() (client.KinstaClient, error) {
	apiKey := os.Getenv("KINSTA_API_KEY")
	companyID := os.Getenv("KINSTA_COMPANY_ID")
	if apiKey == "" || companyID == "" {
		return nil, fmt.Errorf("KINSTA_API_KEY and KINSTA_COMPANY_ID must be set to run sweepers")
	}
	return client.New(apiKey, companyID), nil
}

func sweepWordPressSites(_ string) error {
	c, err := sweepClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	resp, err := c.GetWordPressSites(ctx)
	if err != nil {
		return fmt.Errorf("listing WordPress sites: %w", err)
	}

	var firstErr error
	for _, site := range resp.Company.Sites {
		if !strings.HasPrefix(site.DisplayName, testAccNamePrefix) {
			continue
		}
		if _, err := c.DeleteWordPressSite(ctx, site.ID); err != nil {
			// 404 = already gone (e.g., parallel sweep, manual cleanup,
			// or the env sweeper just deleted the parent). Idempotent
			// behavior keeps the sweep working under contention.
			if client.IsNotFound(err) {
				continue
			}
			// Collect and continue: a single failure shouldn't strand other
			// orphans waiting for a future sweep run.
			if firstErr == nil {
				firstErr = fmt.Errorf("deleting site %s (%s): %w", site.DisplayName, site.ID, err)
			}
		}
	}
	return firstErr
}

func sweepWordPressEnvironments(_ string) error {
	c, err := sweepClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	// Site-list responses don't include nested envs, so we fan out via
	// GetWordPressSite — but only for sites that already look like test
	// orphans. This keeps the fan-out proportional to test debris rather
	// than to total company size and avoids touching production sites
	// that happen to live in the same company.
	sitesResp, err := c.GetWordPressSites(ctx)
	if err != nil {
		return fmt.Errorf("listing WordPress sites: %w", err)
	}

	var firstErr error
	for _, listSite := range sitesResp.Company.Sites {
		if !strings.HasPrefix(listSite.DisplayName, testAccNamePrefix) {
			continue
		}
		siteResp, err := c.GetWordPressSite(ctx, listSite.ID)
		if err != nil {
			// Site already gone (e.g., parallel sweep or manual cleanup):
			// nothing to do for its envs.
			if client.IsNotFound(err) {
				continue
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("getting site %s: %w", listSite.ID, err)
			}
			continue
		}
		for _, env := range siteResp.Site.Environments {
			if !strings.HasPrefix(env.DisplayName, testAccNamePrefix) {
				continue
			}
			// The implicit "live" environment can't be deleted independently
			// of its site — leave it for the site sweeper to handle.
			if env.Name == "live" {
				continue
			}
			if _, err := c.DeleteWordPressEnvironment(ctx, env.ID); err != nil {
				if client.IsNotFound(err) {
					continue
				}
				if firstErr == nil {
					firstErr = fmt.Errorf("deleting env %s (%s): %w", env.DisplayName, env.ID, err)
				}
			}
		}
	}
	return firstErr
}
