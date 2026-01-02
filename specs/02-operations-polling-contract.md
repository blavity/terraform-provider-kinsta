# Operations Polling Contract (Kinsta Provider Only)

**Scope:** terraform-provider-kinsta asynchronous WordPress operations  
**API:** MyKinsta `https://api.kinsta.com/v2`  
**Evidence:** swagger.json v1.87.0  
**Status:** Implementation specification

---

## Overview

This document defines the contract for polling asynchronous operations in the Kinsta provider. Operations polling is required ONLY for WordPress site and environment lifecycle operations which return 202 status with an `operation_id`.

**Evidence of async operations:**
- `swagger.json#/paths/~1sites/post/responses` = `["202", "400", "401", "500"]`
- `swagger.json#/paths/~1sites~1{site_id}/delete/responses` = `["202", "401", "404", "500"]`
- `swagger.json#/paths/~1sites~1{site_id}~1environments/post/responses` = `["202", "400", "401", "500"]`

**Sevalla operations are synchronous:**
- `sevalla.openapi.json#/paths/~1databases/post/responses` = `["200", "401", "404", "500"]`
- No polling required for Sevalla provider resources

---

## Response Schema

### Async Operation Initiation (202)

```json
{
  "operation_id": "sites:add-54fb80af-576c-4fdc-ba4f-b596c83f15a1",
  "message": "Adding site in progress",
  "status": 202
}
```

**Evidence:** `swagger.json#/components/schemas/addWPSite-Response`

**Required fields:**
- `operation_id` (string) - Unique identifier for polling
- `message` (string) - Human-readable status
- `status` (number) - HTTP status code (202)

### Operation Status Response

**Endpoint:** `GET /operations/{operation_id}`

**Evidence:** `swagger.json#/paths/~1operations~1{operation_id}/get/responses`

**Response codes:**
- `200` - Operation completed successfully
- `202` - Operation still in progress
- `404` - Operation not found (may occur during initialization)
- `500` - Operation failed

**Response schema (all codes):**
```json
{
  "status": 200,  // or 202, 404, 500
  "message": "Successfully finished request",
  "data": {}  // OPAQUE - see below
}
```

**Evidence:** 
- `swagger.json#/components/schemas/StatusResponseSchema200`
- `swagger.json#/components/schemas/OperationResponse`
- `swagger.json#/components/schemas/StatusResponseSchema404`
- `swagger.json#/components/schemas/StatusResponseSchema500`

---

## Critical: operation.data is OPAQUE

**Evidence:** `swagger.json#/components/schemas/OperationResponse/properties/data` = `{}`

The `data` field is defined as an empty object `{}` in the OpenAPI spec. This means:

**MUST NOT assume:**
- `data.idSite` exists
- `data.idEnv` exists
- Any typed structure within `data`
- Consistent key names across operations

**Observed behavior (not guaranteed):**
- Site creation MAY include `data.idSite`
- Environment creation typically does NOT include `data.idEnv`
- Keys and structure can change without notice

**Implementation requirement:**
Treat `data` as completely opaque. Use lookup-after-poll strategy instead of relying on data extraction.

---

## Polling Strategy

### Retry Schedule

**Initial interval:** 2 seconds  
**Backoff:** Exponential with cap  
**Sequence:** 2s, 4s, 8s, 15s, 30s, 30s, 30s, ...  
**Maximum interval:** 30 seconds  
**Total timeout:** 10 minutes (configurable)

**Rationale:**
- Quick response for fast operations (2s initial)
- Reduces API load for long operations (30s cap)
- Balances responsiveness with politeness

### 404 Grace Window

**Problem:** Operations may return 404 during initialization phase.

**Evidence from API docs:**
> "Due to a delay in the operation initialization process, the site creation operation can return `404 Not Found` in the first few seconds."

**Solution:** Retry 404 responses for first 30 seconds (6 attempts × 5s).

**Implementation:**
```
attempt = 0
loop:
  response = GET /operations/{operation_id}
  
  if response.status == 404 and attempt < 6:
    wait 5 seconds
    attempt++
    continue loop
  
  if response.status == 404 and attempt >= 6:
    return error "operation not found after 30s grace period"
  
  # Handle 200, 202, 500 normally
```

### Timeout Configuration

**Default:** 10 minutes (600 seconds)  
**Configurable via:**
- Resource timeouts block:
  ```hcl
  resource "kinsta_wordpress_site" "example" {
    # ...
    timeouts {
      create = "15m"
      delete = "10m"
    }
  }
  ```

**Maximum attempts calculation:**
```
grace_attempts = 6 (30 seconds @ 5s interval)
polling_attempts = (timeout - 30s) / exponential_schedule

Example for 10 minute timeout:
grace: 6 attempts (30s)
polling: ~18 attempts (570s with backoff)
total: ~24 attempts
```

### Context Cancellation

**Requirements:**
- Check `ctx.Done()` before each HTTP request
- Check `ctx.Done()` during sleep/wait periods
- Return `ctx.Err()` immediately on cancellation

**Implementation:**
```go
func PollOperation(ctx context.Context, operationID string) error {
    // ... 404 grace period with ctx checks
    
    for attempt := 0; attempt < maxAttempts; attempt++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(interval):
            // continue polling
        }
        
        resp, err := c.GetOperation(ctx, operationID)
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        // ... handle response
    }
}
```

---

## Resource ID Lookup Strategy

### Problem

Since `operation.data` is opaque, resource IDs must be obtained after polling completes.

### WordPress Site Creation

**Strategy:** Use `data.idSite` if present (observed behavior), otherwise list-and-match.

**Implementation:**
```go
// 1. Create site (returns operation_id)
createResp, err := client.CreateSite(ctx, req)
operationID := createResp.OperationID

// 2. Poll until complete
err = client.PollOperation(ctx, operationID)

// 3. Attempt data extraction (optimistic)
if siteID, ok := operationData["idSite"].(string); ok {
    return siteID, nil
}

// 4. Fallback: list sites and match by display_name + recent created_at
sites, err := client.ListSites(ctx)
for _, site := range sites {
    if site.DisplayName == req.DisplayName && 
       site.CreatedAt > operationStartTime {
        return site.ID, nil
    }
}

return "", errors.New("site created but ID not found")
```

**Requirements:**
- `display_name` SHOULD be unique (document this constraint)
- Fallback adds ~2 seconds to create operation
- Log warning if fallback used

### WordPress Environment Creation

**Strategy:** Before/after environment list comparison (current implementation is correct).

**Implementation:**
```go
// 1. Get existing environment IDs (before)
beforeResp, err := client.GetSite(ctx, siteID)
existingEnvIDs := map[string]bool{}
for _, env := range beforeResp.Site.Environments {
    existingEnvIDs[env.ID] = true
}

// 2. Create environment (returns operation_id)
createResp, err := client.CreateEnvironment(ctx, siteID, req)
operationID := createResp.OperationID

// 3. Poll until complete
err = client.PollOperation(ctx, operationID)

// 4. Get environment IDs (after)
afterResp, err := client.GetSite(ctx, siteID)

// 5. Find new environment
for _, env := range afterResp.Site.Environments {
    if !existingEnvIDs[env.ID] {
        return env.ID, nil
    }
}

return "", errors.New("environment created but ID not found")
```

**Requirements:**
- No reliance on `data.idEnv`
- Assumes environment list is immediately consistent after operation complete
- Handles concurrent environment creations on same site

### WordPress Site/Environment Deletion

**Strategy:** No ID needed (deleting by known ID).

**Implementation:**
```go
// 1. Delete (returns operation_id)
deleteResp, err := client.DeleteSite(ctx, siteID)

// 2. Poll until complete
err = client.PollOperation(ctx, deleteResp.OperationID)

// 3. No lookup needed - deletion confirmed
return nil
```

---

## Error Handling

### Operation Failed (500)

**Response:**
```json
{
  "status": 500,
  "message": "Error occurred while processing your request",
  "data": {}
}
```

**Handling:**
- Return diagnostic error to user
- Include operation ID for support
- Do NOT retry (terminal failure)

**Example error:**
```
Error: operation sites:add-12345 failed: Error occurred while processing your request

This is a server-side failure. Please check:
1. Operation ID: sites:add-12345
2. MyKinsta dashboard for details
3. Contact Kinsta support if issue persists
```

### Operation Timeout

**Trigger:** Polling exceeds configured timeout without 200 or 500.

**Handling:**
- Return diagnostic error
- Inform user operation may still be running
- Provide manual verification steps

**Example error:**
```
Error: operation sites:add-12345 timed out after 10m0s

The operation may still be running. To check status:
1. Visit MyKinsta dashboard
2. Or: curl -H "Authorization: Bearer $TOKEN" https://api.kinsta.com/v2/operations/sites:add-12345
3. Consider increasing timeout: timeouts { create = "15m" }
```

### 404 After Grace Period

**Trigger:** Operation returns 404 beyond 30-second grace window.

**Handling:**
- Return error indicating operation not found
- Possible causes: invalid operation_id, operation expired

**Example error:**
```
Error: operation sites:add-12345 not found after 30s grace period

Possible causes:
1. Invalid operation_id (check API response)
2. Operation expired (operations have 24h TTL)
3. API error during operation creation
```

---

## Progress Logging

**Requirements:**
- Log at INFO level during polling
- Include elapsed time and attempt number
- Use terraform-plugin-log for structured logging

**Example log output:**
```
[INFO] Polling operation sites:add-12345 (attempt 1, elapsed 2s)
[INFO] Polling operation sites:add-12345 (attempt 2, elapsed 6s)
[INFO] Polling operation sites:add-12345 (attempt 3, elapsed 14s)
[INFO] Operation sites:add-12345 completed successfully (elapsed 18s)
```

**Implementation:**
```go
import "github.com/hashicorp/terraform-plugin-log/tflog"

startTime := time.Now()
for attempt := 1; attempt <= maxAttempts; attempt++ {
    elapsed := time.Since(startTime)
    tflog.Info(ctx, "Polling operation", map[string]interface{}{
        "operation_id": operationID,
        "attempt":      attempt,
        "elapsed":      elapsed.String(),
    })
    
    // ... poll logic
}
```

---

## Test Expectations

### Unit Tests

**MUST test:**

1. **404 grace period:**
   - Mock 5× 404, then 202, then 200
   - Assert no error, successful completion
   - Mock 7× 404, assert error "not found after grace period"

2. **Exponential backoff:**
   - Mock 5× 202, then 200
   - Assert wait times: 2s, 4s, 8s, 15s, 30s
   - Assert 30s cap maintained

3. **Timeout:**
   - Mock 100× 202 (never completes)
   - Set timeout to 10s
   - Assert error "operation timed out"

4. **Context cancellation:**
   - Start polling, cancel context after 5s
   - Assert returns ctx.Err()
   - Assert no panic, clean termination

5. **Opaque data:**
   - Mock 200 response with empty data: {}
   - Assert no panic
   - Assert fallback to lookup-after-poll

6. **Status 500:**
   - Mock 202, then 500
   - Assert error includes operation_id
   - Assert diagnostic message

### Acceptance Tests

**MUST test:**

1. **Site creation:**
   - Create WordPress site
   - Assert operation polls successfully
   - Assert site ID obtained
   - Assert site exists in state

2. **Environment creation:**
   - Create environment on existing site
   - Assert before/after comparison works
   - Assert environment ID obtained
   - Assert environment in parent site's environments list

3. **Deletion:**
   - Delete site
   - Assert polling completes
   - Assert site no longer exists (404 on GET)

4. **Timeout configuration:**
   - Create site with custom timeout (2m)
   - Mock slow operation
   - Assert respects custom timeout

**Test infrastructure:**
```go
func TestPollOperation_404GracePeriod(t *testing.T) {
    // Mock HTTP responses
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 6 {
            w.WriteHeader(404)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": 404,
                "message": "Operation not found",
            })
            return
        }
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": 200,
            "message": "Success",
            "data": map[string]interface{}{},
        })
    }))
    defer server.Close()
    
    client := NewClient(server.URL, "test-key")
    err := client.PollOperation(context.Background(), "test-op")
    
    assert.NoError(t, err)
    assert.Equal(t, 6, attempts) // 5× 404, 1× 200
}
```

---

## Implementation Checklist

Before merging polling code:

- [ ] 404 grace period implemented (6 attempts × 5s)
- [ ] Exponential backoff implemented (2s → 30s cap)
- [ ] Context cancellation checked in all wait loops
- [ ] operation.data treated as opaque
- [ ] Lookup-after-poll strategy implemented for sites
- [ ] Before/after comparison implemented for environments
- [ ] Progress logging via terraform-plugin-log
- [ ] Error messages include operation_id
- [ ] Unit tests cover all error cases
- [ ] Unit tests verify retry schedule
- [ ] Acceptance tests verify end-to-end operation
- [ ] Documentation updated with timeout configuration
- [ ] display_name uniqueness constraint documented

---

## Known Issues & Workarounds

### Issue: Environment display_name conflicts after deletion

**Problem:** Environment display_name may be reserved for ~30 seconds after deletion due to eventual consistency.

**Symptom:** Create environment immediately after deleting one with same display_name → API error "display name already used".

**Workaround:** Exponential retry on display_name conflict errors.

**Implementation:**
```go
maxRetries := 6 // 2^6 = 64 seconds max
for attempt := 0; attempt <= maxRetries; attempt++ {
    resp, err := client.CreateEnvironment(ctx, siteID, req)
    if err == nil {
        return resp, nil
    }
    
    if isDisplayNameConflict(err) && attempt < maxRetries {
        waitTime := time.Duration(1<<uint(attempt)) * time.Second
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(waitTime):
            continue
        }
    }
    
    return nil, err
}
```

**Status:** Documented in current wordpress_environment_resource.go implementation.

---

## References

- MyKinsta API Spec: `./swagger.json` v1.87.0
- Current implementation: `internal/client/client.go` (PollOperation function)
- WordPress site resource: `internal/provider/wordpress_site_resource.go`
- WordPress environment resource: `internal/provider/wordpress_environment_resource.go`
- ADR: `specs/00-adr-provider-split.md`

---

## Revision History

- 2026-01-01: Initial specification based on swagger.json analysis and current implementation review

---

## Approval

This specification must be validated against:
1. Current PollOperation implementation in `internal/client/client.go`
2. Unit tests in `internal/client/client_test.go`
3. Real API behavior (acceptance tests with TF_ACC=1)

**Next steps:**
1. Review polling logic against this contract
2. Add missing unit tests (opaque data, timeout, cancellation)
3. Update progress logging to use terraform-plugin-log
4. Document display_name uniqueness constraint
