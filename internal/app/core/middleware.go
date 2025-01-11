package core

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LoggingTransport struct {
	Transport http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqID := uuid.New().String()
	startTime := time.Now()

	// Log request
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	log.Printf("[%s] Request %s %s\n"+
		"Body: %s\n",
		reqID, req.Method, req.URL,
		string(reqBody))

	// Perform the request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		log.Printf("[%s] Request failed: %v", reqID, err)
		return nil, err
	}

	// Log response
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}

	duration := time.Since(startTime)
	log.Printf("[%s] Response completed in %v\n"+
		"Status: %s\n"+
		"Body: %s\n",
		reqID, duration,
		resp.Status,
		string(respBody))

	return resp, err
}
