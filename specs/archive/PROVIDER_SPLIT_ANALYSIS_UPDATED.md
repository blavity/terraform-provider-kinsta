# Terraform Provider Split Analysis: Kinsta → Kinsta + Sevalla

**Date:** 2026-01-01  
**Updated:** With Sevalla API Spec v1.80.0  
**Scope:** Split terraform-provider-kinsta into two providers based on API boundaries

---

## Executive Summary

### Current State
- **Implementation:** 3 resources (database, wordpress_site, wordpress_environment)
- **Framework:** terraform-plugin-sdk/v2
- **Test Coverage:** Unit tests (100% for implemented), Acceptance tests (partial)
- **Kinsta API Coverage:** ~5% of WordPress endpoints
- **Sevalla API Coverage:** 0%

### Split Boundary (CONFIRMED)
1. **terraform-provider-kinsta (MyKinsta API):** WordPress sites ONLY
   - Base URL: https://api.kinsta.com/v2
   - Scope: WordPress sites, environments, domains, tools, backups, SFTP
   - Operations endpoint shared across both APIs
   
2. **terraform-provider-sevalla (Sevalla API):** Applications, databases, static sites
   - Base URL: https://api.sevalla.com/v2
   - Scope: Applications, databases, static sites, pipelines
   - Operations endpoint shared across both APIs

### Key Findings

**⚠️ CRITICAL DISCOVERY:** Sevalla API spec shows **IDENTICAL** database schemas to MyKinsta deprecated endpoint:
- Same request fields (company_id, location, resource_type, display_name, db_name, db_password, db_user, type, version)
- Same response format (synchronous 200 with immediate database.id)
- Same update capabilities (resource_type, display_name only)
- **Migration is straightforward** - only base URL changes

**✅ CONFIRMED:** 
- Both APIs use Bearer authentication (same API key)
- Operations polling endpoint exists at same path on both APIs
- Database resource can be migrated with minimal code changes

---

## 1. Sevalla API Analysis (NEW)

### 1.1 API Overview

**Version:** 1.80.0  
**Base URL:** https://api.sevalla.com/v2  
**Authentication:** Bearer token (same as MyKinsta)  
**Total Endpoints:** 60  
**Total Schemas:** 183  

### 1.2 Endpoint Categories

**Applications (18 endpoints):**
- GET /applications (list)
- GET /applications/{id}, /applications/{name}
- PUT /applications/{id} (update)
- DELETE /applications/{id}
- POST /applications/{id}/internal-connections
- POST /applications/{id}/cdn/toggle-status
- POST /applications/{id}/edge-cache/toggle-status
- POST /applications/{id}/clear-cache
- GET /applications/{id}/metrics/* (8 metrics endpoints)
- POST /applications/deployments (create deployment)
- GET /applications/deployments/{deployment_id}
- GET /applications/processes/{process_id}
- PUT /applications/processes/{id}
- POST /applications/promote

**Databases (4 endpoints):**
- GET /databases (list)
- POST /databases (create)
- GET /databases/{id}, /databases/{name}
- PUT /databases/{id} (update)
- DELETE /databases/{id}

**Static Sites (5 endpoints):**
- GET /static-sites (list)
- GET /static-sites/{id}
- PUT /static-sites/{id}
- DELETE /static-sites/{id}
- POST /static-sites/deployments
- GET /static-sites/deployments/{deployment_id}
- POST /static-sites/deployments/redeploy

**Pipelines (2 endpoints):**
- GET /pipelines
- POST /pipelines/{id}/create-preview-app

**WordPress Sites (29 endpoints - ⚠️ OVERLAP WITH KINSTA):**
- All /sites/* endpoints also exist in Sevalla spec
- **Recommendation:** Keep WordPress resources ONLY in kinsta provider
- Sevalla users should use kinsta provider for WordPress

**Operations (1 endpoint):**
- GET /operations/{operation_id} (shared polling mechanism)

**Company (1 endpoint):**
- GET /company/{id}/users

### 1.3 Database Resource - Detailed Comparison

#### Request Schema (POST /databases)

**IDENTICAL to MyKinsta:**
```json
{
  "company_id": "uuid",
  "location": "us-central1",
  "resource_type": "db1-db9",
  "display_name": "my-db",
  "db_name": "mydb",
  "db_password": "password",
  "db_user": "user",  // optional for Redis
  "type": "postgresql|redis|mariadb|mysql",
  "version": "15"
}
```

**Required fields:** All except db_user (optional for Redis)

#### Response (200 Synchronous)

**Create Response:**
```json
{
  "database": {
    "id": "uuid"
  }
}
```

**Read Response (GET /databases/{id}):**
```json
{
  "database": {
    "id": "uuid",
    "name": "unique-db-name",
    "display_name": "my-db",
    "status": "ready",
    "created_at": 1668697088806,
    "memory_limit": 250,
    "cpu_limit": 250,
    "storage_size": 1000,
    "type": "postgresql",
    "version": "14",
    "cluster": {
      "id": "uuid",
      "location": "europe-west3",
      "display_name": "Frankfurt, Germany Europe"
    },
    "resource_type_name": "db1",
    "internal_hostname": "name.dns.svc.cluster.local",
    "internal_port": "5432",
    "internal_connections": [{
      "id": "uuid",
      "type": "appResource"
    }],
    "data": {
      "db_name": "mydb",
      "db_password": "password",
      "db_root_password": "root_password",
      "db_user": "username"
    },
    "external_connection_string": "postgresql://...",
    "external_hostname": "db-postgresql.external.kinsta.app",
    "external_port": "31866"
  }
}
```

**Update Request (PUT /databases/{id}):**
```json
{
  "resource_type": "db2",  // optional
  "display_name": "new-name"  // optional
}
```

**Key Observations:**
- ✅ Synchronous operations (no operation_id polling needed)
- ✅ More computed fields than MyKinsta: internal_hostname, connection strings, credentials
- ✅ Password returned in read (stored by API)
- ⚠️ Only resource_type and display_name updatable
- ⚠️ location, type, version, db_name are immutable (must be ForceNew)

### 1.4 Application Resource Schema (NEW)

**Evidence Required:** Need to examine application schemas in detail. Let me check:

