package centrifugo_api

import (
	"context"
	"fmt"
	"github.com/centrifugal/gocent/v3"
	"net/http"
)

type CentrifugeAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type APICentrifugoClient struct {
	client *gocent.Client
}

func New(baseUrl string, apiKey string) *APICentrifugoClient {

	c := gocent.New(gocent.Config{
		Addr: fmt.Sprintf("%s/api", baseUrl),
		Key:  apiKey,
	})

	return &APICentrifugoClient{client: c}
}

func (api APICentrifugoClient) Publish(ctx context.Context, channel string, data []byte) (gocent.PublishResult, error) {
	return api.client.Publish(ctx, channel, data)
}

func (api APICentrifugoClient) Broadcast(ctx context.Context, channels []string, data []byte) (gocent.BroadcastResult, error) {
	return api.client.Broadcast(ctx, channels, data)
}

func (api APICentrifugoClient) Channels(ctx context.Context) (gocent.ChannelsResult, error) {
	return api.client.Channels(ctx)
}

func (api APICentrifugoClient) Disconnect(ctx context.Context, user string) error {
	return api.client.Disconnect(ctx, user)
}

func (api APICentrifugoClient) History(ctx context.Context, channel string) (gocent.HistoryResult, error) {
	return api.client.History(ctx, channel)
}

func (api APICentrifugoClient) HistoryRemove(ctx context.Context, channel string) error {
	return api.client.HistoryRemove(ctx, channel)
}

func (api APICentrifugoClient) Info(ctx context.Context) (gocent.InfoResult, error) {
	return api.client.Info(ctx)
}

func (api APICentrifugoClient) Pipe() *gocent.Pipe {
	return api.client.Pipe()
}
func (api APICentrifugoClient) Presence(ctx context.Context, channel string) (gocent.PresenceResult, error) {
	return api.client.Presence(ctx, channel)
}

func (api APICentrifugoClient) PresenceStats(ctx context.Context, channel string) (gocent.PresenceStatsResult, error) {
	return api.client.PresenceStats(ctx, channel)
}

func (api APICentrifugoClient) SendPipe(ctx context.Context, pipe *gocent.Pipe) ([]gocent.Reply, error) {
	return api.client.SendPipe(ctx, pipe)
}

func (api APICentrifugoClient) SetHTTPClient(httpClient *http.Client) {
	api.client.SetHTTPClient(httpClient)
}

func (api APICentrifugoClient) Unsubscribe(ctx context.Context, channel, user string) error {
	return api.client.Unsubscribe(ctx, channel, user)
}
