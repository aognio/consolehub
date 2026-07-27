Build a production-quality Golang application called (temporary name) "ConsoleHub". The goal is to provide a centralized web console for monitoring long-running command-line applications running across multiple hosts. This is NOT a generic observability platform, nor an OpenTelemetry clone. It is a remote console with history.

Technology stack

- Go latest stable
- PocketBase embedded
- HTMX
- Alpine.js
- TailwindCSS
- SQLite (through PocketBase)
- Single self-contained binary
- Configuration loaded with:
  --config /path/to/server-config.toml

Do not hardcode ports, hosts, domains, cookie secrets or URLs.

----------------------------------------------------------------------
Configuration
----------------------------------------------------------------------

Implement a TOML configuration file supporting at least:

- server.host
- server.port
- server.scheme
- server.base_path
- server.public_url

- security.cookie_secret
- security.secure_cookies
- security.same_site
- security.session_duration

- pocketbase.data_dir

- logging.level
- logging.retention_days

The cookie/session secret MUST come from configuration so existing login sessions survive binary rebuilds and redeployments.

----------------------------------------------------------------------
Architecture
----------------------------------------------------------------------

Separate the project into clean packages.

Suggested packages:

internal/config
internal/auth
internal/models
internal/services
internal/api
internal/ui
internal/stream
internal/storage
internal/middleware
internal/templates

Keep HTTP handlers thin.

----------------------------------------------------------------------
Authentication
----------------------------------------------------------------------

Implement authentication independent from PocketBase's default admin UI.

Support:

- Super Admin
- Admin
- User

Permissions:

Super Admin

- Manage everything
- Create/delete tenants
- Create Super Admins
- Create Admins
- Create Users
- View all tenants

Admin

- Manage assigned tenants
- Create Users
- Manage hosts/apps/runs
- View all logs inside their tenants

User

- Read-only
- View assigned tenants
- Search logs
- Watch live streams

Sessions must survive server restarts.

----------------------------------------------------------------------
Multi-tenancy
----------------------------------------------------------------------

Tenants are first-class objects.

Users belong to one or more tenants.

A tenant owns:

- Hosts
- Apps
- Runs
- Streams

Everything must be isolated by tenant.

----------------------------------------------------------------------
PocketBase Collections
----------------------------------------------------------------------

Design collections for at least:

Users
Groups (future)
Tenants
TenantMembers
Hosts
Apps
Runs
StreamLines

Feel free to normalize as appropriate.

----------------------------------------------------------------------
Domain Model
----------------------------------------------------------------------

Tenant

- id
- name
- slug
- active

Host

- id
- tenant_id
- hostname
- display_name
- platform
- last_seen
- online

App

- id
- tenant_id
- name
- display_name
- description

Run

- id (UUID)
- tenant_id
- host_id
- app_id

- pid
- started_at
- finished_at

- status
    running
    exited
    crashed
    stopped

- version
- working_directory
- command_line
- exit_code

StreamLine

- run_id
- timestamp

- stream
    stdout
    stderr
    log

- kind
    text
    json

- text

Do NOT over-engineer an event model.

The unit of storage is simply an ordered stream line.

----------------------------------------------------------------------
Backend API
----------------------------------------------------------------------

Prepare endpoints for future clients.

Examples:

POST /api/v1/runs/register

POST /api/v1/runs/{id}/heartbeat

POST /api/v1/runs/{id}/stream

POST /api/v1/runs/{id}/finish

GET /api/v1/runs/live

Do not implement client libraries yet.

Only prepare the backend.

----------------------------------------------------------------------
Streaming
----------------------------------------------------------------------

Abstract the transport.

Create interfaces so future transports can be added.

Initially prepare:

- HTTP ingestion
- WebSocket ingestion

Both should feed the same backend service.

----------------------------------------------------------------------
UI
----------------------------------------------------------------------

Modern SaaS appearance.

Requirements:

- Responsive
- Desktop
- Tablet
- Mobile

Light mode

Dark mode

Theme stored in browser cookies.

Do NOT require login before choosing theme.

Use HTMX aggressively.

Use Alpine.js only where appropriate.

----------------------------------------------------------------------
Pages
----------------------------------------------------------------------

Login

Dashboard

Tenants

Hosts

Apps

Running Processes

Historical Runs

Live Console

Search

User Administration

Settings

----------------------------------------------------------------------
Dashboard
----------------------------------------------------------------------

Show cards for:

Running processes

Offline hosts

Online hosts

Recent runs

Recent failures

Latest stream activity

----------------------------------------------------------------------
Runs
----------------------------------------------------------------------

Table including:

Status

App

Host

PID

Started

Duration

User

Actions

Clicking opens the console viewer.

----------------------------------------------------------------------
Console Viewer
----------------------------------------------------------------------

The flagship page.

Requirements:

Monospace font

Fast rendering

Auto-scroll

Pause

Resume

Tail mode

Search

Copy line

Jump to timestamp

Download

Support:

Plain text

JSON lines

JSON lines should have collapsible formatting.

Future features can include stack trace grouping and progress rendering.

----------------------------------------------------------------------
Search
----------------------------------------------------------------------

Global search across stream lines.

Support:

Text

Time range

Host

App

Run

Tenant

Stream

----------------------------------------------------------------------
Frontend
----------------------------------------------------------------------

Use reusable layouts and components.

Avoid SPA complexity.

HTMX partial rendering wherever practical.

----------------------------------------------------------------------
Roadmap
----------------------------------------------------------------------

Create the following documentation automatically during implementation:

docs/architecture.md

docs/data-model.md

docs/api.md

docs/security.md

docs/ui.md

docs/roadmap.md

docs/backlog.md

docs/decisions.md

docs/journal.md

----------------------------------------------------------------------
Deliverable
----------------------------------------------------------------------

Before implementing large amounts of code, first produce a comprehensive implementation plan describing:

- architecture
- packages
- PocketBase schema
- UI navigation
- API design
- authentication
- authorization
- streaming architecture
- future client integration
- implementation phases

Save it as:

docs/plan.md

Do not begin major implementation until the plan has been completed and reviewed.
