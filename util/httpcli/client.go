// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	var buf bytes.Buffer
	if err := c.DoWith(req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) DoWith(req *http.Request, br BodyReader) error {
	req.Header.Set("accept-encoding", "gzip")
	resp, err := c.Client.Do(req)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	defer resp.Body.Close()
	body := resp.Body
	cl := resp.ContentLength
	if resp.Header.Get("content-encoding") == "gzip" {
		body, err = gzip.NewReader(body)
		if err != nil {
			return &Error{
				Code:    resp.StatusCode,
				Message: err.Error(),
				Header:  req.Header,
			}
		}

		if xcl := resp.Header.Get("x-content-length"); xcl != "" {
			if x, _ := strconv.ParseInt(xcl, 10, 64); x > 0 {
				cl = x
			}
		}
		defer body.Close()
	}

	if cl > MaxContentLength {
		return &Error{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("content length too large, expected <= %d", MaxContentLength),
			Header:  req.Header,
		}
	}

	if resp.StatusCode != http.StatusOK {
		errstr := http.StatusText(resp.StatusCode)
		data, err := io.ReadAll(body)
		if err != nil {
			errstr += ", read body error, " + err.Error()
		}

		return &Error{
			Code:    resp.StatusCode,
			Message: errstr,
			Header:  req.Header,
			Body:    string(data),
		}
	}

	if g, ok := br.(Growable); ok && cl > 0 {
		g.Grow(int(cl))
	}

	if _, err = br.ReadFrom(body); err != nil {
		return &Error{
			Code:    resp.StatusCode,
			Message: err.Error(),
			Header:  req.Header,
		}
	}

	return nil
}

type Growable interface {
	Grow(n int)
}

type BodyReader interface {
	ReadFrom(r io.Reader) (n int64, err error)
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
