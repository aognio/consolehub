package consolehub

import (
	"io"
	"sync"

	"github.com/aognio/consolehub/libraries/go/consolehub/events"
	"github.com/aognio/consolehub/libraries/go/consolehub/progress"
)

var (
	defaultClient *Client
	defaultOnce   sync.Once
)

// Default returns or initializes the thread-safe global default client.
func Default() *Client {
	defaultOnce.Do(func() {
		client, err := New()
		if err != nil {
			client, _ = New(WithDisabled(true))
		}
		defaultClient = client
	})
	return defaultClient
}

// Close closes default client if initialized.
func Close() error {
	if defaultClient != nil {
		return defaultClient.Close()
	}
	return nil
}

// Global Standard Helpers
func Print(v ...any)                 { Default().Print(v...) }
func Printf(format string, v ...any) { Default().Printf(format, v...) }
func Println(v ...any)               { Default().Println(v...) }

func Fprint(w io.Writer, v ...any)                 { Default().Fprint(w, v...) }
func Fprintf(w io.Writer, format string, v ...any) { Default().Fprintf(w, format, v...) }
func Fprintln(w io.Writer, v ...any)               { Default().Fprintln(w, v...) }

func Stdout() io.Writer                      { return Default().Stdout() }
func Stderr() io.Writer                      { return Default().Stderr() }
func Writer(streamName ...string) io.Writer { return Default().Writer(streamName...) }

func Debug(msg string) { Default().Debug(msg) }
func Info(msg string)  { Default().Info(msg) }
func Warn(msg string)  { Default().Warn(msg) }
func Error(msg string) { Default().Error(msg) }

func Debugf(format string, v ...any) { Default().Debugf(format, v...) }
func Infof(format string, v ...any)  { Default().Infof(format, v...) }
func Warnf(format string, v ...any)  { Default().Warnf(format, v...) }
func Errorf(format string, v ...any) { Default().Errorf(format, v...) }

func Log(evt events.Event)    { Default().Log(evt) }
func Report(evt events.Event) { Default().Report(evt) }

func Progress(label string, total int64) *progress.Tracker {
	return Default().Progress(label, total)
}

func UploadProgress(label string, totalBytes int64) *progress.Tracker {
	return Default().UploadProgress(label, totalBytes)
}

func Prompt(promptText, defaultVal string) string {
	return Default().Prompt(promptText, defaultVal)
}

func SecretPrompt(promptText string) string {
	return Default().SecretPrompt(promptText)
}

func Confirm(promptText string, defaultVal bool) bool {
	return Default().Confirm(promptText, defaultVal)
}

func Choice(promptText string, options []string, defaultChoice string) string {
	return Default().Choice(promptText, options, defaultChoice)
}
