package models

import "time"

// Role constants for user authorization.
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// Run status constants.
const (
	RunStatusRunning = "running"
	RunStatusExited  = "exited"
	RunStatusCrashed = "crashed"
	RunStatusStopped = "stopped"
	RunStatusUnknown = "unknown"
)

// Stream type constants.
const (
	StreamStdout = "stdout"
	StreamStderr = "stderr"
	StreamLog    = "log"
)

// Kind constants.
const (
	KindText = "text"
	KindJSON = "json"
)

// Tenant represents an isolated multi-tenant workspace.
type Tenant struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created"`
	UpdatedAt time.Time `json:"updated_at" db:"updated"`
}

// User represents a system account.
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Role         string    `json:"role" db:"role"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"created_at" db:"created"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated"`
}

// TenantMember maps users to tenants with specific permissions.
type TenantMember struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created"`
	UpdatedAt time.Time `json:"updated_at" db:"updated"`
}

// Host represents a registered remote machine (global, can be associated to multiple tenants).
type Host struct {
	ID          string    `json:"id" db:"id"`
	Slug        string    `json:"slug" db:"slug"`
	FQDN        string    `json:"fqdn" db:"fqdn"`
	Hostname    string    `json:"hostname" db:"hostname"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Platform    string    `json:"platform" db:"platform"`
	LastSeen    time.Time `json:"last_seen" db:"last_seen"`
	Online      bool      `json:"online" db:"online"`
	CreatedAt   time.Time `json:"created_at" db:"created"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated"`
}

// HostTenant represents the many-to-many association between hosts and tenants.
type HostTenant struct {
	ID       string `json:"id" db:"id"`
	HostID   string `json:"host_id" db:"host_id"`
	TenantID string `json:"tenant_id" db:"tenant_id"`
}

// App represents an application being monitored.
type App struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated"`
}

// Run represents a single process execution.
type Run struct {
	ID               string     `json:"id" db:"id"`
	ClientRunID      string     `json:"client_run_id" db:"client_run_id"`
	TenantID         string     `json:"tenant_id" db:"tenant_id"`
	HostID           string     `json:"host_id" db:"host_id"`
	AppID            string     `json:"app_id" db:"app_id"`
	PID              int        `json:"pid" db:"pid"`
	StartedAt        time.Time  `json:"started_at" db:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty" db:"finished_at"`
	Status           string     `json:"status" db:"status"`
	Version          string     `json:"version" db:"version"`
	WorkingDirectory string     `json:"working_directory" db:"working_directory"`
	CommandLine      string     `json:"command_line" db:"command_line"`
	ExitCode         int        `json:"exit_code" db:"exit_code"`
	LastSequence     int64      `json:"last_sequence" db:"last_sequence"`
	CreatedAt        time.Time  `json:"created_at" db:"created"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated"`
}

// StreamLine represents an individual log or stdout/stderr line from a process execution.
type StreamLine struct {
	ID        string    `json:"id" db:"id"`
	RunID     string    `json:"run_id" db:"run_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Sequence  int64     `json:"sequence" db:"sequence"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Stream    string    `json:"stream" db:"stream"`
	Kind      string    `json:"kind" db:"kind"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created"`
}

// APIKey represents an authentication API key associated to a tenant.
type APIKey struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	KeyHash     string     `json:"-" db:"key_hash"`            // Argon2id hash of raw key
	KeyPrefix   string     `json:"key_prefix" db:"key_prefix"` // e.g. "sk_3q2Z...2AB9XQ" display prefix
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Active      bool       `json:"active" db:"active"`
	CreatedAt   time.Time  `json:"created_at" db:"created"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated"`
}
