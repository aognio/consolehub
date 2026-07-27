# ConsoleHub Security Architecture

## Session Security
- Authentication is independent of PocketBase's default admin UI.
- Cookie signatures are created via HMAC-SHA256 using `security.cookie_secret` defined in TOML configuration.
- Existing user login sessions survive binary rebuilds and redeployments.

## Role-Based Access Control (RBAC)
- **Super Admin**: Manage platform settings, manage all tenants, create super admins/admins/users.
- **Admin**: Manage assigned tenants, host processes, app catalogs, and tenant users.
- **User**: Read-only access to view assigned tenants, search logs, and watch live streams.

## Multi-Tenancy Isolation
- All database queries and stream broad-casts are filtered by active tenant scope.
