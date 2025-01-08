package nws

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"prediction-risk/internal/infrastructure/external"
)

type client struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
}

func newClient(baseURL string, userAgent string) *client {
	return &client{
		baseURL:   baseURL,
		userAgent: userAgent,
		httpClient: &http.Client{
			Transport: &external.LoggingTransport{Transport: http.DefaultTransport},
		},
	}
}

func (kc *client) queryGeneration(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Set(k, v)
		}
	}
	query := values.Encode()
	if query != "" {
		query = "?" + query
	}
	return query
}

func (c *client) get(path string, params map[string]string) (*http.Response, error) {
	fullUrl := c.baseURL + path
	if params != nil && len(params) > 0 {
		query := c.queryGeneration(params)
		fullUrl += query
	}

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, err
	}

	headers := c.requestHeaders()
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *client) requestHeaders() map[string]string {

	headers := map[string]string{
		"Accept":     "application/geo+json",
		"User-Agent": c.userAgent,
	}

	return headers
}

func handleResponse[T any](resp *http.Response) (*T, error) {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &NWSError{
			Reason:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	var result T
	err := json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close() // Close after decoding
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}
