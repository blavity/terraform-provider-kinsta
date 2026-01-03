# Phase 4: Acceptance Testing - Kinsta Provider

**Status:** ✅ COMPLETE  
**Date:** 2026-01-03  
**Provider:** terraform-provider-kinsta (MyKinsta API)

---

## Objective

Implement comprehensive acceptance tests for the two core WordPress resources (`kinsta_wordpress_site` and `kinsta_wordpress_environment`) following Terraform SDK v2 testing patterns.

---

## Deliverables

### 1. Acceptance Tests Created

#### ✅ `kinsta_wordpress_site` Acceptance Tests

**File:** `internal/provider/wordpress_site_resource_test.go`

**Test Cases:**
- `TestAcc_ResourceWordPressSite_Basic` - Basic site creation with standard settings
- `TestAcc_ResourceWordPressSite_CustomLanguage` - Site with custom language (fr_FR)
- `TestAcc_ResourceWordPressSite_MigrateMode` - Site created in migrate mode

**Coverage:**
- All required fields (display_name, region, admin credentials, site_title)
- Optional fields (wp_language, install_mode)
- Computed fields verification (site_id, environment_id)
- Existence check function (testAccCheckWordPressSiteExists)

#### ✅ `kinsta_wordpress_environment` Acceptance Tests

**File:** `internal/provider/wordpress_environment_resource_test.go`

**Test Cases:**
- `TestAcc_ResourceWordPressEnvironment_Basic` - Standard environment creation
- `TestAcc_ResourceWordPressEnvironment_Premium` - Premium environment with credentials
- `TestAcc_ResourceWordPressEnvironment_CustomSettings` - Custom PHP and debug settings

**Coverage:**
- Required fields (site_id, display_name)
- Premium vs non-premium environments
- Optional fields (php_version, wp_debug flags, admin credentials)
- Computed fields verification (id, site_id)
- Existence check function (testAccCheckWordPressEnvironmentExists)
- Dependency on kinsta_wordpress_site resource

### 2. Test Infrastructure

#### Provider Test Factories

**File:** `internal/provider/provider_test.go`

```go
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"kinsta": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}
```

#### Pre-Check Function

**File:** `internal/provider/database_resource_test.go`

```go
func testAccPreCheck(t *testing.T) {
	// Pre-check logic for acceptance tests
}
```

**Note:** Pre-check is defined but currently minimal. Future enhancement could validate:
- Environment variables (KINSTA_API_KEY, KINSTA_COMPANY_ID)
- API connectivity
- Required permissions

---

## Testing Patterns Established

### 1. Test Structure

All acceptance tests follow the standard SDK v2 pattern:

```go
func TestAcc_ResourceName_Scenario(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigString,
				Check: resource.ComposeTestCheckFunc(
					// Existence checks
					// Attribute checks
				),
			},
		},
	})
}
```

### 2. Configuration as Constants

HCL configurations are defined as constants:

```go
const testAccResourceWordPressSiteConfig = `
provider "kinsta" {
  # Environment variables expected
}

resource "kinsta_wordpress_site" "test" {
  display_name = "Test Site"
  # ... other fields
}
`
```

### 3. Existence Check Functions

Custom check functions verify resource state:

```go
func testAccCheckWordPressSiteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		
		// Additional field checks
		return nil
	}
}
```

### 4. Standard Test Checks

Using SDK test check functions:

```go
resource.ComposeTestCheckFunc(
	testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
	resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "display_name", "Test Site"),
	resource.TestCheckResourceAttrSet("kinsta_wordpress_site.test", "site_id"),
)
```

---

## Running Acceptance Tests

### Environment Setup

```bash
export TF_ACC=1
export KINSTA_API_KEY="your-api-key"
export KINSTA_COMPANY_ID="your-company-id"
```

### Run All Acceptance Tests

```bash
go test ./internal/provider -v -timeout 30m
```

### Run Specific Test

```bash
# WordPress Site tests
TF_ACC=1 go test ./internal/provider -run TestAcc_ResourceWordPressSite -v

# WordPress Environment tests
TF_ACC=1 go test ./internal/provider -run TestAcc_ResourceWordPressEnvironment -v

# Specific test case
TF_ACC=1 go test ./internal/provider -run TestAcc_ResourceWordPressSite_Basic -v
```

---

## Test Coverage Summary

### kinsta_wordpress_site

| Scenario | Test Name | Coverage |
|----------|-----------|----------|
| Basic creation | TestAcc_ResourceWordPressSite_Basic | Core functionality, required fields, computed outputs |
| Custom language | TestAcc_ResourceWordPressSite_CustomLanguage | Optional wp_language field, different region |
| Migrate mode | TestAcc_ResourceWordPressSite_MigrateMode | install_mode variations |

**Fields Tested:**
- Required: display_name, region, admin_email, admin_password, admin_user, site_title
- Optional: wp_language, install_mode
- Computed: site_id, environment_id

### kinsta_wordpress_environment

| Scenario | Test Name | Coverage |
|----------|-----------|----------|
| Basic environment | TestAcc_ResourceWordPressEnvironment_Basic | Standard environment, non-premium |
| Premium environment | TestAcc_ResourceWordPressEnvironment_Premium | Premium flag, admin credentials |
| Custom settings | TestAcc_ResourceWordPressEnvironment_CustomSettings | PHP version, debug flags |

**Fields Tested:**
- Required: site_id, display_name
- Optional: is_premium, admin_email, admin_password, php_version, wp_debug, wp_debug_display, wp_debug_log
- Computed: id

**Dependencies Tested:**
- Environment creation depends on site creation
- site_id reference from kinsta_wordpress_site

---

## Known Limitations

### 1. Real API Calls

Acceptance tests make real API calls to Kinsta. This means:
- Tests require valid API credentials
- Tests create real resources (may incur costs)
- Tests depend on API availability and rate limits
- Tests require cleanup (manual or automated sweepers)

### 2. No Sweepers Yet

Resource sweepers (cleanup) not yet implemented. Consider adding:

```go
func init() {
	resource.AddTestSweepers("kinsta_wordpress_site", &resource.Sweeper{
		Name: "kinsta_wordpress_site",
		F:    sweepWordPressSites,
	})
}
```

### 3. Minimal Pre-Check

Current `testAccPreCheck` function is empty. Should validate:
- Required environment variables
- API connectivity
- Valid credentials

### 4. No Update Tests

Current tests only cover Create and Read operations. Future additions:
- Update tests (though most fields are ForceNew)
- Import tests
- Destroy tests with verification

### 5. Limited Error Scenarios

Tests focus on happy paths. Should add:
- Invalid input tests
- Conflict scenarios
- Quota/limit tests
- Network error handling

---

## Future Enhancements

### Priority 1 (Next Sprint)

1. **Add Resource Sweepers**
   - Implement cleanup for test resources
   - Prevent accumulation of test data
   - Handle orphaned resources

2. **Enhanced Pre-Check**
   - Validate environment variables
   - Test API connectivity
   - Check account permissions

3. **Import Tests**
   ```go
   {
       ResourceName:      "kinsta_wordpress_site.test",
       ImportState:       true,
       ImportStateVerify: true,
       ImportStateVerifyIgnore: []string{"admin_password"},
   }
   ```

### Priority 2 (Future)

4. **Error Scenario Tests**
   - Invalid configuration tests
   - API error handling tests
   - Rate limit behavior tests

5. **Update/Modify Tests**
   - Test ForceNew behavior on immutable fields
   - Test update on mutable fields (if any exist)

6. **Parallel Test Execution**
   - Enable t.Parallel() where safe
   - Avoid resource name conflicts
   - Use random name generation

---

## Verification Checklist

- [x] All test files compile without errors
- [x] Tests skip appropriately without TF_ACC
- [x] Test configurations are valid HCL
- [x] Existence check functions validate required fields
- [x] Tests verify both required and optional fields
- [x] Tests verify computed fields are set
- [x] Tests follow SDK v2 patterns consistently
- [x] Test names follow convention: TestAcc_Resource{Name}_{Scenario}
- [x] Each test has meaningful assertions
- [x] Tests document expected environment variables

---

## Testing Against Real API

### Before Running Tests

1. **Set environment variables:**
   ```bash
   export TF_ACC=1
   export KINSTA_API_KEY="your-key"
   export KINSTA_COMPANY_ID="your-company-id"
   ```

2. **Verify API access:**
   ```bash
   curl -H "Authorization: Bearer $KINSTA_API_KEY" \
        https://api.kinsta.com/v2/sites
   ```

3. **Review quotas/limits:**
   - Check site creation limits
   - Check environment limits per site
   - Check rate limits

### After Running Tests

1. **Manual cleanup (until sweepers implemented):**
   ```bash
   # List test resources via API or Kinsta dashboard
   # Delete test sites manually
   ```

2. **Review logs for failures:**
   ```bash
   go test ./internal/provider -v -timeout 30m 2>&1 | tee test.log
   ```

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Acceptance Tests

on:
  push:
    branches: [main]
  pull_request:

jobs:
  acceptance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Acceptance Tests
        env:
          TF_ACC: 1
          KINSTA_API_KEY: ${{ secrets.KINSTA_API_KEY }}
          KINSTA_COMPANY_ID: ${{ secrets.KINSTA_COMPANY_ID }}
        run: |
          go test ./internal/provider -v -timeout 30m -run TestAcc
```

---

## Conclusion

Phase 4 establishes a solid foundation for acceptance testing:

✅ **Complete:**
- Comprehensive test cases for both WordPress resources
- Consistent testing patterns following SDK v2 conventions
- Multiple scenarios per resource
- Existence verification functions
- Documentation of testing approach

🔄 **Ready for:**
- Phase 5: Additional WordPress resources (domains, backups, SFTP)
- Continuous integration
- Real API validation
- Provider publication

📋 **Next Steps:**
- Implement resource sweepers for cleanup
- Add import tests
- Enhance pre-check validation
- Add error scenario tests
- Set up CI/CD pipeline

---

**Phase Status:** ✅ COMPLETE  
**Next Phase:** Phase 5 - Additional WordPress Resources (Domains, Backups, SFTP)  
**Blocked By:** None  
**Ready For:** Provider publication, real-world usage
