package ui

import (
	"net/http"
	"strconv"
	"time"

	"consolehub/internal/auth"
	"consolehub/internal/config"
	"consolehub/internal/models"
	"consolehub/internal/services"
	"consolehub/internal/templates"
)

type Handler struct {
	cfg      *config.Config
	services *services.Services
	tmpl     *templates.TemplateEngine
}

func NewHandler(cfg *config.Config, services *services.Services, tmpl *templates.TemplateEngine) *Handler {
	return &Handler{
		cfg:      cfg,
		services: services,
		tmpl:     tmpl,
	}
}

func (h *Handler) preparePageData(r *http.Request, title string, extra map[string]any) map[string]any {
	if extra == nil {
		extra = make(map[string]any)
	}

	extra["Title"] = title

	tenants, err := h.services.ListTenants(r.Context())
	if err == nil {
		extra["Tenants"] = tenants
		cookie, errCookie := r.Cookie("consolehub_tenant")
		if errCookie == nil && cookie.Value != "" {
			for _, t := range tenants {
				if t.ID == cookie.Value {
					extra["ActiveTenant"] = t
					break
				}
			}
		}
	}

	return extra
}

func (h *Handler) getActiveTenantID(r *http.Request) string {
	cookie, err := r.Cookie("consolehub_tenant")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/login":
		if r.Method == http.MethodPost {
			h.handleLoginPost(w, r)
			return
		}
		_ = h.tmpl.Render(w, "login.html", map[string]any{"Title": "Login"})

	case "/logout":
		http.SetCookie(w, &http.Cookie{
			Name:    "consolehub_session",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)

	case "/switch-tenant":
		if r.Method == http.MethodPost {
			tenantID := r.FormValue("tenant_id")
			http.SetCookie(w, &http.Cookie{
				Name:    "consolehub_tenant",
				Value:   tenantID,
				Path:    "/",
				Expires: time.Now().Add(30 * 24 * time.Hour),
			})
		}
		ref := r.Referer()
		if ref == "" {
			ref = "/dashboard"
		}
		http.Redirect(w, r, ref, http.StatusSeeOther)

	case "/dashboard":
		hosts, _ := h.services.ListHosts(r.Context())
		runs, _ := h.services.ListRuns(r.Context())

		onlineHosts := 0
		offlineHosts := 0
		for _, host := range hosts {
			if host.Online {
				onlineHosts++
			} else {
				offlineHosts++
			}
		}

		runningCount := 0
		recentFailures := 0
		for _, run := range runs {
			if run.Status == models.RunStatusRunning {
				runningCount++
			} else if run.Status == models.RunStatusCrashed || run.ExitCode != 0 {
				recentFailures++
			}
		}

		_ = h.tmpl.Render(w, "dashboard.html", h.preparePageData(r, "Dashboard", map[string]any{
			"RunningCount":   runningCount,
			"OnlineHosts":    onlineHosts,
			"OfflineHosts":   offlineHosts,
			"RecentRuns":     len(runs),
			"RecentFailures": recentFailures,
		}))

	case "/tenants":
		if r.Method == http.MethodPost {
			h.handleTenantPost(w, r)
			return
		}
		_ = h.tmpl.Render(w, "tenants.html", h.preparePageData(r, "Tenants", nil))

	case "/hosts":
		if r.Method == http.MethodPost {
			h.handleHostPost(w, r)
			return
		}
		hosts, _ := h.services.ListHosts(r.Context())
		tenants, _ := h.services.ListTenants(r.Context())
		_ = h.tmpl.Render(w, "hosts.html", h.preparePageData(r, "Hosts", map[string]any{
			"Hosts":   hosts,
			"Tenants": tenants,
		}))

	case "/apps":
		if r.Method == http.MethodPost {
			h.handleAppPost(w, r)
			return
		}
		apps, _ := h.services.ListApps(r.Context())
		_ = h.tmpl.Render(w, "apps.html", h.preparePageData(r, "Apps", map[string]any{
			"Apps": apps,
		}))

	case "/runs":
		if r.Method == http.MethodPost {
			h.handleRunPost(w, r)
			return
		}
		runs, _ := h.services.ListRuns(r.Context())
		_ = h.tmpl.Render(w, "runs.html", h.preparePageData(r, "Runs", map[string]any{
			"Runs": runs,
		}))

	case "/search":
		_ = h.tmpl.Render(w, "search.html", h.preparePageData(r, "Search", nil))

	case "/users":
		if r.Method == http.MethodPost {
			h.handleUserAdminPost(w, r)
			return
		}
		users, _ := h.services.ListUsers(r.Context())
		_ = h.tmpl.Render(w, "users.html", h.preparePageData(r, "User Admin", map[string]any{
			"Users": users,
		}))

	case "/api-keys":
		if r.Method == http.MethodPost {
			h.handleAPIKeyPost(w, r)
			return
		}
		activeTenantID := h.getActiveTenantID(r)
		apiKeys, _ := h.services.ListAPIKeys(r.Context(), activeTenantID)
		tenants, _ := h.services.ListTenants(r.Context())
		_ = h.tmpl.Render(w, "api_keys.html", h.preparePageData(r, "API Keys", map[string]any{
			"APIKeys": apiKeys,
			"Tenants": tenants,
		}))

	case "/settings":
		_ = h.tmpl.Render(w, "settings.html", h.preparePageData(r, "Settings", map[string]any{
			"Config": h.cfg,
		}))

	default:
		if len(r.URL.Path) > 6 && r.URL.Path[:6] == "/runs/" {
			h.renderConsole(w, r)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (h *Handler) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.services.AuthenticateUser(r.Context(), email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = h.tmpl.Render(w, "login.html", map[string]any{
			"Title": "Login",
			"Error": "Invalid email address or password.",
		})
		return
	}

	token, err := auth.CreateSessionToken(user, h.cfg.Security.CookieSecret, h.cfg.Security.Duration())
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "consolehub_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Security.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(h.cfg.Security.Duration()),
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) handleTenantPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "create":
		name := r.FormValue("name")
		slug := r.FormValue("slug")
		if name != "" && slug != "" {
			_, _ = h.services.CreateTenant(r.Context(), name, slug)
		}
	case "update":
		id := r.FormValue("tenant_id")
		name := r.FormValue("name")
		slug := r.FormValue("slug")
		active := r.FormValue("active") == "true"
		if id != "" {
			_, _ = h.services.UpdateTenant(r.Context(), id, name, slug, active)
		}
	case "delete":
		id := r.FormValue("tenant_id")
		if id != "" {
			_ = h.services.DeleteTenant(r.Context(), id)
		}
	}
	http.Redirect(w, r, "/tenants", http.StatusSeeOther)
}

func (h *Handler) handleHostPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "create":
		slug := r.FormValue("slug")
		hostname := r.FormValue("hostname")
		fqdn := r.FormValue("fqdn")
		displayName := r.FormValue("display_name")
		platform := r.FormValue("platform")
		if slug != "" && hostname != "" {
			_, _ = h.services.RegisterHost(r.Context(), slug, hostname, fqdn, displayName, platform)
		}
	case "update":
		id := r.FormValue("host_id")
		slug := r.FormValue("slug")
		hostname := r.FormValue("hostname")
		fqdn := r.FormValue("fqdn")
		displayName := r.FormValue("display_name")
		platform := r.FormValue("platform")
		online := r.FormValue("online") == "true"
		if id != "" {
			_, _ = h.services.UpdateHost(r.Context(), id, slug, hostname, fqdn, displayName, platform, online)
		}
	case "delete":
		id := r.FormValue("host_id")
		if id != "" {
			_ = h.services.DeleteHost(r.Context(), id)
		}
	case "associate_tenant":
		hostID := r.FormValue("host_id")
		tenantID := r.FormValue("tenant_id")
		if hostID != "" && tenantID != "" {
			_ = h.services.AssociateHostTenant(r.Context(), hostID, tenantID)
		}
	case "dissociate_tenant":
		hostID := r.FormValue("host_id")
		tenantID := r.FormValue("tenant_id")
		if hostID != "" && tenantID != "" {
			_ = h.services.DissociateHostTenant(r.Context(), hostID, tenantID)
		}
	}
	http.Redirect(w, r, "/hosts", http.StatusSeeOther)
}

func (h *Handler) handleAppPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "create":
		tenantID := r.FormValue("tenant_id")
		name := r.FormValue("name")
		displayName := r.FormValue("display_name")
		description := r.FormValue("description")
		if tenantID != "" && name != "" {
			_, _ = h.services.CreateApp(r.Context(), tenantID, name, displayName, description)
		}
	case "update":
		id := r.FormValue("app_id")
		name := r.FormValue("name")
		displayName := r.FormValue("display_name")
		description := r.FormValue("description")
		if id != "" {
			_, _ = h.services.UpdateApp(r.Context(), id, name, displayName, description)
		}
	case "delete":
		id := r.FormValue("app_id")
		if id != "" {
			_ = h.services.DeleteApp(r.Context(), id)
		}
	}
	http.Redirect(w, r, "/apps", http.StatusSeeOther)
}

func (h *Handler) handleRunPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "update_status":
		id := r.FormValue("run_id")
		status := r.FormValue("status")
		exitCode, _ := strconv.Atoi(r.FormValue("exit_code"))
		if id != "" && status != "" {
			_, _ = h.services.UpdateRunStatus(r.Context(), id, status, exitCode)
		}
	case "delete":
		id := r.FormValue("run_id")
		if id != "" {
			_ = h.services.DeleteRun(r.Context(), id)
		}
	}
	http.Redirect(w, r, "/runs", http.StatusSeeOther)
}

func (h *Handler) handleUserAdminPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")

	switch action {
	case "create":
		email := r.FormValue("email")
		password := r.FormValue("password")
		name := r.FormValue("name")
		role := r.FormValue("role")
		if role == "" {
			role = models.RoleUser
		}
		user, err := h.services.CreateUser(r.Context(), email, password, name, role)
		if err == nil && user != nil {
			tenantID := r.FormValue("tenant_id")
			if tenantID != "" {
				_, _ = h.services.AddUserToTenant(r.Context(), tenantID, user.ID, models.RoleUser)
			}
		}

	case "update":
		userID := r.FormValue("user_id")
		email := r.FormValue("email")
		name := r.FormValue("name")
		role := r.FormValue("role")
		active := r.FormValue("active") == "true"
		if userID != "" {
			_, _ = h.services.UpdateUser(r.Context(), userID, email, name, role, active)
		}

	case "delete":
		userID := r.FormValue("user_id")
		if userID != "" {
			_ = h.services.DeleteUser(r.Context(), userID)
		}

	case "associate_tenant":
		userID := r.FormValue("user_id")
		tenantID := r.FormValue("tenant_id")
		tenantRole := r.FormValue("tenant_role")
		if tenantRole == "" {
			tenantRole = models.RoleUser
		}
		if userID != "" && tenantID != "" {
			_, _ = h.services.AddUserToTenant(r.Context(), tenantID, userID, tenantRole)
		}

	case "remove_tenant":
		userID := r.FormValue("user_id")
		tenantID := r.FormValue("tenant_id")
		if userID != "" && tenantID != "" {
			_ = h.services.RemoveUserFromTenant(r.Context(), tenantID, userID)
		}

	case "reset_password":
		userID := r.FormValue("user_id")
		newPassword := r.FormValue("new_password")
		if userID != "" && newPassword != "" {
			_ = h.services.ChangeUserPassword(r.Context(), userID, newPassword)
		}

	case "toggle_status":
		userID := r.FormValue("user_id")
		user, err := h.services.GetUserByID(r.Context(), userID)
		if err == nil {
			_, _ = h.services.UpdateUser(r.Context(), user.ID, user.Email, user.Name, user.Role, !user.Active)
		}
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *Handler) handleAPIKeyPost(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	switch action {
	case "create":
		tenantID := r.FormValue("tenant_id")
		if tenantID == "" {
			tenantID = h.getActiveTenantID(r)
		}
		title := r.FormValue("title")
		desc := r.FormValue("description")
		expStr := r.FormValue("expires_at")

		var expiresAt *time.Time
		if expStr != "" {
			if t, err := time.Parse("2006-01-02", expStr); err == nil {
				expiresAt = &t
			} else if t, err := time.Parse(time.RFC3339, expStr); err == nil {
				expiresAt = &t
			}
		}

		if tenantID != "" {
			_, rawKey, err := h.services.CreateAPIKey(r.Context(), tenantID, title, desc, expiresAt)
			if err == nil && rawKey != "" {
				activeTenantID := h.getActiveTenantID(r)
				apiKeys, _ := h.services.ListAPIKeys(r.Context(), activeTenantID)
				tenants, _ := h.services.ListTenants(r.Context())
				_ = h.tmpl.Render(w, "api_keys.html", h.preparePageData(r, "API Keys", map[string]any{
					"APIKeys":       apiKeys,
					"Tenants":       tenants,
					"CreatedRawKey": rawKey,
				}))
				return
			}
		}

	case "update":
		id := r.FormValue("api_key_id")
		title := r.FormValue("title")
		desc := r.FormValue("description")
		active := r.FormValue("active") == "true"
		expStr := r.FormValue("expires_at")

		var expiresAt *time.Time
		if expStr != "" {
			if t, err := time.Parse("2006-01-02", expStr); err == nil {
				expiresAt = &t
			} else if t, err := time.Parse(time.RFC3339, expStr); err == nil {
				expiresAt = &t
			}
		}

		if id != "" {
			_ = h.services.UpdateAPIKey(r.Context(), id, title, desc, active, expiresAt)
		}

	case "delete":
		id := r.FormValue("api_key_id")
		if id != "" {
			_ = h.services.DeleteAPIKey(r.Context(), id)
		}
	}
	http.Redirect(w, r, "/api-keys", http.StatusSeeOther)
}

func (h *Handler) renderConsole(w http.ResponseWriter, r *http.Request) {
	_ = h.tmpl.Render(w, "console.html", h.preparePageData(r, "Console Viewer", map[string]any{
		"Run": &models.Run{
			ID:  "run-100",
			PID: 42817,
		},
	}))
}
