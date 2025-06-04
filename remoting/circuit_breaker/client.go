package circuit_breaker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/a-aslani/wotop/logger"
	"github.com/sony/gobreaker"
	"io"
	"net/http"
	"time"
)

type Client struct {
	log        logger.Logger
	baseURL    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

type Authentication struct {
	ApiKey, SecretKey string
}

type ClientConfig struct {
	BaseURL          string
	Timeout          time.Duration
	MaxFailures      uint32
	IntervalDuration time.Duration
	TimeoutDuration  time.Duration
}

type Response struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func NewClient(name string, log logger.Logger, cfg ClientConfig) *Client {
	cbSettings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    cfg.IntervalDuration,
		Timeout:     cfg.TimeoutDuration,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= cfg.MaxFailures && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// You could add logging here
		},
	}

	return &Client{
		log:     log,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *Client) Execute(ctx context.Context, auth Authentication, method, path string, body interface{}) ([]byte, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		var reqBody []byte
		var err error

		if body != nil {
			reqBody, err = json.Marshal(body)
			if err != nil {
				c.log.Error(ctx, "failed to marshal request body: %s", err.Error())
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(reqBody))
		if err != nil {
			c.log.Error(ctx, "failed to create request: %s", err.Error())
			return nil, err
		}

		c.setHeaders(req, auth.ApiKey, auth.SecretKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.log.Error(ctx, "failed to execute request: %s", err.Error())
			return nil, err
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.log.Error(ctx, "failed to read response body: %s", err.Error())
			return nil, err
		}

		if resp.StatusCode >= 400 {

			if resp.StatusCode >= 500 {

				var response Response

				err = json.Unmarshal(responseBody, &response)
				if err != nil {
					c.log.Error(ctx, "failed to unmarshal response body: %s", err.Error())
					return nil, err
				}

				if !response.Success {
					c.log.Error(ctx, "service returned error status: %d, errorMsg: %s", resp.StatusCode, response.ErrorMessage)
					return nil, fmt.Errorf("service returned error status: %d, errorMsg: %s", resp.StatusCode, response.ErrorMessage)
				}
			}

			c.log.Error(ctx, "service returned error status: %d", resp.StatusCode)
			return nil, fmt.Errorf("service returned error status: %d", resp.StatusCode)
		}

		return responseBody, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (c *Client) basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) setHeaders(req *http.Request, apiKey, secretKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth(apiKey, secretKey))
}
