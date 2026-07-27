package consolehub

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aognio/consolehub/libraries/go/consolehub/events"
	"github.com/aognio/consolehub/libraries/go/consolehub/progress"
	"github.com/aognio/consolehub/libraries/go/consolehub/prompt"
	"github.com/aognio/consolehub/libraries/go/consolehub/queue"
	"github.com/aognio/consolehub/libraries/go/consolehub/transport"
	"github.com/aognio/consolehub/libraries/go/consolehub/worker"
	"github.com/aognio/consolehub/libraries/go/consolehub/writer"
)

type Client struct {
	opts    Options
	q       *queue.BoundedQueue
	w       *worker.Worker
	stdoutW *writer.StreamWriter
	stderrW *writer.StreamWriter
	cancel  context.CancelFunc
}

func New(optFns ...Option) (*Client, error) {
	opts := defaultOptions()
	for _, fn := range optFns {
		fn(&opts)
	}

	q := queue.New(opts.QueueCapacity)

	var workerInst *worker.Worker
	ctx, cancel := context.WithCancel(context.Background())

	if !opts.Disabled {
		trans := opts.Transport
		if trans == nil {
			trans = transport.NewWebSocketTransport()
		}

		wOpts := worker.WorkerOptions{
			Endpoint:         opts.Endpoint,
			Token:            opts.Token,
			Tenant:           opts.Tenant,
			App:              opts.App,
			ClientRunID:      opts.GetClientRunID(),
			Hostname:         opts.Hostname,
			PID:              opts.PID,
			AppVersion:       opts.AppVersion,
			OSName:           opts.OSName,
			Environment:      opts.Env,
			MaxBatchSize:     250,
			HeartbeatSeconds: 30,
		}

		workerInst = worker.New(wOpts, trans, q)
		workerInst.Start(ctx)
	}

	c := &Client{
		opts:    opts,
		q:       q,
		w:       workerInst,
		stdoutW: writer.Stdout(q),
		stderrW: writer.Stderr(q),
		cancel:  cancel,
	}

	return c, nil
}

func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.w != nil {
		c.w.Stop()
	}
	c.stdoutW.Flush()
	c.stderrW.Flush()
	return nil
}

// Writer returns an io.Writer targeting a specific stream name (defaults to "stdout").
func (c *Client) Writer(streamName ...string) io.Writer {
	name := events.StreamStdout
	if len(streamName) > 0 && streamName[0] != "" {
		name = streamName[0]
	}
	return writer.New(name, c.q, os.Stdout)
}

func (c *Client) Stdout() io.Writer { return c.stdoutW }
func (c *Client) Stderr() io.Writer { return c.stderrW }

// Print methods
func (c *Client) Print(v ...any) {
	str := fmt.Sprint(v...)
	fmt.Print(str)
	c.q.Push(events.NewTextLine(events.StreamStdout, str))
}

func (c *Client) Printf(format string, v ...any) {
	str := fmt.Sprintf(format, v...)
	fmt.Print(str)
	c.q.Push(events.NewTextLine(events.StreamStdout, str))
}

func (c *Client) Println(v ...any) {
	str := fmt.Sprintln(v...)
	fmt.Print(str)
	c.q.Push(events.NewTextLine(events.StreamStdout, str))
}

func (c *Client) Fprint(w io.Writer, v ...any) {
	str := fmt.Sprint(v...)
	_, _ = w.Write([]byte(str))
}

func (c *Client) Fprintf(w io.Writer, format string, v ...any) {
	str := fmt.Sprintf(format, v...)
	_, _ = w.Write([]byte(str))
}

func (c *Client) Fprintln(w io.Writer, v ...any) {
	str := fmt.Sprintln(v...)
	_, _ = w.Write([]byte(str))
}

// Structured Logging
func (c *Client) Debug(msg string) { c.log("debug", msg, nil) }
func (c *Client) Info(msg string)  { c.log("info", msg, nil) }
func (c *Client) Warn(msg string)  { c.log("warn", msg, nil) }
func (c *Client) Error(msg string) { c.log("error", msg, nil) }

func (c *Client) Debugf(format string, v ...any) { c.log("debug", fmt.Sprintf(format, v...), nil) }
func (c *Client) Infof(format string, v ...any)  { c.log("info", fmt.Sprintf(format, v...), nil) }
func (c *Client) Warnf(format string, v ...any)  { c.log("warn", fmt.Sprintf(format, v...), nil) }
func (c *Client) Errorf(format string, v ...any) { c.log("error", fmt.Sprintf(format, v...), nil) }

func (c *Client) Log(evt events.Event) {
	if evt != nil {
		c.q.Push(evt)
	}
}

func (c *Client) Report(evt events.Event) {
	c.Log(evt)
}

func (c *Client) log(level, message string, fields map[string]any) {
	evt := events.NewLogEvent(level, message, fields)
	fmt.Printf("[%s] %s\n", level, message)
	c.q.Push(evt)
}

// Progress Trackers
func (c *Client) Progress(label string, total int64) *progress.Tracker {
	id := fmt.Sprintf("p-%d", c.q.NextSequence())
	return progress.New(id, label, total, c.q)
}

func (c *Client) UploadProgress(label string, totalBytes int64) *progress.Tracker {
	return c.Progress(label, totalBytes)
}

// Interactive Console Prompts
func (c *Client) Prompt(promptText, defaultVal string) string {
	return prompt.Prompt(promptText, defaultVal, c.q)
}

func (c *Client) SecretPrompt(promptText string) string {
	return prompt.SecretPrompt(promptText, c.q)
}

func (c *Client) Confirm(promptText string, defaultVal bool) bool {
	return prompt.Confirm(promptText, defaultVal, c.q)
}

func (c *Client) Choice(promptText string, options []string, defaultChoice string) string {
	return prompt.Choice(promptText, options, defaultChoice, c.q)
}
