package sources

import (
	"fmt"

	"github.com/corpix/uarand"
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

// Get makes a GET request to a URL with extended parameters
func (session *Session) Get(URL, cookies string, headers map[string]string) (*fasthttp.Response, error) {
	return session.Request(fasthttp.MethodGet, URL, cookies, headers, nil)
}

// SimpleGet makes a simple GET request to a URL
func (session *Session) SimpleGet(URL string) (*fasthttp.Response, error) {
	return session.Request(fasthttp.MethodGet, URL, "", map[string]string{}, nil)
}

// Post makes a POST request to a URL with extended parameters
func (session *Session) Post(URL, cookies string, headers map[string]string, body []byte) (*fasthttp.Response, error) {
	return session.Request(fasthttp.MethodPost, URL, cookies, headers, body)
}

// SimplePost makes a simple POST request to a URL
func (session *Session) SimplePost(URL, contentType string, body []byte) (*fasthttp.Response, error) {
	return session.Request(fasthttp.MethodPost, URL, "", map[string]string{"Content-Type": contentType}, body)
}

// Request makes any HTTP request to a URL with extended parameters
func (session *Session) Request(method, URL, cookies string, headers map[string]string, body []byte) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()

	req.SetRequestURI(URL)
	req.SetBody(body)
	req.Header.SetMethod(method)

	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "close")

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
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
