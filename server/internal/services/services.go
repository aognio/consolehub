package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"consolehub/internal/apikey"
	"consolehub/internal/auth"
	"consolehub/internal/models"
	"consolehub/internal/storage"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrHostNotFound   = errors.New("host not found")
	ErrAppNotFound    = errors.New("app not found")
	ErrInvalidCreds   = errors.New("invalid email or password")
	ErrRunNotFound    = errors.New("run process not found")
	ErrSequenceGap    = errors.New("sequence gap detected")
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyExpired  = errors.New("api key expired")
	ErrAPIKeyInactive = errors.New("api key inactive")
)

type RegisterRunParams struct {
	TenantID         string
	HostID           string
	AppID            string
	ClientRunID      string
	PID              int
	StartedAt        time.Time
	Version          string
	CommandLine      string
	WorkingDirectory string
}

type Services struct {
	store       *storage.Storage
	batchMu     sync.Mutex
	seenBatches map[string]int64
}

func New(store *storage.Storage) *Services {
	return &Services{
		store:       store,
		seenBatches: make(map[string]int64),
	}
}

// -------------------------------------------------------------------
// Tenant Management
// -------------------------------------------------------------------

func (s *Services) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	col, err := s.store.App.FindCollectionByNameOrId("ch_tenants")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("slug", slug)
	rec.Set("active", true)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return &models.Tenant{
		ID:        rec.Id,
		Name:      name,
		Slug:      slug,
		Active:    true,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetTenantByID(ctx context.Context, id string) (*models.Tenant, error) {
	rec, err := s.store.App.FindRecordById("ch_tenants", id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return &models.Tenant{
		ID:        rec.Id,
		Name:      rec.GetString("name"),
		Slug:      rec.GetString("slug"),
		Active:    rec.GetBool("active"),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_tenants", "slug = {:slug}", "", 1, 0, dbx.Params{"slug": slug})
	if err != nil || len(recs) == 0 {
		return nil, ErrTenantNotFound
	}
	rec := recs[0]
	return &models.Tenant{
		ID:        rec.Id,
		Name:      rec.GetString("name"),
		Slug:      rec.GetString("slug"),
		Active:    rec.GetBool("active"),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) ListTenants(ctx context.Context) ([]*models.Tenant, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_tenants", "", "+name", 500, 0)
	if err != nil {
		return nil, err
	}
	tenants := make([]*models.Tenant, 0, len(recs))
	for _, r := range recs {
		tenants = append(tenants, &models.Tenant{
			ID:        r.Id,
			Name:      r.GetString("name"),
			Slug:      r.GetString("slug"),
			Active:    r.GetBool("active"),
			CreatedAt: r.GetDateTime("created").Time(),
			UpdatedAt: r.GetDateTime("updated").Time(),
		})
	}
	return tenants, nil
}

func (s *Services) UpdateTenant(ctx context.Context, id, name, slug string, active bool) (*models.Tenant, error) {
	rec, err := s.store.App.FindRecordById("ch_tenants", id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	rec.Set("name", name)
	rec.Set("slug", slug)
	rec.Set("active", active)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("update tenant: %w", err)
	}

	return &models.Tenant{
		ID:        rec.Id,
		Name:      name,
		Slug:      slug,
		Active:    active,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) DeleteTenant(ctx context.Context, id string) error {
	rec, err := s.store.App.FindRecordById("ch_tenants", id)
	if err != nil {
		return ErrTenantNotFound
	}
	return s.store.App.Delete(rec)
}

// -------------------------------------------------------------------
// Host Ingestion & Registration (Global Hosts)
// -------------------------------------------------------------------

func (s *Services) RegisterHost(ctx context.Context, slug, hostname, fqdn, displayName, platform string) (*models.Host, error) {
	existing, err := s.store.GetHostBySlug(slug)
	if err == nil {
		_ = s.store.UpdateHost(existing.ID, slug, hostname, fqdn, displayName, platform, true)
		return s.store.GetHostByID(existing.ID)
	}
	return s.store.CreateHost(slug, hostname, fqdn, displayName, platform)
}

func (s *Services) GetHostByID(ctx context.Context, id string) (*models.Host, error) {
	return s.store.GetHostByID(id)
}

func (s *Services) GetHostBySlug(ctx context.Context, slug string) (*models.Host, error) {
	return s.store.GetHostBySlug(slug)
}

func (s *Services) ListHosts(ctx context.Context) ([]*models.Host, error) {
	hosts, err := s.store.ListHosts()
	if err != nil {
		return nil, err
	}
	result := make([]*models.Host, len(hosts))
	for i := range hosts {
		result[i] = &hosts[i]
	}
	return result, nil
}

func (s *Services) UpdateHost(ctx context.Context, id, slug, hostname, fqdn, displayName, platform string, online bool) (*models.Host, error) {
	if err := s.store.UpdateHost(id, slug, hostname, fqdn, displayName, platform, online); err != nil {
		return nil, err
	}
	return s.store.GetHostByID(id)
}

func (s *Services) ListHostsByTenant(ctx context.Context, tenantID string) ([]*models.Host, error) {
	hosts, err := s.store.ListHostsByTenant(tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Host, len(hosts))
	for i := range hosts {
		result[i] = &hosts[i]
	}
	return result, nil
}

func (s *Services) ListTenantsByHost(ctx context.Context, hostID string) ([]*models.Tenant, error) {
	tenants, err := s.store.ListTenantsByHost(hostID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Tenant, len(tenants))
	for i := range tenants {
		result[i] = &tenants[i]
	}
	return result, nil
}

func (s *Services) DeleteHost(ctx context.Context, id string) error {
	return s.store.DeleteHost(id)
}

func (s *Services) AssociateHostTenant(ctx context.Context, hostID, tenantID string) error {
	return s.store.AssociateHostTenant(hostID, tenantID)
}

func (s *Services) DissociateHostTenant(ctx context.Context, hostID, tenantID string) error {
	return s.store.DissociateHostTenant(hostID, tenantID)
}

func (s *Services) SyncHostTenants(ctx context.Context, hostID string, tenantIDs []string) error {
	return s.store.SyncHostTenants(hostID, tenantIDs)
}

// -------------------------------------------------------------------
// Application Management
// -------------------------------------------------------------------

func (s *Services) CreateApp(ctx context.Context, tenantID, name, displayName, description string) (*models.App, error) {
	if name == "" {
		name = "default-app"
	}
	if displayName == "" {
		displayName = name
	}

	existing, err := s.GetAppByName(ctx, tenantID, name)
	if err == nil && existing != nil {
		return existing, nil
	}

	col, err := s.store.App.FindCollectionByNameOrId("ch_apps")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("tenant_id", tenantID)
	rec.Set("name", name)
	rec.Set("display_name", displayName)
	rec.Set("description", description)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("create app: %w", err)
	}

	return &models.App{
		ID:          rec.Id,
		TenantID:    tenantID,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetAppByName(ctx context.Context, tenantID, name string) (*models.App, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_apps", "tenant_id = {:tenant_id} && name = {:name}", "", 1, 0,
		dbx.Params{"tenant_id": tenantID, "name": name})
	if err != nil || len(recs) == 0 {
		return nil, ErrAppNotFound
	}
	r := recs[0]
	return &models.App{
		ID:          r.Id,
		TenantID:    r.GetString("tenant_id"),
		Name:        r.GetString("name"),
		DisplayName: r.GetString("display_name"),
		Description: r.GetString("description"),
		CreatedAt:   r.GetDateTime("created").Time(),
		UpdatedAt:   r.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetAppByID(ctx context.Context, id string) (*models.App, error) {
	rec, err := s.store.App.FindRecordById("ch_apps", id)
	if err != nil {
		return nil, ErrAppNotFound
	}
	return &models.App{
		ID:          rec.Id,
		TenantID:    rec.GetString("tenant_id"),
		Name:        rec.GetString("name"),
		DisplayName: rec.GetString("display_name"),
		Description: rec.GetString("description"),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) ListApps(ctx context.Context) ([]*models.App, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_apps", "", "+name", 500, 0)
	if err != nil {
		return nil, err
	}
	apps := make([]*models.App, 0, len(recs))
	for _, r := range recs {
		apps = append(apps, &models.App{
			ID:          r.Id,
			TenantID:    r.GetString("tenant_id"),
			Name:        r.GetString("name"),
			DisplayName: r.GetString("display_name"),
			Description: r.GetString("description"),
			CreatedAt:   r.GetDateTime("created").Time(),
			UpdatedAt:   r.GetDateTime("updated").Time(),
		})
	}
	return apps, nil
}

func (s *Services) ListAppsByTenant(ctx context.Context, tenantID string) ([]*models.App, error) {
	if tenantID == "" {
		return s.ListApps(ctx)
	}
	recs, err := s.store.App.FindRecordsByFilter("ch_apps", "tenant_id = {:tenant_id}", "+name", 500, 0, dbx.Params{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	apps := make([]*models.App, 0, len(recs))
	for _, r := range recs {
		apps = append(apps, &models.App{
			ID:          r.Id,
			TenantID:    r.GetString("tenant_id"),
			Name:        r.GetString("name"),
			DisplayName: r.GetString("display_name"),
			Description: r.GetString("description"),
			CreatedAt:   r.GetDateTime("created").Time(),
			UpdatedAt:   r.GetDateTime("updated").Time(),
		})
	}
	return apps, nil
}

func (s *Services) UpdateApp(ctx context.Context, id, name, displayName, description string) (*models.App, error) {
	rec, err := s.store.App.FindRecordById("ch_apps", id)
	if err != nil {
		return nil, ErrAppNotFound
	}

	rec.Set("name", name)
	rec.Set("display_name", displayName)
	rec.Set("description", description)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("update app: %w", err)
	}

	return &models.App{
		ID:          rec.Id,
		TenantID:    rec.GetString("tenant_id"),
		Name:        name,
		DisplayName: displayName,
		Description: description,
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) DeleteApp(ctx context.Context, id string) error {
	rec, err := s.store.App.FindRecordById("ch_apps", id)
	if err != nil {
		return ErrAppNotFound
	}
	return s.store.App.Delete(rec)
}

// -------------------------------------------------------------------
// Process Run Registration & Lifecycle
// -------------------------------------------------------------------

func (s *Services) RegisterProcessRun(ctx context.Context, p RegisterRunParams) (*models.Run, error) {
	if p.ClientRunID == "" {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		p.ClientRunID = fmt.Sprintf("run-%d-%x", time.Now().UnixNano(), b)
	} else {
		existing, err := s.GetRunByClientRunID(ctx, p.TenantID, p.ClientRunID)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	col, err := s.store.App.FindCollectionByNameOrId("ch_runs")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("client_run_id", p.ClientRunID)
	rec.Set("tenant_id", p.TenantID)
	rec.Set("host_id", p.HostID)
	rec.Set("app_id", p.AppID)
	rec.Set("pid", p.PID)
	rec.Set("started_at", p.StartedAt)
	rec.Set("status", models.RunStatusRunning)
	rec.Set("version", p.Version)
	rec.Set("command_line", p.CommandLine)
	rec.Set("working_directory", p.WorkingDirectory)
	rec.Set("last_sequence", 0)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("register process run: %w", err)
	}

	return &models.Run{
		ID:               rec.Id,
		ClientRunID:      p.ClientRunID,
		TenantID:         p.TenantID,
		HostID:           p.HostID,
		AppID:            p.AppID,
		PID:              p.PID,
		StartedAt:        p.StartedAt,
		Status:           models.RunStatusRunning,
		Version:          p.Version,
		CommandLine:      p.CommandLine,
		WorkingDirectory: p.WorkingDirectory,
		LastSequence:     0,
		CreatedAt:        rec.GetDateTime("created").Time(),
		UpdatedAt:        rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetRunByID(ctx context.Context, id string) (*models.Run, error) {
	r, err := s.store.App.FindRecordById("ch_runs", id)
	if err != nil {
		return nil, ErrRunNotFound
	}
	var finished *time.Time
	if !r.GetDateTime("finished_at").IsZero() {
		t := r.GetDateTime("finished_at").Time()
		finished = &t
	}

	return &models.Run{
		ID:               r.Id,
		ClientRunID:      r.GetString("client_run_id"),
		TenantID:         r.GetString("tenant_id"),
		HostID:           r.GetString("host_id"),
		AppID:            r.GetString("app_id"),
		PID:              r.GetInt("pid"),
		StartedAt:        r.GetDateTime("started_at").Time(),
		FinishedAt:       finished,
		Status:           r.GetString("status"),
		Version:          r.GetString("version"),
		CommandLine:      r.GetString("command_line"),
		WorkingDirectory: r.GetString("working_directory"),
		ExitCode:         r.GetInt("exit_code"),
		LastSequence:     int64(r.GetInt("last_sequence")),
		CreatedAt:        r.GetDateTime("created").Time(),
		UpdatedAt:        r.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetRunByClientRunID(ctx context.Context, tenantID, clientRunID string) (*models.Run, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_runs", "tenant_id = {:tenant_id} && client_run_id = {:client_run_id}", "", 1, 0,
		dbx.Params{"tenant_id": tenantID, "client_run_id": clientRunID})
	if err != nil || len(recs) == 0 {
		return nil, ErrRunNotFound
	}
	return s.GetRunByID(ctx, recs[0].Id)
}

func (s *Services) ListRuns(ctx context.Context) ([]*models.Run, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_runs", "", "-started_at", 500, 0)
	if err != nil {
		return nil, err
	}
	runs := make([]*models.Run, 0, len(recs))
	for _, r := range recs {
		run, err := s.GetRunByID(ctx, r.Id)
		if err == nil {
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func (s *Services) ListRunsByTenant(ctx context.Context, tenantID string) ([]*models.Run, error) {
	if tenantID == "" {
		return s.ListRuns(ctx)
	}
	recs, err := s.store.App.FindRecordsByFilter("ch_runs", "tenant_id = {:tenant_id}", "-started_at", 500, 0, dbx.Params{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	runs := make([]*models.Run, 0, len(recs))
	for _, r := range recs {
		run, err := s.GetRunByID(ctx, r.Id)
		if err == nil {
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func (s *Services) UpdateRunStatus(ctx context.Context, id, status string, exitCode int) (*models.Run, error) {
	r, err := s.store.App.FindRecordById("ch_runs", id)
	if err != nil {
		return nil, ErrRunNotFound
	}

	r.Set("status", status)
	r.Set("exit_code", exitCode)
	if status == models.RunStatusExited || status == models.RunStatusCrashed || status == models.RunStatusStopped {
		r.Set("finished_at", time.Now())
	}

	if err := s.store.App.Save(r); err != nil {
		return nil, fmt.Errorf("update run status: %w", err)
	}

	return s.GetRunByID(ctx, id)
}

func (s *Services) DeleteRun(ctx context.Context, id string) error {
	r, err := s.store.App.FindRecordById("ch_runs", id)
	if err != nil {
		return ErrRunNotFound
	}
	return s.store.App.Delete(r)
}

func (s *Services) FinishProcessRun(ctx context.Context, runID string, finishedAt time.Time, status string, exitCode int, lastSeq int64) error {
	rec, err := s.store.App.FindRecordById("ch_runs", runID)
	if err != nil {
		return ErrRunNotFound
	}

	rec.Set("finished_at", finishedAt)
	rec.Set("status", status)
	rec.Set("exit_code", exitCode)
	if lastSeq > 0 {
		rec.Set("last_sequence", lastSeq)
	}

	return s.store.App.Save(rec)
}

// -------------------------------------------------------------------
// User Administration Management
// -------------------------------------------------------------------

func (s *Services) CreateUser(ctx context.Context, email, password, name, role string) (*models.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	col, errCol := s.store.App.FindCollectionByNameOrId("ch_users")
	if errCol != nil {
		return nil, errCol
	}

	rec := core.NewRecord(col)
	rec.Set("email", email)
	rec.Set("password_hash", hash)
	rec.Set("name", name)
	rec.Set("role", role)
	rec.Set("active", true)

	if err := s.store.App.Save(rec); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &models.User{
		ID:        rec.Id,
		Email:     email,
		Name:      name,
		Role:      role,
		Active:    true,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) ListUsers(ctx context.Context) ([]*models.User, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_users", "", "+email", 500, 0)
	if err != nil {
		return nil, err
	}
	users := make([]*models.User, 0, len(recs))
	for _, r := range recs {
		users = append(users, &models.User{
			ID:        r.Id,
			Email:     r.GetString("email"),
			Name:      r.GetString("name"),
			Role:      r.GetString("role"),
			Active:    r.GetBool("active"),
			CreatedAt: r.GetDateTime("created").Time(),
			UpdatedAt: r.GetDateTime("updated").Time(),
		})
	}
	return users, nil
}

func (s *Services) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	r, err := s.store.App.FindRecordById("ch_users", id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &models.User{
		ID:        r.Id,
		Email:     r.GetString("email"),
		Name:      r.GetString("name"),
		Role:      r.GetString("role"),
		Active:    r.GetBool("active"),
		CreatedAt: r.GetDateTime("created").Time(),
		UpdatedAt: r.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_users", "email = {:email}", "", 1, 0, dbx.Params{"email": email})
	if err != nil || len(recs) == 0 {
		return nil, ErrUserNotFound
	}
	r := recs[0]
	return &models.User{
		ID:        r.Id,
		Email:     r.GetString("email"),
		Name:      r.GetString("name"),
		Role:      r.GetString("role"),
		Active:    r.GetBool("active"),
		CreatedAt: r.GetDateTime("created").Time(),
		UpdatedAt: r.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) ChangeUserPassword(ctx context.Context, id, newPassword string) error {
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	r, err := s.store.App.FindRecordById("ch_users", id)
	if err != nil {
		return ErrUserNotFound
	}

	r.Set("password_hash", hash)
	return s.store.App.Save(r)
}

func (s *Services) UpdateUser(ctx context.Context, id, email, name, role string, active bool) (*models.User, error) {
	r, err := s.store.App.FindRecordById("ch_users", id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	r.Set("email", email)
	r.Set("name", name)
	r.Set("role", role)
	r.Set("active", active)

	if err := s.store.App.Save(r); err != nil {
		return nil, err
	}

	return &models.User{
		ID:        r.Id,
		Email:     email,
		Name:      name,
		Role:      role,
		Active:    active,
		CreatedAt: r.GetDateTime("created").Time(),
		UpdatedAt: r.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) AddUserToTenant(ctx context.Context, tenantID, userID, role string) (*models.TenantMember, error) {
	col, err := s.store.App.FindCollectionByNameOrId("ch_tenant_members")
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(col)
	rec.Set("tenant_id", tenantID)
	rec.Set("user_id", userID)
	rec.Set("role", role)
	if err := s.store.App.Save(rec); err != nil {
		return nil, err
	}
	return &models.TenantMember{
		ID:        rec.Id,
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}, nil
}

func (s *Services) RemoveUserFromTenant(ctx context.Context, tenantID, userID string) error {
	recs, err := s.store.App.FindRecordsByFilter("ch_tenant_members", "tenant_id = {:tenant_id} && user_id = {:user_id}", "", 1, 0,
		dbx.Params{"tenant_id": tenantID, "user_id": userID})
	if err != nil || len(recs) == 0 {
		return nil
	}
	return s.store.App.Delete(recs[0])
}

func (s *Services) DeleteUser(ctx context.Context, id string) error {
	r, err := s.store.App.FindRecordById("ch_users", id)
	if err != nil {
		return ErrUserNotFound
	}
	return s.store.App.Delete(r)
}

func (s *Services) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	recs, err := s.store.App.FindRecordsByFilter("ch_users", "email = {:email}", "", 1, 0, dbx.Params{"email": email})
	if err != nil || len(recs) == 0 {
		return nil, ErrInvalidCreds
	}
	r := recs[0]
	if !r.GetBool("active") {
		return nil, ErrInvalidCreds
	}
	hash := r.GetString("password_hash")
	if !auth.CheckPasswordHash(password, hash) {
		return nil, ErrInvalidCreds
	}

	return &models.User{
		ID:        r.Id,
		Email:     r.GetString("email"),
		Name:      r.GetString("name"),
		Role:      r.GetString("role"),
		Active:    r.GetBool("active"),
		CreatedAt: r.GetDateTime("created").Time(),
		UpdatedAt: r.GetDateTime("updated").Time(),
	}, nil
}

// -------------------------------------------------------------------
// API Keys Management
// -------------------------------------------------------------------

func (s *Services) CreateAPIKey(ctx context.Context, tenantID, title, description string, expiresAt *time.Time) (*models.APIKey, string, error) {
	if tenantID == "" {
		return nil, "", errors.New("tenant_id is required")
	}

	rawKey := apikey.MustGenerate()
	keyHash, err := auth.HashArgon2(rawKey)
	if err != nil {
		return nil, "", fmt.Errorf("hash key: %w", err)
	}

	keyPrefix := formatDisplayPrefix(rawKey)

	apiKey, err := s.store.CreateAPIKey(tenantID, keyHash, keyPrefix, title, description, expiresAt)
	if err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *Services) ListAPIKeys(ctx context.Context, tenantID string) ([]*models.APIKey, error) {
	keys, err := s.store.ListAPIKeysByTenant(tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.APIKey, len(keys))
	for i := range keys {
		result[i] = &keys[i]
	}
	return result, nil
}

func (s *Services) UpdateAPIKey(ctx context.Context, id, title, description string, active bool, expiresAt *time.Time) error {
	return s.store.UpdateAPIKey(id, title, description, active, expiresAt)
}

func (s *Services) DeleteAPIKey(ctx context.Context, id string) error {
	return s.store.DeleteAPIKey(id)
}

func (s *Services) ValidateAPIKey(ctx context.Context, rawKey string) (*models.APIKey, error) {
	if rawKey == "" {
		return nil, ErrAPIKeyNotFound
	}

	// Verify key format & checksum before database search
	if !apikey.Verify(rawKey) {
		return nil, ErrAPIKeyNotFound
	}

	// Query API keys from storage
	keys, err := s.store.ListAPIKeysByTenant("")
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	for _, k := range keys {
		if !k.Active {
			continue
		}
		if match, _ := auth.VerifyArgon2(rawKey, k.KeyHash); match {
			if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
				return nil, ErrAPIKeyExpired
			}
			return &k, nil
		}
	}

	return nil, ErrAPIKeyNotFound
}

func formatDisplayPrefix(key string) string {
	if len(key) <= 14 {
		return key
	}
	return key[:8] + "..." + key[len(key)-6:]
}

// -------------------------------------------------------------------
// Stream Lines Append
// -------------------------------------------------------------------

func (s *Services) AppendStreamLines(ctx context.Context, runID, batchID string, firstSeq int64, lines []models.StreamLine) (int64, bool, error) {
	if batchID != "" {
		s.batchMu.Lock()
		if prevSeq, exists := s.seenBatches[batchID]; exists {
			s.batchMu.Unlock()
			return prevSeq, true, nil
		}
		s.batchMu.Unlock()
	}

	rec, err := s.store.App.FindRecordById("ch_runs", runID)
	if err != nil {
		return 0, false, ErrRunNotFound
	}

	tenantID := rec.GetString("tenant_id")
	lastSeq := int64(rec.GetInt("last_sequence"))

	col, errCol := s.store.App.FindCollectionByNameOrId("ch_stream_lines")
	if errCol != nil {
		return 0, false, errCol
	}

	currentSeq := lastSeq
	for _, l := range lines {
		currentSeq = l.Sequence
		r := core.NewRecord(col)
		r.Set("run_id", runID)
		r.Set("tenant_id", tenantID)
		r.Set("sequence", l.Sequence)
		r.Set("timestamp", l.Timestamp)
		r.Set("stream", l.Stream)
		r.Set("kind", l.Kind)
		r.Set("text", l.Text)

		_ = s.store.App.Save(r)
	}

	rec.Set("last_sequence", currentSeq)
	_ = s.store.App.Save(rec)

	if batchID != "" {
		s.batchMu.Lock()
		s.seenBatches[batchID] = currentSeq
		s.batchMu.Unlock()
	}

	return currentSeq, false, nil
}
