package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	client = &http.Client{}
)

func init() {
	timeout := 30

	Transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // intentonal
		},
		Dial: (&net.Dialer{
			Timeout: time.Duration(timeout) * time.Second,
		}).Dial,
	}

	client = &http.Client{
		Transport: Transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}
}

func httpRequestWrapper(request *http.Request) (*http.Response, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		requestURL, _ := url.QueryUnescape(request.URL.String())

		return response, fmt.Errorf("unexpected status code %d received from %s", response.StatusCode, requestURL)
	}

	return response, nil
}

// HTTPRequest makes any HTTP request to a URL with extended parameters
func HTTPRequest(method, requestURL, cookies string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, requestURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "close")

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return httpRequestWrapper(req)
}

// Get makes a GET request to a URL with extended parameters
func Get(getURL, cookies string, headers map[string]string) (*http.Response, error) {
	return HTTPRequest(http.MethodGet, getURL, cookies, headers, nil)
}

// SimpleGet makes a simple GET request to a URL
func SimpleGet(getURL string) (*http.Response, error) {
	return HTTPRequest(http.MethodGet, getURL, "", map[string]string{}, nil)
}

// Post makes a POST request to a URL with extended parameters
func Post(postURL, cookies string, headers map[string]string, body io.Reader) (*http.Response, error) {
	return HTTPRequest(http.MethodPost, postURL, cookies, headers, body)
}

// SimplePost makes a simple POST request to a URL
func SimplePost(postURL, contentType string, body io.Reader) (*http.Response, error) {
	return HTTPRequest(http.MethodPost, postURL, "", map[string]string{"Content-Type": contentType}, body)
}
