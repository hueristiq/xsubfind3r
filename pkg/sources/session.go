package sources

import (
	"fmt"
	"io"

	"github.com/valyala/fasthttp"
)

func NewSession(domain string, keys *Keys) (*Session, error) {
	var session *Session

	client := &fasthttp.Client{}

	extractor, err := NewExtractor(domain)
	if err != nil {
		return session, err
	}

	return &Session{
		Client:    client,
		Keys:      keys,
		Extractor: extractor,
	}, nil
}

// SimpleGet makes a simple GET request to a URL
func (session *Session) SimpleGet(getURL string) (*fasthttp.Response, error) {
	return session.Request(fasthttp.MethodGet, getURL, "", map[string]string{}, nil)
}

// Request makes any HTTP request to a URL with extended parameters
func (session *Session) Request(method, requestURL, cookies string, headers map[string]string, body io.Reader) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()

	req.SetRequestURI(requestURL)

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return httpRequestWrapper(session.Client, req)
}

func httpRequestWrapper(client *fasthttp.Client, req *fasthttp.Request) (*fasthttp.Response, error) {
	res := fasthttp.AcquireResponse()

	client.Do(req, res)

	if res.StatusCode() != fasthttp.StatusOK {
		return res, fmt.Errorf("Unexpected status code")
	}

	return res, nil
}
