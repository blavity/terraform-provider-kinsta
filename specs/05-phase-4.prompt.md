You are implementing action-style resources (imperative triggers) for Sevalla.

Context
- POST exists for deployments endpoints.
- Terraform should model these as separate resources (or terraform_data triggers) rather than pretending they’re long-lived infrastructure.

Goal
- Implement deployment trigger resources with clear semantics.

Tasks
1. Implement:
   - sevalla_application_deployment (POST /applications/deployments + GET status)
   - sevalla_static_site_deployment (POST /static-sites/deployments + GET status)
   - sevalla_static_site_redeploy (POST /static-sites/deployments/redeploy) if it exists in spec
2. Document lifecycle semantics clearly:
   - create triggers a deployment
   - read reports latest status
   - delete is a no-op or removes from state (document)
3. Add acceptance tests that do not leak resources.

Output
- Action resources + docs + tests
