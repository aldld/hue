package hue

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tmaxmax/go-sse"
	"golang.org/x/exp/slog"
)

const (
	hueAppKeyHeader = "hue-application-key"
)

type ErrorResponse struct {
	Errors []HueError `json:"errors"`
}

type HueError struct {
	Description string `json:"description"`
}

func (e HueError) Error() string {
	return e.Description
}

func joinHueErrors(hueErrors []HueError) error {
	errs := make([]error, len(hueErrors))
	for i, e := range hueErrors {
		errs[i] = e
	}
	return errors.Join(errs...)
}

type Config struct {
	Addr   string
	AppKey string
}

type Client struct {
	Config

	log        *slog.Logger
	httpClient *http.Client
	sseClient  *sse.Client
}

func NewClient(log *slog.Logger, config Config) *Client {
	// Skip certificate verification. TODO: Make this work properly.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{Transport: transport}

	sseClient := &sse.Client{HTTPClient: httpClient}

	return &Client{
		Config:     config,
		log:        log,
		httpClient: httpClient,
		sseClient:  sseClient,
	}
}

func (c *Client) absURL(endpoint string) string {
	return fmt.Sprintf("https://%s%s", c.Addr, endpoint)
}

func (c *Client) resourceURL(endpoint string) string {
	return c.absURL("/clip/v2/resource" + endpoint)
}

func (c *Client) get(endpoint string, response any) error {
	req, err := http.NewRequest(http.MethodGet, c.resourceURL(endpoint), nil)
	if err != nil {
		return err
	}
	return c.do(req, response)
}

func (c *Client) put(endpoint string, body any, response any) error {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(bodyJson)

	req, err := http.NewRequest(http.MethodPut, c.resourceURL(endpoint), bodyReader)
	if err != nil {
		return err
	}
	return c.do(req, response)
}

func (c *Client) do(req *http.Request, response any) error {
	req.Header.Add(hueAppKeyHeader, c.AppKey)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	c.log.Debug("request complete",
		slog.String("status", res.Status),
		slog.String("url", req.URL.String()),
		slog.String("method", req.Method),
	)

	dec := json.NewDecoder(res.Body)
	if res.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := dec.Decode(&errResp); err != nil {
			return err
		}
		err := joinHueErrors(errResp.Errors)

		c.log.Error("request error", slog.Any("error", err))
		return err
	}

	if err := dec.Decode(response); err != nil {
		return err
	}

	return nil
}
