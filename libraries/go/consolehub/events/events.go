package events

import (
	"time"
)

// StreamKind defines stream classification constants.
const (
	StreamStdout = "stdout"
	StreamStderr = "stderr"
	StreamLog    = "log"
	StreamSystem = "system"
)

// KindType defines line payload types.
const (
	KindText     = "text"
	KindJSON     = "json"
	KindProgress = "progress"
	KindPrompt   = "prompt"
)

// Event is the interface implemented by all typed telemetry events.
type Event interface {
	GetSequence() int64
	SetSequence(seq int64)
	GetTimestamp() time.Time
	GetStream() string
	GetKind() string
	ToMap() map[string]any
}

// BaseEvent provides common sequence and timestamp fields.
type BaseEvent struct {
	Sequence  int64     `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"`
	Kind      string    `json:"kind"`
}

func (b *BaseEvent) GetSequence() int64         { return b.Sequence }
func (b *BaseEvent) SetSequence(seq int64)      { b.Sequence = seq }
func (b *BaseEvent) GetTimestamp() time.Time    { return b.Timestamp }
func (b *BaseEvent) GetStream() string          { return b.Stream }
func (b *BaseEvent) GetKind() string            { return b.Kind }

// TextLine represents a raw text line emitted to stdout/stderr.
type TextLine struct {
	BaseEvent
	Text string `json:"text"`
}

func NewTextLine(stream, text string) *TextLine {
	return &TextLine{
		BaseEvent: BaseEvent{
			Timestamp: time.Now(),
			Stream:    stream,
			Kind:      KindText,
		},
		Text: text,
	}
}

func (t *TextLine) ToMap() map[string]any {
	return map[string]any{
		"sequence":  t.Sequence,
		"timestamp": t.Timestamp.Format(time.RFC3339Nano),
		"stream":    t.Stream,
		"kind":      t.Kind,
		"text":      t.Text,
	}
}

// LogEvent represents a structured log record.
type LogEvent struct {
	BaseEvent
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields,omitempty"`
}

func NewLogEvent(level, message string, fields map[string]any) *LogEvent {
	return &LogEvent{
		BaseEvent: BaseEvent{
			Timestamp: time.Now(),
			Stream:    StreamLog,
			Kind:      KindJSON,
		},
		Level:   level,
		Message: message,
		Fields:  fields,
	}
}

func (l *LogEvent) ToMap() map[string]any {
	data := map[string]any{
		"level":   l.Level,
		"message": l.Message,
	}
	if l.Fields != nil {
		data["fields"] = l.Fields
	}

	return map[string]any{
		"sequence":  l.Sequence,
		"timestamp": l.Timestamp.Format(time.RFC3339Nano),
		"stream":    l.Stream,
		"kind":      l.Kind,
		"data":      data,
	}
}

// ProgressEvent represents progress bar tracking updates.
type ProgressEvent struct {
	BaseEvent
	ID       string `json:"id"`
	Label    string `json:"label"`
	Current  int64  `json:"current"`
	Total    int64  `json:"total"`
	Finished bool   `json:"finished"`
}

func NewProgressEvent(id, label string, current, total int64, finished bool) *ProgressEvent {
	return &ProgressEvent{
		BaseEvent: BaseEvent{
			Timestamp: time.Now(),
			Stream:    StreamSystem,
			Kind:      KindProgress,
		},
		ID:       id,
		Label:    label,
		Current:  current,
		Total:    total,
		Finished: finished,
	}
}

func (p *ProgressEvent) ToMap() map[string]any {
	return map[string]any{
		"sequence":  p.Sequence,
		"timestamp": p.Timestamp.Format(time.RFC3339Nano),
		"stream":    p.Stream,
		"kind":      p.Kind,
		"data": map[string]any{
			"id":       p.ID,
			"label":    p.Label,
			"current":  p.Current,
			"total":    p.Total,
			"finished": p.Finished,
		},
	}
}

// PromptEvent represents an interactive prompt emission.
type PromptEvent struct {
	BaseEvent
	ID       string `json:"id"`
	Prompt   string `json:"prompt"`
	Response string `json:"response,omitempty"`
	Secret   bool   `json:"secret"`
}

func NewPromptEvent(id, promptText, response string, secret bool) *PromptEvent {
	return &PromptEvent{
		BaseEvent: BaseEvent{
			Timestamp: time.Now(),
			Stream:    StreamSystem,
			Kind:      KindPrompt,
		},
		ID:       id,
		Prompt:   promptText,
		Response: response,
		Secret:   secret,
	}
}

func (pr *PromptEvent) ToMap() map[string]any {
	respText := pr.Response
	if pr.Secret {
		respText = "********"
	}
	return map[string]any{
		"sequence":  pr.Sequence,
		"timestamp": pr.Timestamp.Format(time.RFC3339Nano),
		"stream":    pr.Stream,
		"kind":      pr.Kind,
		"data": map[string]any{
			"id":       pr.ID,
			"prompt":   pr.Prompt,
			"response": respText,
			"secret":   pr.Secret,
		},
	}
}
