You are extending terraform-provider-sevalla with read-only data sources.

Context
- Sevalla spec has GET/PUT/DELETE /applications but no POST /applications.
- We will ship data sources for visibility and integration now.

Goal
- Implement stable data sources that do not require create endpoints.

Tasks
1. Implement data sources:
   - sevalla_applications
   - sevalla_databases
   - sevalla_static_sites
   - sevalla_pipelines
2. Support common filtering and pagination (limit/offset) as exposed by the API.
3. Add docs + examples for each data source.
4. Add unit tests for pagination and schema mapping.

Output
- Data source implementations + docs/examples + tests
