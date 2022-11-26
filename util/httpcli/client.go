// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/klauspost/compress/gzhttp"

	"github.com/ldclabs/ldvm/util/erring"
)

type ctxKey int

const (
	// CtxHeaderKey ...
	ctxHeaderKey ctxKey = 0
)

const (
	MaxContentLength int64 = 1024 * 1024 * 10
)

// Client ...
type Client struct {
	cli http.Client
}

// NewClient ...
func NewClient(rt http.RoundTripper) *Client {
	httpClient := http.DefaultClient
	if rt != nil {
		httpClient = &http.Client{Transport: gzhttp.Transport(rt)}
	}

	return &Client{cli: *httpClient}
}

// Do ...
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.cli.Do(req)
}

func (c *Client) DoAndRead(req *http.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.DoWithReader(req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) DoWithReader(req *http.Request, br io.ReaderFrom) error {
	resp, err := c.cli.Do(req)
	if err != nil {
		return &erring.Error{
			Code:    499,
			Message: fmt.Sprintf("http request failed, %v", err),
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, er := io.ReadAll(resp.Body)
		err := &erring.Error{
			Code:    resp.StatusCode,
			Message: http.StatusText(resp.StatusCode),
			Data: map[string]interface{}{
				"reqHeader":  req.Header,
				"respHeader": resp.Header,
				"respBody":   string(data),
			},
		}
		err.CatchIf(er)
		return err
	}

	if br == nil || req.Method == "HEAD" {
		return nil
	}

	if g, ok := br.(interface{ Grow(n int) }); ok {
		cl := resp.ContentLength
		if xcl := resp.Header.Get("x-content-length"); xcl != "" {
			if x, _ := strconv.ParseInt(xcl, 10, 64); x > 0 {
				cl = x
			}
		}
		if cl > 0 {
			g.Grow(int(cl))
		}
	}

	if _, err = br.ReadFrom(resp.Body); err != nil {
		return &erring.Error{
			Code:    resp.StatusCode,
			Message: err.Error(),
			Data: map[string]interface{}{
				"reqHeader":  req.Header,
				"respHeader": resp.Header,
			},
		}
	}

	return nil
}

func CtxWithHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, ctxHeaderKey, header)
}

func HeaderCtxValue(ctx context.Context) http.Header {
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
