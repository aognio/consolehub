# Security Architecture

## Authentication & Session Management
- Independent cookie-based session management.
- Cookies are signed using `security.cookie_secret` defined in TOML configuration, allowing user sessions to survive server restarts and binary rebuilds.
- Supports `security.secure_cookies` and `security.same_site` flags (`lax`, `strict`, `none`).

## Ingestion Token Security
- Host agents authenticate over WebSockets using Bearer Tokens or `auth.authenticate` procedures.
- Authentication token is validated against assigned tenant credentials before process registration is granted.

## RBAC & Multi-Tenant Scoping
- Every request context carries active tenant isolation scopes.
- Super Admins access cross-tenant administration.
- Admins access assigned tenant resources.
- Users have read-only access strictly within assigned tenant scopes.

## Input Validation & Sanitization
- All incoming JSON-RPC payloads validate schema, line byte limits (max 256 KB per frame), and monotonic sequence numbers.
- HTML rendering auto-escapes stream text content to prevent XSS.
