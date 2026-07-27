package protocol

// Method constants
const (
	MethodAuthAuthenticate = "auth.authenticate"
	MethodProcessRegister  = "process.register"
	MethodStreamAppend     = "stream.append"
	MethodProcessHeartbeat = "process.heartbeat"
	MethodStreamResume     = "stream.resume"
	MethodProcessFinish    = "process.finish"
)

// Standard JSON-RPC 2.0 Error Codes
const (
	ErrParseError           = -32700
	ErrInvalidRequest       = -32600
	ErrMethodNotFound       = -32601
	ErrInvalidParams        = -32602
	ErrInternalError        = -32603
	ErrAuthRequired         = -32001
	ErrPermissionDenied     = -32002
	ErrTenantNotFound       = -32003
	ErrProcessNotFound      = -32004
	ErrProcessFinished      = -32005
	ErrInvalidSequence      = -32006
	ErrSequenceGap          = -32007
	ErrBatchTooLarge        = -32008
	ErrRateLimited          = -32009
	ErrIncompatibleProtocol = -32010
)

// RequestFrame represents a standard JSON-RPC 2.0 request frame.
type RequestFrame struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// ResponseFrame represents a standard JSON-RPC 2.0 response frame.
type ResponseFrame struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents JSON-RPC 2.0 error object.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// AuthParams represents auth.authenticate request parameters.
type AuthParams struct {
	Token    string            `json:"token"`
	Protocol map[string]string `json:"protocol,omitempty"`
}

// AuthResult represents auth.authenticate response payload.
type AuthResult struct {
	Authenticated bool   `json:"authenticated"`
	TenantID      string `json:"tenant_id,omitempty"`
	TenantSlug    string `json:"tenant_slug,omitempty"`
}

// ProcessRegisterParams represents process.register request parameters.
type ProcessRegisterParams struct {
	Tenant  string         `json:"tenant"`
	App     string         `json:"app"`
	Host    HostParams     `json:"host"`
	Process ProcessDetails `json:"process"`
}

type HostParams struct {
	Hostname    string `json:"hostname"`
	DisplayName string `json:"display_name,omitempty"`
	Platform    string `json:"platform,omitempty"`
}

type ProcessDetails struct {
	ClientRunID      string `json:"client_run_id"`
	PID              int    `json:"pid"`
	StartedAt        string `json:"started_at"`
	Version          string `json:"version"`
	CommandLine      string `json:"command_line"`
	WorkingDirectory string `json:"working_directory"`
}

// ProcessRegisterResult represents process.register response payload.
type ProcessRegisterResult struct {
	ProcessID        string `json:"process_id"`
	AcceptedRunID    string `json:"accepted_client_run_id"`
	HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
}

// StreamAppendParams represents stream.append request parameters.
type StreamAppendParams struct {
	ProcessID     string           `json:"process_id"`
	BatchID       string           `json:"batch_id"`
	FirstSequence int64            `json:"first_sequence"`
	Lines         []map[string]any `json:"lines"`
}

// StreamAppendResult represents stream.append response payload.
type StreamAppendResult struct {
	BatchID         string `json:"batch_id"`
	AcceptedThrough int64  `json:"accepted_through_sequence"`
	Duplicate       bool   `json:"duplicate"`
}

// StreamResumeParams represents stream.resume request parameters.
type StreamResumeParams struct {
	ProcessID   string `json:"process_id"`
	ClientRunID string `json:"client_run_id"`
}

// StreamResumeResult represents stream.resume response payload.
type StreamResumeResult struct {
	AcceptedThrough int64 `json:"accepted_through_sequence"`
}

// ProcessFinishParams represents process.finish request parameters.
type ProcessFinishParams struct {
	ProcessID    string `json:"process_id"`
	FinishedAt   string `json:"finished_at"`
	Status       string `json:"status"`
	ExitCode     int    `json:"exit_code"`
	LastSequence int64  `json:"last_sequence"`
}
