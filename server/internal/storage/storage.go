package storage

import (
	"fmt"
	"time"

	"consolehub/internal/config"
	"consolehub/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Storage struct {
	App *pocketbase.PocketBase
}

func New(cfg *config.Config) (*Storage, error) {
	dataDir := "./pb_data"
	if cfg != nil && cfg.PocketBase.DataDir != "" {
		dataDir = cfg.PocketBase.DataDir
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: dataDir,
	})

	if err := app.Bootstrap(); err != nil {
		return nil, fmt.Errorf("bootstrap pocketbase: %w", err)
	}

	s := &Storage{App: app}
	if err := s.InitSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

// InitSchema ensures custom collections and indexes exist.
func (s *Storage) InitSchema() error {
	collections := []string{
		"ch_users",
		"ch_tenants",
		"ch_tenant_members",
		"ch_hosts",
		"ch_host_tenants",
		"ch_apps",
		"ch_runs",
		"ch_stream_lines",
		"ch_groups",
		"ch_api_keys",
	}

	for _, colName := range collections {
		existing, _ := s.App.FindCollectionByNameOrId(colName)
		if existing != nil {
			continue
		}

		col := core.NewCollection(colName, colName)
		col.Type = core.CollectionTypeBase

		switch colName {
		case "ch_users":
			col.Fields.Add(
				&core.TextField{Name: "email", Required: true},
				&core.TextField{Name: "password_hash", Required: true},
				&core.TextField{Name: "name", Required: true},
				&core.TextField{Name: "role", Required: true},
				&core.BoolField{Name: "active"},
			)
			col.AddIndex("idx_ch_users_email", true, "email", "")

		case "ch_tenants":
			col.Fields.Add(
				&core.TextField{Name: "name", Required: true},
				&core.TextField{Name: "slug", Required: true},
				&core.BoolField{Name: "active"},
			)
			col.AddIndex("idx_ch_tenants_slug", true, "slug", "")

		case "ch_tenant_members":
			col.Fields.Add(
				&core.TextField{Name: "tenant_id", Required: true},
				&core.TextField{Name: "user_id", Required: true},
				&core.TextField{Name: "role", Required: true},
			)
			col.AddIndex("idx_ch_tenant_members_unique", true, "tenant_id, user_id", "")

		case "ch_hosts":
			col.Fields.Add(
				&core.TextField{Name: "slug", Required: true},
				&core.TextField{Name: "fqdn"},
				&core.TextField{Name: "hostname", Required: true},
				&core.TextField{Name: "display_name"},
				&core.TextField{Name: "platform"},
				&core.DateField{Name: "last_seen"},
				&core.BoolField{Name: "online"},
			)
			col.AddIndex("idx_ch_hosts_slug", true, "slug", "")

		case "ch_host_tenants":
			col.Fields.Add(
				&core.TextField{Name: "host_id", Required: true},
				&core.TextField{Name: "tenant_id", Required: true},
			)
			col.AddIndex("idx_ch_host_tenants_unique", true, "host_id, tenant_id", "")
			col.AddIndex("idx_ch_host_tenants_tenant", false, "tenant_id", "")

		case "ch_apps":
			col.Fields.Add(
				&core.TextField{Name: "tenant_id", Required: true},
				&core.TextField{Name: "name", Required: true},
				&core.TextField{Name: "display_name"},
				&core.TextField{Name: "description"},
			)
			col.AddIndex("idx_ch_apps_tenant", false, "tenant_id", "")

		case "ch_runs":
			col.Fields.Add(
				&core.TextField{Name: "client_run_id", Required: true},
				&core.TextField{Name: "tenant_id", Required: true},
				&core.TextField{Name: "host_id", Required: true},
				&core.TextField{Name: "app_id", Required: true},
				&core.NumberField{Name: "pid"},
				&core.DateField{Name: "started_at"},
				&core.DateField{Name: "finished_at"},
				&core.TextField{Name: "status", Required: true},
				&core.TextField{Name: "version"},
				&core.TextField{Name: "working_directory"},
				&core.TextField{Name: "command_line"},
				&core.NumberField{Name: "exit_code"},
				&core.NumberField{Name: "last_sequence"},
			)
			col.AddIndex("idx_ch_runs_client_run_id", true, "tenant_id, client_run_id", "")

		case "ch_stream_lines":
			col.Fields.Add(
				&core.TextField{Name: "run_id", Required: true},
				&core.TextField{Name: "tenant_id", Required: true},
				&core.NumberField{Name: "sequence", Required: true},
				&core.DateField{Name: "timestamp", Required: true},
				&core.TextField{Name: "stream", Required: true},
				&core.TextField{Name: "kind", Required: true},
				&core.TextField{Name: "text"},
			)
			col.AddIndex("idx_ch_stream_lines_run_seq", true, "run_id, sequence", "")

		case "ch_groups":
			col.Fields.Add(
				&core.TextField{Name: "name", Required: true},
			)

		case "ch_api_keys":
			col.Fields.Add(
				&core.TextField{Name: "tenant_id", Required: true},
				&core.TextField{Name: "key_hash", Required: true},
				&core.TextField{Name: "key_prefix", Required: true},
				&core.TextField{Name: "title"},
				&core.TextField{Name: "description"},
				&core.DateField{Name: "expires_at"},
				&core.BoolField{Name: "active"},
			)
			col.AddIndex("idx_ch_api_keys_tenant", false, "tenant_id", "")
		}

		if err := s.App.Save(col); err != nil {
			return fmt.Errorf("save collection %s: %w", colName, err)
		}
	}

	return nil
}

// CreateOrUpdateSuperAdmin creates a super admin user or updates password if user exists.
func (s *Storage) CreateOrUpdateSuperAdmin(email, passwordHash string) (*models.User, error) {
	recs, err := s.App.FindRecordsByFilter("ch_users", "email = {:email} || role = 'super_admin'", "", 1, 0, dbx.Params{"email": email})
	col, errCol := s.App.FindCollectionByNameOrId("ch_users")
	if errCol != nil {
		return nil, errCol
	}

	var rec *core.Record
	if err == nil && len(recs) > 0 {
		rec = recs[0]
		rec.Set("email", email)
		rec.Set("password_hash", passwordHash)
		rec.Set("active", true)
	} else {
		rec = core.NewRecord(col)
		rec.Set("email", email)
		rec.Set("password_hash", passwordHash)
		rec.Set("name", "Super Administrator")
		rec.Set("role", models.RoleSuperAdmin)
		rec.Set("active", true)
	}

	if err := s.App.Save(rec); err != nil {
		return nil, fmt.Errorf("save super admin: %w", err)
	}

	return &models.User{
		ID:           rec.Id,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         rec.GetString("name"),
		Role:         models.RoleSuperAdmin,
		Active:       true,
	}, nil
}

// EnsureDefaultSuperAdmin creates a super admin user if no users exist.
func (s *Storage) EnsureDefaultSuperAdmin(email, passwordHash string) (*models.User, error) {
	users, err := s.App.FindRecordsByFilter("ch_users", "role = 'super_admin'", "", 1, 0)
	if err == nil && len(users) > 0 {
		rec := users[0]
		return &models.User{
			ID:           rec.Id,
			Email:        rec.GetString("email"),
			PasswordHash: rec.GetString("password_hash"),
			Name:         rec.GetString("name"),
			Role:         rec.GetString("role"),
			Active:       rec.GetBool("active"),
		}, nil
	}

	return s.CreateOrUpdateSuperAdmin(email, passwordHash)
}

// ListAllAdmins returns all users with super_admin or admin roles.
func (s *Storage) ListAllAdmins() ([]models.User, error) {
	recs, err := s.App.FindRecordsByFilter("ch_users", "role = 'super_admin' || role = 'admin'", "", 100, 0)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, len(recs))
	for _, rec := range recs {
		users = append(users, models.User{
			ID:     rec.Id,
			Email:  rec.GetString("email"),
			Name:   rec.GetString("name"),
			Role:   rec.GetString("role"),
			Active: rec.GetBool("active"),
		})
	}
	return users, nil
}

// CreateHost creates a new global host.
func (s *Storage) CreateHost(slug, hostname, fqdn, displayName, platform string) (*models.Host, error) {
	col, err := s.App.FindCollectionByNameOrId("ch_hosts")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("slug", slug)
	rec.Set("hostname", hostname)
	rec.Set("fqdn", fqdn)
	rec.Set("display_name", displayName)
	rec.Set("platform", platform)
	rec.Set("online", false)

	if err := s.App.Save(rec); err != nil {
		return nil, fmt.Errorf("create host: %w", err)
	}

	return &models.Host{
		ID:          rec.Id,
		Slug:        slug,
		FQDN:        fqdn,
		Hostname:    hostname,
		DisplayName: displayName,
		Platform:    platform,
		Online:      false,
	}, nil
}

// GetHostByID retrieves a host by its ID.
func (s *Storage) GetHostByID(id string) (*models.Host, error) {
	rec, err := s.App.FindRecordById("ch_hosts", id)
	if err != nil {
		return nil, err
	}
	return recordToHost(rec), nil
}

// GetHostBySlug retrieves a host by its slug.
func (s *Storage) GetHostBySlug(slug string) (*models.Host, error) {
	recs, err := s.App.FindRecordsByFilter("ch_hosts", "slug = {:slug}", "", 1, 0, dbx.Params{"slug": slug})
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("host not found")
	}
	return recordToHost(recs[0]), nil
}

// ListHosts returns all global hosts.
func (s *Storage) ListHosts() ([]models.Host, error) {
	recs, err := s.App.FindRecordsByFilter("ch_hosts", "", "slug", 500, 0)
	if err != nil {
		return nil, err
	}
	hosts := make([]models.Host, 0, len(recs))
	for _, rec := range recs {
		hosts = append(hosts, *recordToHost(rec))
	}
	return hosts, nil
}

// ListHostsByTenant returns all hosts associated with a specific tenant.
func (s *Storage) ListHostsByTenant(tenantID string) ([]models.Host, error) {
	assocRecs, err := s.App.FindRecordsByFilter("ch_host_tenants", "tenant_id = {:tenant_id}", "", 100, 0, dbx.Params{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}

	hosts := make([]models.Host, 0, len(assocRecs))
	for _, assoc := range assocRecs {
		hostID := assoc.GetString("host_id")
		host, err := s.GetHostByID(hostID)
		if err == nil {
			hosts = append(hosts, *host)
		}
	}
	return hosts, nil
}

// UpdateHost updates an existing host.
func (s *Storage) UpdateHost(id, slug, hostname, fqdn, displayName, platform string, online bool) error {
	rec, err := s.App.FindRecordById("ch_hosts", id)
	if err != nil {
		return err
	}
	rec.Set("slug", slug)
	rec.Set("hostname", hostname)
	rec.Set("fqdn", fqdn)
	rec.Set("display_name", displayName)
	rec.Set("platform", platform)
	rec.Set("online", online)
	return s.App.Save(rec)
}

// DeleteHost deletes a host and its tenant associations.
func (s *Storage) DeleteHost(id string) error {
	assocs, _ := s.App.FindRecordsByFilter("ch_host_tenants", "host_id = {:host_id}", "", 100, 0, dbx.Params{"host_id": id})
	for _, assoc := range assocs {
		s.App.Delete(assoc)
	}
	rec, err := s.App.FindRecordById("ch_hosts", id)
	if err != nil {
		return err
	}
	return s.App.Delete(rec)
}

// AssociateHostTenant associates a host with a tenant.
func (s *Storage) AssociateHostTenant(hostID, tenantID string) error {
	col, err := s.App.FindCollectionByNameOrId("ch_host_tenants")
	if err != nil {
		return err
	}

	existing, _ := s.App.FindRecordsByFilter("ch_host_tenants", "host_id = {:host_id} && tenant_id = {:tenant_id}", "", 1, 0,
		dbx.Params{"host_id": hostID, "tenant_id": tenantID})
	if len(existing) > 0 {
		return nil
	}

	rec := core.NewRecord(col)
	rec.Set("host_id", hostID)
	rec.Set("tenant_id", tenantID)
	return s.App.Save(rec)
}

// DissociateHostTenant removes a host from a tenant.
func (s *Storage) DissociateHostTenant(hostID, tenantID string) error {
	assocs, err := s.App.FindRecordsByFilter("ch_host_tenants", "host_id = {:host_id} && tenant_id = {:tenant_id}", "", 1, 0,
		dbx.Params{"host_id": hostID, "tenant_id": tenantID})
	if err != nil {
		return err
	}
	for _, assoc := range assocs {
		s.App.Delete(assoc)
	}
	return nil
}

// GetHostTenantIDs returns all tenant IDs associated with a host.
func (s *Storage) GetHostTenantIDs(hostID string) ([]string, error) {
	assocs, err := s.App.FindRecordsByFilter("ch_host_tenants", "host_id = {:host_id}", "", 100, 0, dbx.Params{"host_id": hostID})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(assocs))
	for _, a := range assocs {
		ids = append(ids, a.GetString("tenant_id"))
	}
	return ids, nil
}

// API Key Storage Methods

// CreateAPIKey creates a new API key record with an Argon2id hash and prefix.
func (s *Storage) CreateAPIKey(tenantID, keyHash, keyPrefix, title, description string, expiresAt *time.Time) (*models.APIKey, error) {
	col, err := s.App.FindCollectionByNameOrId("ch_api_keys")
	if err != nil {
		return nil, err
	}

	rec := core.NewRecord(col)
	rec.Set("tenant_id", tenantID)
	rec.Set("key_hash", keyHash)
	rec.Set("key_prefix", keyPrefix)
	rec.Set("title", title)
	rec.Set("description", description)
	if expiresAt != nil {
		rec.Set("expires_at", expiresAt.Format("2006-01-02 15:04:05.000Z"))
	}
	rec.Set("active", true)

	if err := s.App.Save(rec); err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return recordToAPIKey(rec), nil
}

// ListAPIKeysByTenant returns all API keys for a tenant.
func (s *Storage) ListAPIKeysByTenant(tenantID string) ([]models.APIKey, error) {
	var recs []*core.Record
	var err error
	if tenantID != "" {
		recs, err = s.App.FindRecordsByFilter("ch_api_keys", "tenant_id = {:tenant_id}", "-created", 500, 0, dbx.Params{"tenant_id": tenantID})
	} else {
		recs, err = s.App.FindAllRecords("ch_api_keys")
	}
	if err != nil {
		return nil, err
	}

	keys := make([]models.APIKey, 0, len(recs))
	for _, rec := range recs {
		keys = append(keys, *recordToAPIKey(rec))
	}
	return keys, nil
}

// UpdateAPIKey updates title, description, active status, and optional expiration of an API key.
func (s *Storage) UpdateAPIKey(id, title, description string, active bool, expiresAt *time.Time) error {
	rec, err := s.App.FindRecordById("ch_api_keys", id)
	if err != nil {
		return err
	}

	rec.Set("title", title)
	rec.Set("description", description)
	rec.Set("active", active)
	if expiresAt != nil {
		rec.Set("expires_at", expiresAt.Format("2006-01-02 15:04:05.000Z"))
	} else {
		rec.Set("expires_at", nil)
	}

	return s.App.Save(rec)
}

// DeleteAPIKey deletes an API key record.
func (s *Storage) DeleteAPIKey(id string) error {
	rec, err := s.App.FindRecordById("ch_api_keys", id)
	if err != nil {
		return err
	}
	return s.App.Delete(rec)
}

func recordToHost(rec *core.Record) *models.Host {
	return &models.Host{
		ID:          rec.Id,
		Slug:        rec.GetString("slug"),
		FQDN:        rec.GetString("fqdn"),
		Hostname:    rec.GetString("hostname"),
		DisplayName: rec.GetString("display_name"),
		Platform:    rec.GetString("platform"),
		LastSeen:    rec.GetDateTime("last_seen").Time(),
		Online:      rec.GetBool("online"),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}

func recordToAPIKey(rec *core.Record) *models.APIKey {
	var expPtr *time.Time
	if !rec.GetDateTime("expires_at").IsZero() {
		t := rec.GetDateTime("expires_at").Time()
		expPtr = &t
	}

	return &models.APIKey{
		ID:          rec.Id,
		TenantID:    rec.GetString("tenant_id"),
		KeyHash:     rec.GetString("key_hash"),
		KeyPrefix:   rec.GetString("key_prefix"),
		Title:       rec.GetString("title"),
		Description: rec.GetString("description"),
		ExpiresAt:   expPtr,
		Active:      rec.GetBool("active"),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}
