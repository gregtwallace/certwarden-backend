package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a custom http.Client that includes the userAgent as
// required by rfc8555 (section 6.1)
type Client struct {
	http      http.Client
	userAgent string
}

// New creates a new Client.  Client is just an http.Client with some wrapping to
// add user-agent and accept-language headers
func New(userAgent string) (client *Client) {
	// create *Client
	client = &Client{
		http: http.Client{
			// set client timeout
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport,
		},
		userAgent: userAgent,
	}

	return client
}

// do creates a request with the specified parameters, modifies it in accord with ACME
// spec and then executes the request
func (c *Client) do(method string, url string, body io.Reader, addlHeader http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// return err if invalid scheme
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return nil, fmt.Errorf("invalid scheme (%s) in http client request to url (%s); 'http' or 'https' must be explicitly specified", req.URL.Scheme, req.URL.String())
	}

	// add any additionally specified headers
	for k, v := range addlHeader {
		for _, i := range v {
			req.Header.Add(k, i)
		}
	}

	// use Set below to ensure override of any 'addlHeaders' that may conflict

	// set user agent, required per RFC8555 6.1
	req.Header.Set("User-Agent", c.userAgent)

	// set preferred language (SHOULD do this per RFC 8555, 6.1)
	// TODO: Implement user choice?
	req.Header.Set("Accept-Language", "en-US, en;q=0.8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Get does a get request to the specified url and additionally
// specified headers.
func (c *Client) GetWithHeader(url string, header http.Header) (*http.Response, error) {
	return c.do(http.MethodGet, url, nil, header)
}

// Get does a get request to the specified url
func (c *Client) Get(url string) (*http.Response, error) {
	return c.GetWithHeader(url, nil)
}

// Head does a head request to the specified url
// a head request is the same as Get but without the body
func (c *Client) Head(url string) (*http.Response, error) {
	return c.do(http.MethodHead, url, nil, nil)
}

// PostWithHeader does a post request using the specified url, content type, body,
// and additionally specified headers.
func (c *Client) PostWithHeader(url string, contentType string, body io.Reader, header http.Header) (resp *http.Response, err error) {
	// if no headers, make empty for Content-Type
	if header == nil {
		header = make(http.Header)
	}

	// explicitly set (override) content type header
	header.Set("Content-Type", contentType)

	return c.do(http.MethodPost, url, body, header)
}

// Post does a post request using the specified url, content type, and body
func (c *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return c.PostWithHeader(url, contentType, body, nil)
}
