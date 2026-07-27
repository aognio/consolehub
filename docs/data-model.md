# ConsoleHub Data Model

ConsoleHub uses PocketBase collections backed by SQLite for data persistence.

## PocketBase Collections

### 1. `ch_tenants`
- `id` (text, primary key)
- `name` (text)
- `slug` (text, unique)
- `active` (bool)
- `created`, `updated`

### 2. `ch_users`
- `id` (text, primary key)
- `email` (text, unique)
- `password_hash` (text)
- `name` (text)
- `role` (text: `super_admin`, `admin`, `user`)
- `active` (bool)
- `created`, `updated`

### 3. `ch_tenant_members`
- `id` (text, primary key)
- `tenant_id` (text, relation)
- `user_id` (text, relation)
- `role` (text)
- `created`, `updated`

### 4. `ch_hosts`
- `id` (text, primary key)
- `tenant_id` (text, relation)
- `hostname` (text)
- `display_name` (text)
- `platform` (text)
- `last_seen` (date/time)
- `online` (bool)
- `created`, `updated`

### 5. `ch_apps`
- `id` (text, primary key)
- `tenant_id` (text, relation)
- `name` (text)
- `display_name` (text)
- `description` (text)
- `created`, `updated`

### 6. `ch_runs`
- `id` (text, primary key)
- `client_run_id` (text, client UUID for idempotency)
- `tenant_id` (text, relation)
- `host_id` (text, relation)
- `app_id` (text, relation)
- `pid` (number)
- `started_at` (date/time)
- `finished_at` (date/time)
- `status` (text: `running`, `exited`, `crashed`, `stopped`, `unknown`)
- `version` (text)
- `working_directory` (text)
- `command_line` (text)
- `exit_code` (number)
- `last_sequence` (number)
- `created`, `updated`

### 7. `ch_stream_lines`
- `id` (text, primary key)
- `run_id` (text, relation)
- `tenant_id` (text, relation)
- `sequence` (number, monotonic sequence)
- `timestamp` (date/time)
- `stream` (text: `stdout`, `stderr`, `log`)
- `kind` (text: `text`, `json`)
- `text` (text)
- `created`
