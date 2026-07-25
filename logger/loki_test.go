package logger

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aslani/wotop"
	"github.com/stretchr/testify/require"
)

func TestNewLokiLog_AppendsPushEndpoint(t *testing.T) {
	endpoint, err := lokiEndpoint("http://localhost:3100")

	require.NoError(t, err)
	require.Equal(t, "http://localhost:3100/loki/api/v1/push", endpoint)
}

func TestLokiLog_PushesLogEntry(t *testing.T) {
	var payload lokiPushPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, lokiPushPath, r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	appData := wotop.ApplicationData{
		AppName:       "product",
		AppInstanceID: "abcd",
		StartTime:     "2026-07-25 10:00:00",
	}

	log, err := NewLokiLog(server.URL, appData, "DEVELOPMENT")
	require.NoError(t, err)

	log.Info(SetTraceID(context.Background(), "trace-1"), "created product %d", 12)

	require.NoError(t, log.Sync())
	require.Len(t, payload.Streams, 1)
	require.Equal(t, "product", payload.Streams[0].Stream["app"])
	require.Equal(t, "abcd", payload.Streams[0].Stream["app_inst_id"])
	require.Equal(t, "development", payload.Streams[0].Stream["stage"])
	require.Equal(t, "info", payload.Streams[0].Stream["severity"])
	require.Len(t, payload.Streams[0].Values, 1)

	var line jsonLogModel
	require.NoError(t, json.Unmarshal([]byte(payload.Streams[0].Values[0][1]), &line))
	require.Equal(t, "product", line.AppName)
	require.Equal(t, "abcd", line.AppInstID)
	require.Equal(t, "INFO", line.Severity)
	require.Equal(t, "trace-1 created product 12", line.Message)
	require.NotEmpty(t, line.Location)
	require.NotEmpty(t, line.Time)
}

func TestLokiLog_StoresPushError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	log, err := NewLokiLog(server.URL, wotop.ApplicationData{AppName: "product"}, "production")
	require.NoError(t, err)

	log.Error(context.Background(), "failed")

	require.ErrorContains(t, log.Sync(), "loki push failed")
}
