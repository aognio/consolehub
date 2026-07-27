package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"consolehub/internal/config"
	"consolehub/internal/models"
	"consolehub/internal/services"
	"consolehub/internal/stream"

	"github.com/gorilla/websocket"
)

const (
	ErrCodeParseError      = -32700
	ErrCodeInvalidRequest  = -32600
	ErrCodeMethodNotFound  = -32601
	ErrCodeInvalidParams   = -32602
	ErrCodeInternalError   = -32603
	ErrCodeAuthRequired    = -32001
	ErrCodeTenantNotFound  = -32003
	ErrCodeProcessNotFound = -32004
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Handler struct {
	cfg      *config.Config
	services *services.Services
	hub      *stream.Hub
	upgrader websocket.Upgrader
}

func NewHandler(cfg *config.Config, services *services.Services, hub *stream.Hub) *Handler {
	return &Handler{
		cfg:      cfg,
		services: services,
		hub:      hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type connectionSession struct {
	authenticated bool
	tenantID      string
	tenantSlug    string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	session := &connectionSession{}

	// Check HTTP Bearer token upgrade header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		session.authenticated = true
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req Request
		if err := json.Unmarshal(message, &req); err != nil {
			conn.WriteJSON(Response{
				JSONRPC: "2.0",
				Error:   &Error{Code: ErrCodeParseError, Message: "Parse error"},
			})
			continue
		}

		resp := h.handleRequest(r.Context(), session, &req)
		if resp != nil {
			_ = conn.WriteJSON(resp)
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context, session *connectionSession, req *Request) *Response {
	if req.JSONRPC != "2.0" {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: ErrCodeInvalidRequest, Message: "Invalid JSON-RPC version"},
		}
	}

	if req.Method == "auth.authenticate" {
		var params struct {
			Token string `json:"token"`
		}
		_ = json.Unmarshal(req.Params, &params)
		
		tenantID := ""
		if params.Token != "" {
			if apiKey, err := h.services.ValidateAPIKey(ctx, params.Token); err == nil && apiKey != nil {
				tenantID = apiKey.TenantID
			}
		}

		session.authenticated = true
		session.tenantID = tenantID
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"authenticated": true,
				"tenant_id":     tenantID,
			},
		}
	}

	if !session.authenticated {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: ErrCodeAuthRequired, Message: "Authentication required"},
		}
	}

	switch req.Method {
	case "process.register":
		var params struct {
			Tenant  string `json:"tenant"`
			App     string `json:"app"`
			Host    struct {
				Slug        string `json:"slug"`
				Hostname    string `json:"hostname"`
				FQDN        string `json:"fqdn"`
				DisplayName string `json:"display_name"`
				Platform    string `json:"platform"`
			} `json:"host"`
			Process struct {
				ClientRunID      string `json:"client_run_id"`
				PID              int    `json:"pid"`
				StartedAt        string `json:"started_at"`
				Version          string `json:"version"`
				CommandLine      string `json:"command_line"`
				WorkingDirectory string `json:"working_directory"`
			} `json:"process"`
		}

		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: ErrCodeInvalidParams, Message: "Invalid params"},
			}
		}

		tenant, err := h.services.GetTenantBySlug(ctx, params.Tenant)
		if err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: ErrCodeTenantNotFound, Message: "Tenant not found"},
			}
		}

		slug := params.Host.Slug
		if slug == "" {
			slug = params.Host.Hostname
		}
		host, _ := h.services.RegisterHost(ctx, slug, params.Host.Hostname, params.Host.FQDN, params.Host.DisplayName, params.Host.Platform)
		app, _ := h.services.CreateApp(ctx, tenant.ID, params.App, params.App, "")

		startedAt, _ := time.Parse(time.RFC3339, params.Process.StartedAt)
		if startedAt.IsZero() {
			startedAt = time.Now()
		}

		run, err := h.services.RegisterProcessRun(ctx, services.RegisterRunParams{
			TenantID:         tenant.ID,
			HostID:           host.ID,
			AppID:            app.ID,
			ClientRunID:      params.Process.ClientRunID,
			PID:              params.Process.PID,
			StartedAt:        startedAt,
			Version:          params.Process.Version,
			WorkingDirectory: params.Process.WorkingDirectory,
			CommandLine:      params.Process.CommandLine,
		})
		if err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: ErrCodeInternalError, Message: err.Error()},
			}
		}

		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"process_id":                 run.ID,
				"accepted_client_run_id":     params.Process.ClientRunID,
				"heartbeat_interval_seconds": 30,
				"maximum_batch_lines":        250,
				"maximum_batch_bytes":        262144,
			},
		}

	case "stream.append":
		var params struct {
			ProcessID     string `json:"process_id"`
			BatchID       string `json:"batch_id"`
			FirstSequence int64  `json:"first_sequence"`
			Lines         []struct {
				Sequence  int64  `json:"sequence"`
				Timestamp string `json:"timestamp"`
				Stream    string `json:"stream"`
				Kind      string `json:"kind"`
				Text      string `json:"text"`
			} `json:"lines"`
		}

		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: ErrCodeInvalidParams, Message: "Invalid params"},
			}
		}

		streamLines := make([]models.StreamLine, len(params.Lines))
		for i, l := range params.Lines {
			ts, _ := time.Parse(time.RFC3339, l.Timestamp)
			if ts.IsZero() {
				ts = time.Now()
			}
			line := models.StreamLine{
				RunID:     params.ProcessID,
				Sequence:  l.Sequence,
				Timestamp: ts,
				Stream:    l.Stream,
				Kind:      l.Kind,
				Text:      l.Text,
			}
			streamLines[i] = line
			h.hub.Publish(line)
		}

		acceptedThrough, dup, err := h.services.AppendStreamLines(nil, params.ProcessID, params.BatchID, params.FirstSequence, streamLines)
		if err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: ErrCodeProcessNotFound, Message: "Process not found"},
			}
		}

		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"batch_id":                  params.BatchID,
				"accepted_through_sequence": acceptedThrough,
				"duplicate":                 dup,
			},
		}

	case "connection.ping", "process.heartbeat":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"status": "ok",
			},
		}

	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: ErrCodeMethodNotFound, Message: fmt.Sprintf("Method '%s' not found", req.Method)},
		}
	}
}
