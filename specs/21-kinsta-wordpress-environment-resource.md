# Spec: kinsta_wordpress_environment Resource

**Provider:** terraform-provider-kinsta  
**Resource Type:** Managed Resource  
**API Base:** https://api.kinsta.com/v2  
**Primary Endpoint:** POST /sites/environments  

---

## Overview

The `kinsta_wordpress_environment` resource manages WordPress environments (staging or premium staging) for an existing Kinsta WordPress site. Environments are additional instances of a WordPress site used for development, testing, or preview purposes before pushing changes to the live environment.

### Key Characteristics

- **Parent-Child Relationship**: Must reference an existing `kinsta_wordpress_site` via `site_id`
- **Async Operation**: Environment creation returns 202 + operation_id and requires polling
- **ForceNew Everything**: All fields are immutable; any change triggers replacement
- **Write-Only Fields**: Several fields (is_premium, admin credentials, site_title) cannot be read back from API
- **Environment ID Discovery**: Created environment ID is discovered by comparing site environments before/after creation
- **Import Support**: Environments can be imported using `site_id:env_id` format

---

## API Mapping

### Endpoints

| Operation | Method | Path | Response | Polling |
|-----------|--------|------|----------|---------|
| Create | POST | /sites/environments | 202 + operation_id | Required |
| Read | GET | /sites/{site_id}/environments | 200 (list) | No |
| Update | - | - | Not supported | - |
| Delete | DELETE | /sites/environments/{env_id} | 202 + operation_id | Required |

### Request Schema (POST /sites/environments)

```json
{
  "site_id": "string (required)",
  "display_name": "string (required)",
  "is_premium": "boolean (optional, default: false)",
  "install_mode": "string (optional: 'new' | 'clone' | 'none', default: 'clone')",
  "admin_email": "string (optional, required if install_mode='new')",
  "admin_user": "string (optional, required if install_mode='new')",
  "admin_password": "string (optional, required if install_mode='new')",
  "site_title": "string (optional)",
  "source_env_id": "string (optional, required if install_mode='clone')",
  "php_version": "string (optional: '8.0' | '8.1' | '8.2' | '8.3')",
  "wp_debug": "boolean (optional)",
  "wp_debug_display": "boolean (optional)",
  "wp_debug_log": "boolean (optional)"
}
```

### Response Schema (GET /sites/{site_id}/environments)

```json
{
  "company": {
    "environments": [
      {
        "id": "string (environment_id)",
        "display_name": "string",
        "is_blocked": "boolean",
        "is_premium": "boolean (NOT RETURNED - write-only)",
        "edge_caching": { "enabled": "boolean" },
        "blocked_ips": [],
        "container": {
          "id": "string",
          "display_name": "string",
          "site_path": "string"
        },
        "ssh_connection": {},
        "php_version": "string"
      }
    ]
  }
}
```

**Critical**: The response does NOT include:
- `is_premium` (write-only)
- `admin_email`, `admin_user`, `admin_password` (write-only)
- `site_title` (write-only)
- `wp_debug`, `wp_debug_display`, `wp_debug_log` (not exposed in list)

---

## Terraform Schema

### Arguments

| Name | Type | Required | ForceNew | Sensitive | Description |
|------|------|----------|----------|-----------|-------------|
| site_id | string | Yes | Yes | No | ID of parent WordPress site |
| display_name | string | Yes | Yes | No | Environment name (e.g., "staging", "premium-staging") |
| is_premium | bool | No | Yes | No | Whether this is a premium staging environment |
| install_mode | string | No | Yes | No | Installation mode: "new", "clone", "none" (default: "clone") |
| source_env_id | string | No | Yes | No | Source environment ID if install_mode="clone" |
| admin_email | string | No | Yes | Yes | WordPress admin email (required if install_mode="new") |
| admin_user | string | No | Yes | Yes | WordPress admin username (required if install_mode="new") |
| admin_password | string | No | Yes | Yes | WordPress admin password (required if install_mode="new") |
| site_title | string | No | Yes | No | WordPress site title |
| php_version | string | No | Yes | No | PHP version: "8.0", "8.1", "8.2", "8.3" |
| wp_debug | bool | No | Yes | No | Enable WP_DEBUG |
| wp_debug_display | bool | No | Yes | No | Enable WP_DEBUG_DISPLAY |
| wp_debug_log | bool | No | Yes | No | Enable WP_DEBUG_LOG |

### Attributes (Computed)

| Name | Type | Description |
|------|------|-------------|
| id | string | Environment ID (discovered after creation) |
| is_blocked | bool | Whether environment is blocked |
| ssh_connection_string | string | SSH connection details |
| container_id | string | Container ID |

### DiffSuppressFunc Pattern

For write-only fields (`is_premium`, `site_title`, `admin_*`), use:

```go
DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
    // Suppress diff if value already exists in state (after create/import)
    return old != ""
}
```

This prevents Terraform from detecting drift on fields that cannot be read back from the API.

---

## Implementation Details

### Create Flow

1. **Validate Dependencies**
   - Confirm install_mode is valid
   - If install_mode="new", require admin_email, admin_user, admin_password
   - If install_mode="clone", require source_env_id

2. **Capture Before State**
   ```go
   beforeEnvs := client.GetSiteEnvironments(site_id)
   ```

3. **Submit Create Request**
   - POST /sites/environments
   - Receive 202 + operation_id

4. **Poll Operation**
   - Call PollOperation(operation_id) until status = "success"
   - Handle 404 grace period (first 5-10 seconds may return 404)

5. **Discover Environment ID**
   ```go
   afterEnvs := client.GetSiteEnvironments(site_id)
   newEnv := findNewEnvironment(beforeEnvs, afterEnvs, display_name)
   d.SetId(newEnv.ID)
   ```

6. **Handle ID Discovery Failure**
   - If no new environment found with matching display_name, return error
   - This can happen if:
     - display_name is not unique
     - API delay in returning new environment
     - Operation succeeded but environment not yet visible

### Read Flow

1. **Fetch All Environments for Site**
   ```go
   envs := client.GetSiteEnvironments(site_id)
   ```

2. **Find Environment by ID**
   ```go
   env := findEnvironmentByID(envs, d.Id())
   if env == nil {
       d.SetId("") // Mark as deleted
       return nil
   }
   ```

3. **Set Readable Fields Only**
   - display_name
   - is_blocked
   - php_version
   - ssh_connection_string
   - container_id

4. **Leave Write-Only Fields Alone**
   - Do NOT call d.Set() for: is_premium, admin_*, site_title
   - These remain in state as originally configured

### Update Flow

**Not Supported** - All fields are ForceNew. Update function should:

```go
func resourceWordPressEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    return diag.Errorf("updates are not supported for kinsta_wordpress_environment; all changes require replacement")
}
```

### Delete Flow

1. **Submit Delete Request**
   - DELETE /sites/environments/{env_id}
   - Receive 202 + operation_id

2. **Poll Operation**
   - Call PollOperation(operation_id) until status = "success"

3. **Clear State**
   ```go
   d.SetId("")
   ```

### Import Flow

**Format:** `site_id:env_id`

Example: `terraform import kinsta_wordpress_environment.staging fbab4927-e354-4044-b226-29ac0fbd20ca:c84ce214-69b9-4a32-8e67-880672cf1d38`

```go
func resourceWordPressEnvironmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
    parts := strings.Split(d.Id(), ":")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid import format; expected site_id:env_id")
    }
    
    siteID := parts[0]
    envID := parts[1]
    
    d.Set("site_id", siteID)
    d.SetId(envID)
    
    // Trigger Read to populate remaining fields
    return []*schema.ResourceData{d}, nil
}
```

---

## Validation Rules

### Field Constraints

1. **display_name**
   - Non-empty string
   - Should be unique within site (API may not enforce, but helps with discovery)

2. **install_mode**
   - ValidateFunc: must be "new", "clone", or "none"
   - Default: "clone"

3. **Admin Credentials (Conditional)**
   - If install_mode="new": admin_email, admin_user, admin_password are required
   - Use ConflictsWith or custom validation

4. **source_env_id (Conditional)**
   - If install_mode="clone": source_env_id is required

5. **php_version**
   - ValidateFunc: must be "8.0", "8.1", "8.2", or "8.3"

### Cross-Field Validation

```go
if installMode := d.Get("install_mode").(string); installMode == "new" {
    if d.Get("admin_email").(string) == "" {
        return diag.Errorf("admin_email is required when install_mode is 'new'")
    }
    // Similar for admin_user, admin_password
}

if installMode := d.Get("install_mode").(string); installMode == "clone" {
    if d.Get("source_env_id").(string) == "" {
        return diag.Errorf("source_env_id is required when install_mode is 'clone'")
    }
}
```

---

## Error Handling

### Known Error Scenarios

| Error | Cause | Resolution |
|-------|-------|------------|
| 404 on operation | Operation not yet initialized | Retry with exponential backoff (5-30s) |
| Environment ID not found after poll | Non-unique display_name or API delay | Retry read or fail with actionable error |
| 403 Forbidden | Insufficient permissions | Check API key permissions |
| 409 Conflict | Environment name already exists | Return clear error suggesting unique name |
| Operation status = "failed" | API-side error | Return operation message to user |

### Eventual Consistency

- After operation completes, wait 2-5 seconds before calling GetSiteEnvironments
- If new environment not found, retry up to 3 times with 3-second delays

---

## Testing Plan

### Unit Tests

**File:** `internal/provider/wordpress_environment_resource_unit_test.go`

1. **Schema Validation**
   - Test required fields (site_id, display_name)
   - Test ForceNew enforcement on all fields
   - Test Sensitive marking on admin credentials
   - Test default values (install_mode)

2. **DiffSuppressFunc**
   - Test write-only fields don't generate diffs after initial create
   - Test that changes to write-only fields still trigger replacement if explicitly changed

3. **Import Format**
   - Test valid format: "site_id:env_id"
   - Test invalid format: single string, wrong delimiter

### Acceptance Tests

**File:** `internal/provider/wordpress_environment_resource_test.go`

**Pre-requisites:**
- Existing WordPress site (can use kinsta_wordpress_site in config)
- Valid API key with environment creation permissions

**Test Cases:**

1. **TestAcc_ResourceWordPressEnvironment_BasicStaging**
   - Create standard staging environment with install_mode="clone"
   - Verify computed fields (id, container_id)
   - Verify dependency on parent site

2. **TestAcc_ResourceWordPressEnvironment_PremiumStaging**
   - Create premium staging with is_premium=true
   - Verify environment creation succeeds

3. **TestAcc_ResourceWordPressEnvironment_NewInstall**
   - Create environment with install_mode="new"
   - Provide admin credentials
   - Verify successful creation

4. **TestAcc_ResourceWordPressEnvironment_CustomPHP**
   - Create environment with php_version="8.3"
   - Verify php_version is set correctly

5. **TestAcc_ResourceWordPressEnvironment_Import**
   - Create environment
   - Import using site_id:env_id format
   - Verify state matches

### Test Configuration Example

```hcl
resource "kinsta_wordpress_site" "test" {
  display_name    = "Test Site for Environments"
  region          = "us-central1"
  admin_email     = "admin@example.com"
  admin_password  = "SecureP@ss123"
  admin_user      = "admin"
  site_title      = "Test Site"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.test.site_id
  display_name = "staging"
  install_mode = "clone"
  source_env_id = kinsta_wordpress_site.test.environment_id
}

resource "kinsta_wordpress_environment" "premium" {
  site_id      = kinsta_wordpress_site.test.site_id
  display_name = "premium-staging"
  is_premium   = true
  install_mode = "new"
  admin_email  = "staging@example.com"
  admin_user   = "stagingadmin"
  admin_password = "StagingP@ss123"
  site_title   = "Premium Staging"
}
```

---

## Documentation Requirements

### Resource Doc (docs/resources/wordpress_environment.md)

Must include:
- Clear description of what environments are
- Parent-child relationship with wordpress_site
- Explanation of write-only fields and DiffSuppressFunc behavior
- install_mode scenarios with examples
- Import format and examples
- Limitations (no updates, ForceNew behavior)

### Example (examples/wordpress_environment/main.tf)

Must demonstrate:
- Basic staging environment
- Premium staging environment
- New install vs clone scenarios
- Dependency on wordpress_site

---

## Known Limitations

1. **No Update Support**
   - All fields are ForceNew
   - Any change triggers environment replacement
   - This matches Kinsta's API behavior

2. **Write-Only Fields Cannot Be Verified**
   - is_premium status not returned by API
   - Admin credentials not returned
   - DiffSuppressFunc prevents false drift detection

3. **Environment ID Discovery Requires Uniqueness**
   - display_name should be unique per site
   - Non-unique names may cause ID discovery failures
   - Document this as best practice

4. **Eventual Consistency Window**
   - New environment may not appear immediately in list
   - Implement retry logic with reasonable timeouts

5. **No Direct Environment Operations**
   - Push to live, clone, etc. not yet implemented
   - May be added as separate action resources in future

---

## Future Enhancements

1. **Environment Actions**
   - `kinsta_wordpress_environment_push` resource for staging → live
   - `kinsta_wordpress_environment_clear_cache` action
   - SFTP access management

2. **Computed Fields**
   - Add more container/SSH details
   - Cache settings
   - Performance metrics

3. **Validation Improvements**
   - Check source_env_id exists before creating clone
   - Validate display_name uniqueness

4. **Sweepers**
   - Automated cleanup for acceptance tests
   - Prevent orphaned test environments

---

## References

- Kinsta API Docs: https://api-docs.kinsta.com/tag/WordPress-Site-Environments
- Swagger Spec: ./swagger.json (paths: /sites/environments, /sites/{site_id}/environments)
- Operations Polling Contract: specs/02-operations-polling-contract.md
- WordPress Site Resource: specs/20-kinsta-wordpress-site-resource.md
