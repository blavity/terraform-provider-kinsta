package provider

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// seedWriteOnlyFromConfig copies user-declared values out of the raw HCL
// config and into state for write-only fields the upstream MyKinsta API
// does not return on GET.
//
// Why this exists:
//   After `terraform import kinsta_wordpress_site.x <id>`, state contains
//   only the fields Kinsta's GET endpoint surfaces (`id`, `site_id`,
//   `display_name`, `environment_id`). Every write-only attribute on the
//   schema — admin_*, site_title, wp_language, region, install_mode, the
//   multisite/WooCommerce/SEO booleans — is absent. Combined with
//   ForceNew on every one of them, the next `terraform plan` proposes
//   the resource for destroy+recreate, with each write-only field
//   showing as "+ X = \"Y\" # forces replacement". Applying that plan
//   would delete the just-imported live Kinsta site.
//
//   The fix is the standard hashicorp/aws pattern: during Read, pull
//   the config-declared values out of d.GetRawConfig() and write them
//   into state for any field where state is still empty/default. On the
//   first post-import plan the DiffSuppressFunc (`return old != ""` on
//   strings) sees the seeded value, suppresses the diff, and the plan
//   reports the resource as in-sync.
//
//   stringFields: state == "" is the "seed me" signal — preserves any
//     value the API actually returned, only fills empties.
//   boolFields: always copy config when present. Bool defaults are all
//     `false`, so overwriting `false` with another `false` is a no-op,
//     and overwriting `false` with a config-declared `true` is the
//     exact post-import correction we want.
//
//   During import-time Read (Importer.StateContext flow) there is no
//   config yet — d.GetRawConfig() returns a null cty.Value. The
//   IsNull() guard at the top short-circuits in that case; seeding
//   happens on the subsequent plan-time Read once the user has written
//   the resource block.
func seedWriteOnlyFromConfig(d *schema.ResourceData, stringFields, boolFields []string) error {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return nil
	}

	for _, name := range stringFields {
		if d.Get(name).(string) != "" {
			// State already has a value (from a previous create/apply or from
			// the API on this Read). Don't overwrite — preserves the user's
			// ability to clear a field intentionally (which would still
			// trigger ForceNew, the expected behavior).
			continue
		}
		v := raw.GetAttr(name)
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		if v.Type() != cty.String {
			continue
		}
		if err := d.Set(name, v.AsString()); err != nil {
			return err
		}
	}

	for _, name := range boolFields {
		v := raw.GetAttr(name)
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		if v.Type() != cty.Bool {
			continue
		}
		if err := d.Set(name, v.True()); err != nil {
			return err
		}
	}

	return nil
}
