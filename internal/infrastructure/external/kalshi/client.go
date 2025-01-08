package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"prediction-risk/internal/infrastructure/external"
	"strings"
	"sync"
	"time"
)

const (
	baseAPIPath   = "/trade-api/v2"
	portfolioPath = baseAPIPath + "/portfolio"
	exchangePath  = baseAPIPath + "/exchange"
	marketsPath   = baseAPIPath + "/markets"
	eventsPath    = baseAPIPath + "/events"
)

/*
Represents a base client to interact with the Kalshi API
Requires:
- A vlid API key
- Base URL representing the environment
- Underlying HTTP client
*/
type client struct {
	host        string
	keyID       string
	privateKey  *rsa.PrivateKey
	lastAPICall time.Time
	httpClient  *http.Client
	mutex       sync.Mutex // For thread-safe rate limiting
}

func newClient(host, keyID string, privateKey *rsa.PrivateKey) *client {
	return &client{
		host:        host,
		keyID:       keyID,
		privateKey:  privateKey,
		lastAPICall: time.Now(),
		httpClient: &http.Client{
			Transport: &external.LoggingTransport{Transport: http.DefaultTransport},
		},
	}
}

func (kc *client) rateLimit() {
	kc.mutex.Lock()
	defer kc.mutex.Unlock()

	const thresholdInMilliseconds = 100
	threshold := time.Duration(thresholdInMilliseconds) * time.Millisecond

	now := time.Now()
	if now.Sub(kc.lastAPICall) < threshold {
		time.Sleep(threshold - now.Sub(kc.lastAPICall))
	}
	kc.lastAPICall = time.Now()
}

func (kc *client) get(path string, params map[string]string) (*http.Response, error) {
	kc.rateLimit()

	fullURL := kc.host + path
	if params != nil && len(params) > 0 {
		query := kc.queryGeneration(params)
		fullURL += query
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	headers, err := kc.requestHeaders("GET", path)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return kc.httpClient.Do(req)
}

func (kc *client) post(path string, body interface{}) (*http.Response, error) {
	kc.rateLimit()

	fullURL := kc.host + path

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}

	headers, err := kc.requestHeaders("POST", path)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return kc.httpClient.Do(req)
}

func (kc *client) requestHeaders(method, path string) (map[string]string, error) {
	currentTimeMilliseconds := time.Now().UnixNano() / int64(time.Millisecond)
	timestampStr := fmt.Sprintf("%d", currentTimeMilliseconds)

	// Remove query params from path
	pathParts := strings.Split(path, "?")

	// Construct the message string
	msgString := timestampStr + method + pathParts[0]

	// Sign the message
	signature, err := kc.signPSSText(msgString)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":            "application/json",
		"KALSHI-ACCESS-KEY":       kc.keyID,
		"KALSHI-ACCESS-SIGNATURE": signature,
		"KALSHI-ACCESS-TIMESTAMP": timestampStr,
	}

	return headers, nil
}

func (kc *client) signPSSText(text string) (string, error) {
	message := []byte(text)

	hashed := sha256.Sum256(message)

	signature, err := rsa.SignPSS(rand.Reader, kc.privateKey, crypto.SHA256, hashed[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	})
	if err != nil {
		return "", fmt.Errorf("RSA sign PSS failed: %v", err)
	}

	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	return signatureBase64, nil
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

func handleResponse[T any](resp *http.Response) (*T, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &KalshiError{
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
