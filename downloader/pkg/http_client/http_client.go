package http_client

import (
	"io"
	"net/http"
)

type HTTPClient struct {
	Client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{Client: &http.Client{
		Timeout: 0,
	}}
}

func (h *HTTPClient) CreateRequest(method string, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (h *HTTPClient) CreateOctetStreamRequest(method string, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (h *HTTPClient) DoRequest(req *http.Request) ([]byte, int, error) {
	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, 500, err
	}
	defer resp.Body.Close()
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}
	return dat, resp.StatusCode, nil
}
