package httpclient

import "github.com/cloudflare/cloudflare-go"

// AsCloudflareOptions returns the client as options that can be used in the Cloudflare API
func (client *Client) AsCloudflareOptions() (opts []cloudflare.Option) {
	return []cloudflare.Option{
		cloudflare.HTTPClient(&client.http),
		cloudflare.UserAgent(client.userAgent),
	}
}
