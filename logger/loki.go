package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/a-aslani/wotop"
)

const lokiPushPath = "/loki/api/v1/push"

// lokiModel sends log entries to Grafana Loki using its HTTP push API.
type lokiModel struct {
	appData wotop.ApplicationData
	stage   string

	client   *http.Client
	endpoint string

	mu      sync.Mutex
	lastErr error
}

// NewLokiLog creates a logger that pushes JSON log lines to Grafana Loki.
//
// The lokiAddress may be a base URL such as "http://localhost:3100" or the
// full push endpoint "http://localhost:3100/loki/api/v1/push".
func NewLokiLog(lokiAddress string, appData wotop.ApplicationData, stage string) (*lokiModel, error) {
	endpoint, err := lokiEndpoint(lokiAddress)
	if err != nil {
		return nil, err
	}

	return &lokiModel{
		appData: appData,
		stage:   stage,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		endpoint: endpoint,
	}, nil
}

// Error logs an error message with optional arguments.
func (l *lokiModel) Error(ctx context.Context, message string, args ...any) {
	l.log(ctx, "ERROR", message, args...)
}

// Info logs an informational message with optional arguments.
func (l *lokiModel) Info(ctx context.Context, message string, args ...any) {
	l.log(ctx, "INFO", message, args...)
}

// Warning logs a warning message with optional arguments.
func (l *lokiModel) Warning(ctx context.Context, message string, args ...any) {
	l.log(ctx, "WARNING", message, args...)
}

// Sync returns the last delivery error observed by the Loki logger.
func (l *lokiModel) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastErr
}

func (l *lokiModel) log(ctx context.Context, severity string, message string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	messageWithArgs := fmt.Sprintf(message, args...)
	location := getFileLocationInfo(3)
	traceID := GetTraceID(ctx)

	line := toJsonString(jsonLogModel{
		AppName:   l.appData.AppName,
		AppInstID: l.appData.AppInstanceID,
		Start:     l.appData.StartTime,
		Severity:  severity,
		Message:   fmt.Sprintf("%s %s", traceID, messageWithArgs),
		Location:  location,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
	})

	payload := toJsonString(lokiPushPayload{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"app":         l.appData.AppName,
					"app_inst_id": l.appData.AppInstanceID,
					"stage":       strings.ToLower(strings.TrimSpace(l.stage)),
					"severity":    strings.ToLower(severity),
				},
				Values: [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), line},
				},
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.endpoint, bytes.NewBufferString(payload))
	if err != nil {
		l.setErr(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		l.setErr(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		l.setErr(fmt.Errorf("loki push failed with status %s", resp.Status))
		return
	}

	l.setErr(nil)
}

func (l *lokiModel) setErr(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lastErr = err
}

type lokiPushPayload struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func lokiEndpoint(lokiAddress string) (string, error) {
	trimmed := strings.TrimSpace(lokiAddress)
	if trimmed == "" {
		return "", errors.New("loki address is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid loki address %q", lokiAddress)
	}
	if strings.HasSuffix(parsed.Path, lokiPushPath) {
		return parsed.String(), nil
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + lokiPushPath
	return parsed.String(), nil
}
