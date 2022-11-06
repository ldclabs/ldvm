// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ctxKey int

const (
	// CtxHeaderKey ...
	ctxHeaderKey ctxKey = 0
)

// Client ...
type Client struct {
	http.Client
	Header http.Header
}

// NewClient ...
func NewClient(rt http.RoundTripper) *Client {
	httpClient := http.DefaultClient
	if rt != nil {
		httpClient = &http.Client{Transport: rt}
	}

	return &Client{Client: *httpClient, Header: http.Header{}}
}

// Do ...
func (c *Client) Do(req *http.Request) ([]byte, error) {
	req.Header.Set("accept-encoding", "gzip")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do http request error, %v", err)
	}

	defer resp.Body.Close()
	body := resp.Body
	if resp.Header.Get("content-encoding") == "gzip" {
		body, err = gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error, %v", err)
		}
		defer body.Close()
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response error, %s, status code, %v", err.Error(), resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code, %v, response, %q", resp.StatusCode, string(data))
	}
	return data, nil
}

func CtxWithHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, ctxHeaderKey, header)
}

func HeaderFromCtx(ctx context.Context) http.Header {
	if val := ctx.Value(ctxHeaderKey); val != nil {
		return val.(http.Header)
	}
	return nil
}

func CopyHeader(dst http.Header, src http.Header) {
	for k, vv := range src {
		switch len(vv) {
		case 1:
			dst.Set(k, vv[0])
		default:
			dst.Del(k)
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}
}
