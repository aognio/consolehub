package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"consolehub/internal/api/jsonrpc"
	"consolehub/internal/auth"
	"consolehub/internal/config"
	"consolehub/internal/logger"
	"consolehub/internal/middleware"
	"consolehub/internal/models"
	"consolehub/internal/services"
	"consolehub/internal/storage"
	"consolehub/internal/stream"
	"consolehub/internal/templates"
	"consolehub/internal/ui"
	"consolehub/internal/version"

	"golang.org/x/term"
)

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

func main() {
	var (
		showVersion      bool
		showVersionAlias bool
		configPath       string
		createSuperadmin string
		changePassword   string
		deleteAccount    string
		listAdmins       bool
		monochrome       bool
	)

	flag.BoolVar(&showVersion, "version", false, "Print ConsoleHub server version and exit")
	flag.BoolVar(&showVersionAlias, "v", false, "Print ConsoleHub server version and exit (alias)")
	flag.StringVar(&configPath, "config", "", "Path to server-config.toml configuration file")
	flag.StringVar(&configPath, "config-path", "", "Path to server-config.toml configuration file (alias)")
	flag.StringVar(&createSuperadmin, "create-superadmin", "", "Create a new superadmin account with the specified email")
	flag.StringVar(&changePassword, "change-password", "", "Change password interactively for an existing user email")
	flag.StringVar(&deleteAccount, "delete-account", "", "Delete user account with the specified email")
	flag.BoolVar(&listAdmins, "list-admins", false, "List all superadmin and admin user accounts")
	flag.BoolVar(&monochrome, "monochrome", false, "Disable color and ANSI formatting in CLI output")
	flag.Parse()

	if showVersion || showVersionAlias {
		fmt.Println(version.String())
		os.Exit(0)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	l, err := logger.Init(cfg.Logging)
	if err != nil {
		log.Printf("Warning: failed to initialize logger: %v", err)
	} else {
		defer l.Close()
		l.Info("server", "ConsoleHub server initializing", map[string]any{
			"version":  version.Version,
			"log_file": cfg.Logging.LogFile,
		})
	}

	store, err := storage.New(cfg)
	if err != nil {
		logger.Error("server", "Failed to initialize storage", map[string]any{"error": err.Error()})
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	svc := services.New(store)
	ctx := context.Background()

	// Single Action 1: List Admin Accounts
	if listAdmins {
		users, err := svc.ListUsers(ctx)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}

		if monochrome {
			fmt.Println("EMAIL\t\t\tNAME\t\tROLE\t\tSTATUS")
			for _, u := range users {
				if u.Role == models.RoleSuperAdmin || u.Role == models.RoleAdmin {
					status := "active"
					if !u.Active {
						status = "disabled"
					}
					fmt.Printf("%-24s %-16s %-12s %s\n", u.Email, u.Name, u.Role, status)
				}
			}
		} else {
			fmt.Println("\033[1;34m=== ConsoleHub Administrator Accounts ===\033[0m")
			fmt.Printf("\033[1m%-28s %-20s %-14s %-10s\033[0m\n", "EMAIL", "NAME", "ROLE", "STATUS")
			fmt.Println("-------------------------------------------------------------------------")
			for _, u := range users {
				if u.Role == models.RoleSuperAdmin || u.Role == models.RoleAdmin {
					statusStr := "\033[32mActive\033[0m"
					if !u.Active {
						statusStr = "\033[31mDisabled\033[0m"
					}
					fmt.Printf("%-28s %-20s \033[36m%-14s\033[0m %s\n", u.Email, u.Name, u.Role, statusStr)
				}
			}
		}
		os.Exit(0)
	}

	// Single Action 2: Create Superadmin
	if createSuperadmin != "" {
		existing, err := svc.GetUserByEmail(ctx, createSuperadmin)
		if err == nil && existing != nil {
			log.Fatalf("Error: user with email '%s' already exists", createSuperadmin)
		}

		pwd, err := readPassword(fmt.Sprintf("Enter password for new superadmin (%s): ", createSuperadmin))
		if err != nil || pwd == "" {
			log.Fatalf("Error: invalid password input")
		}

		confirm, err := readPassword("Confirm password: ")
		if err != nil || pwd != confirm {
			log.Fatalf("Error: passwords do not match")
		}

		user, err := svc.CreateUser(ctx, createSuperadmin, pwd, "Super Administrator", models.RoleSuperAdmin)
		if err != nil {
			log.Fatalf("Failed to create superadmin: %v", err)
		}
		if monochrome {
			fmt.Printf("Successfully created superadmin account for %s (ID: %s)\n", user.Email, user.ID)
		} else {
			fmt.Printf("\033[32m✓ Successfully created superadmin account for %s (ID: %s)\033[0m\n", user.Email, user.ID)
		}
		os.Exit(0)
	}

	// Single Action 3: Change Password
	if changePassword != "" {
		user, err := svc.GetUserByEmail(ctx, changePassword)
		if err != nil || user == nil {
			log.Fatalf("Error: user account with email '%s' not found", changePassword)
		}

		pwd, err := readPassword(fmt.Sprintf("Enter new password for %s: ", changePassword))
		if err != nil || pwd == "" {
			log.Fatalf("Error: invalid password input")
		}

		confirm, err := readPassword("Confirm password: ")
		if err != nil || pwd != confirm {
			log.Fatalf("Error: passwords do not match")
		}

		if err := svc.ChangeUserPassword(ctx, user.ID, pwd); err != nil {
			log.Fatalf("Failed to update password: %v", err)
		}
		if monochrome {
			fmt.Printf("Successfully updated password for %s\n", changePassword)
		} else {
			fmt.Printf("\033[32m✓ Successfully updated password for %s\033[0m\n", changePassword)
		}
		os.Exit(0)
	}

	// Single Action 4: Delete Account
	if deleteAccount != "" {
		user, err := svc.GetUserByEmail(ctx, deleteAccount)
		if err != nil || user == nil {
			log.Fatalf("Error: user account with email '%s' not found", deleteAccount)
		}

		if err := svc.DeleteUser(ctx, user.ID); err != nil {
			log.Fatalf("Failed to delete user account: %v", err)
		}
		if monochrome {
			fmt.Printf("Successfully deleted user account %s\n", deleteAccount)
		} else {
			fmt.Printf("\033[32m✓ Successfully deleted user account %s\033[0m\n", deleteAccount)
		}
		os.Exit(0)
	}

	// Default Server Startup Flow
	log.Printf("Starting ConsoleHub Server %s (consolehub-server) on %s:%d (Scheme: %s)...", version.Version, cfg.Server.Host, cfg.Server.Port, cfg.Server.Scheme)
	log.Printf("Public URL: %s", cfg.Server.PublicURL)

	initialPassHash, _ := auth.HashPassword("admin123456")
	admin, err := store.EnsureDefaultSuperAdmin("admin@consolehub.local", initialPassHash)
	if err != nil {
		log.Printf("Warning: default super admin setup: %v", err)
	} else {
		log.Printf("Super admin verified: %s (%s)", admin.Email, admin.Role)
	}

	hub := stream.NewHub()
	jsonrpcHandler := jsonrpc.NewHandler(cfg, svc, hub)
	tmpl, err := templates.New(cfg.Server.Timezone)
	if err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	uiHandler := ui.NewHandler(cfg, svc, tmpl)
	mw := middleware.New(cfg.Security.CookieSecret)

	mux := http.NewServeMux()

	// Serve embedded static assets (/static/)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(templates.StaticFS())))

	// JSON-RPC 2.0 over WebSocket endpoint
	mux.Handle("/api/v1/rpc/ws", jsonrpcHandler)

	// SSE Live stream endpoint for UI Console Viewer
	mux.HandleFunc("/api/v1/runs/live", func(w http.ResponseWriter, r *http.Request) {
		runID := r.URL.Query().Get("run_id")
		if runID == "" {
			http.Error(w, "run_id required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := hub.Subscribe(runID)
		defer hub.Unsubscribe(runID, ch)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case line, open := <-ch:
				if !open {
					return
				}
				fmt.Fprintf(w, "data: {\"sequence\":%d, \"timestamp\":\"%s\", \"stream\":\"%s\", \"kind\":\"%s\", \"text\":\"%s\"}\n\n",
					line.Sequence, line.Timestamp.Format("15:04:05"), line.Stream, line.Kind, line.Text)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// Unauthenticated routes
	mux.HandleFunc("/login", uiHandler.ServeHTTP)

	// Authenticated UI routes
	authenticatedUI := mw.RequireAuth(uiHandler)
	mux.Handle("/logout", authenticatedUI)
	mux.Handle("/switch-tenant", authenticatedUI)
	mux.Handle("/dashboard", authenticatedUI)
	mux.Handle("/tenants", authenticatedUI)
	mux.Handle("/hosts", authenticatedUI)
	mux.Handle("/apps", authenticatedUI)
	mux.Handle("/runs", authenticatedUI)
	mux.Handle("/runs/", authenticatedUI)
	mux.Handle("/search", authenticatedUI)
	mux.Handle("/users", authenticatedUI)
	mux.Handle("/settings", authenticatedUI)

	// Root redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		uiHandler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
