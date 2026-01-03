# Spec: kinsta_wordpress_site Resource

**Resource:** `kinsta_wordpress_site`  
**API Endpoints:**
- `POST /sites` (async, returns operation_id)
- `GET /sites/{site_id}` (sync)
- `DELETE /sites/{site_id}` (async, returns operation_id)
- `PUT /sites` - **Does NOT exist** (no update support)

**Evidence:**
- `swagger.json#/paths/~1sites/post` (deprecated=false)
- `swagger.json#/paths/~1sites~1{site_id}/get` (deprecated=false)
- `swagger.json#/paths/~1sites~1{site_id}/delete` (deprecated=false)

---

## Schema Mapping

### Input Fields (Required)

| Terraform Field | API Field | Type | Required | ForceNew | Sensitive |
|----------------|-----------|------|----------|----------|-----------|
| `display_name` | `display_name` | string | Yes | Yes | No |
| `region` | `region` | string | Yes | Yes | No |
| `admin_email` | `admin_email` | string | Yes | Yes | Yes |
| `admin_password` | `admin_password` | string | Yes | Yes | Yes |
| `admin_user` | `admin_user` | string | Yes | Yes | No |
| `site_title` | `site_title` | string | Yes | Yes | No |

### Input Fields (Optional)

| Terraform Field | API Field | Type | Default | ForceNew | Notes |
|----------------|-----------|------|---------|----------|-------|
| `install_mode` | `install_mode` | string | `"new"` | Yes | Values: `"new"`, `"plain"`, `"clone"` |
| `wp_language` | `wp_language` | string | `"en_US"` | Yes | WordPress locale code |
| `is_multisite` | `is_multisite` | bool | `false` | Yes | Enable WordPress Multisite |
| `is_subdomain_multisite` | `is_subdomain_multisite` | bool | `false` | Yes | Subdomain vs subdirectory multisite |
| `woocommerce` | `woocommerce` | bool | `false` | Yes | Install WooCommerce plugin |
| `wordpressseo` | `wordpressseo` | bool | `false` | Yes | Install Yoast SEO plugin |

**Note:** The four new fields (`is_multisite`, `is_subdomain_multisite`, `woocommerce`, `wordpressseo`) are **write-only**. They are accepted in POST /sites but NOT returned in GET /sites/{site_id}. Therefore:
- They must be Optional + ForceNew
- They should NOT be set in Read() function
- DiffSuppressFunc not needed (no read drift)

### Computed Outputs

| Terraform Field | API Field | Type | Description |
|----------------|-----------|------|-------------|
| `site_id` | `site.id` | string | WordPress site ID (also used as resource ID) |
| `environment_id` | `site.environments[0].id` | string | Primary environment ID |

**Evidence:** `swagger.json#/components/schemas/SiteById-Site/properties/id` and `/environments`

---

## Lifecycle Behavior

### Create
1. Call `POST /sites` with all input fields → Returns `{"operation_id": "..."}`
2. Poll `GET /operations/{operation_id}` until status = "200 OK"
3. Extract `site_id` from polling result (see polling contract)
4. Call `GET /sites/{site_id}` to populate computed fields
5. Set `d.SetId(site_id)`

**Async Contract:** See `specs/02-operations-polling-contract.md` for polling details. Operation data is opaque; implement lookup-after-poll if site_id not reliably returned.

### Read
1. Call `GET /sites/{d.Id()}`
2. If 404 → `d.SetId("")` (resource deleted externally)
3. Map response fields to schema:
   - `site_id` ← `site.id`
   - `environment_id` ← `site.environments[0].id`
   - **DO NOT** set the four write-only fields (not in response)

### Update
**NOT SUPPORTED.** All fields are ForceNew. If Update is called, return error:
```go
return diag.Errorf("kinsta_wordpress_site does not support updates; all fields are immutable")
```

### Delete
1. Call `DELETE /sites/{d.Id()}` → Returns `{"operation_id": "..."}`
2. Poll `GET /operations/{operation_id}` until complete
3. 404 errors during polling are acceptable (resource already gone)

### Import
Support `terraform import kinsta_wordpress_site.example <site_id>`
- StateContext: `schema.ImportStatePassthroughContext`

---

## Validation Rules

### Field Constraints
- `display_name`: Must be unique per company (enforced by API)
- `region`: Must be valid Kinsta region (e.g., `"us-central1"`, `"europe-west2"`)
- `install_mode`: Enum: `"new"`, `"plain"`, `"clone"`
- `admin_email`: Must be valid email format
- `wp_language`: Must be valid WordPress locale (e.g., `"en_US"`, `"es_ES"`)
- `is_subdomain_multisite`: Only meaningful if `is_multisite = true`

### API Validation
The API will reject invalid values. No client-side validation needed beyond Terraform schema types.

---

## Error Handling

### Common Errors
| Error | HTTP Code | Handling |
|-------|-----------|----------|
| Duplicate display_name | 400 | Return clear error to user |
| Invalid region | 400 | Return clear error to user |
| Quota exceeded | 403 | Return clear error to user |
| Site not found | 404 | In Read: `d.SetId("")` |
| Operation polling timeout | N/A | Return error with operation_id |

---

## Test Plan

### Unit Tests (`wordpress_site_resource_unit_test.go`)
- [x] Schema validation (existing)
- [ ] Create request struct marshaling with new fields
- [ ] Read response unmarshaling (verify write-only fields not set)
- [ ] Error handling for 404 in Read

### Acceptance Tests (`wordpress_site_resource_test.go`)
- [x] Basic create/read/delete cycle (existing)
- [ ] Create with `is_multisite = true`
- [ ] Create with `woocommerce = true`
- [ ] Create with `wordpressseo = true`
- [ ] Create with `install_mode = "plain"`
- [ ] Import existing site
- [ ] Verify all fields ForceNew (change triggers replacement)

**Test Helpers Needed:**
- Random name generation (avoid conflicts)
- Pre-check for valid credentials
- Cleanup/sweeper for orphaned test resources

---

## Documentation

### docs/resources/wordpress_site.md
- Description and use cases
- Complete argument reference (including new fields)
- Complete attribute reference
- Example configurations:
  - Basic WordPress site
  - Multisite with WooCommerce
  - Clone from existing
- Import instructions
- Timeout configuration

### examples/wordpress_site/
- `main.tf` - Basic example
- `multisite.tf` - Multisite example
- `variables.tf` - Input variables
- `outputs.tf` - Output values

---

## Implementation Checklist

- [x] Resource exists at `internal/provider/wordpress_site_resource.go`
- [ ] Add 4 new schema fields with ForceNew=true
- [ ] Update `internal/client/wordpress.go` structs with new fields
- [ ] Update Create to pass new fields to API
- [ ] Verify Read does NOT set new fields (write-only)
- [ ] Ensure Update returns error (not supported)
- [ ] Update unit tests
- [ ] Add acceptance tests for new fields
- [ ] Update documentation
- [ ] Create examples

---

**Status:** In Progress  
**Last Updated:** 2026-01-02  
**Next Review:** After implementation complete
