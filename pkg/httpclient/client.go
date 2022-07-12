package httpclient

import (
	"io"
	"net"
	"net/http"
	"time"
)

// Client is a custom http.Client that includes the userAgent as
// required by rfc8555 (section 6.1)
type Client struct {
	http      http.Client
	userAgent string
}

// New creates a new http client using the custom struct. It also
// specifies timeout options so the client behaves sanely.
func New(userAgent string, devMode bool) (client *Client) {
	// make default timeouts lower
	dialTimeout := 5 * time.Second
	tlsTimeout := 5 * time.Second
	clientTimeout := 10 * time.Second

	// if development, allow higher timeouts
	if devMode {
		dialTimeout = 30 * time.Second
		tlsTimeout = 30 * time.Second
		clientTimeout = 30 * time.Second
	}

	// configure transport
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: dialTimeout,
		}).Dial,
		TLSHandshakeTimeout: tlsTimeout,
		// raise max connections per host
		MaxConnsPerHost:     25,
		MaxIdleConnsPerHost: 25,
	}

	// create *Client
	client = new(Client)
	client.http.Timeout = clientTimeout
	client.http.Transport = transport
	client.userAgent = userAgent

	return client
}

// newRequest creates an http request for the client to later do
func (client *Client) newRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// set user agent, required per RFC8555 6.1
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}

// do does the specified request
func (client *Client) do(request *http.Request) (*http.Response, error) {
	response, err := client.http.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Get does a get request to the specified url
func (client *Client) Get(url string) (*http.Response, error) {
	request, err := client.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return client.do(request)
}

// Head does a head request to the specified url
// a head request is the same as Get but without the body
func (client *Client) Head(url string) (*http.Response, error) {
	request, err := client.newRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return client.do(request)
}

// Post does a post request using the specified url, content type, and
// body
func (client *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	request, err := client.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", contentType)

	return client.do(request)
}
