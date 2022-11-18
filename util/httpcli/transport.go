// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

type TransportOptions struct {
	MaxIdleConnsPerHost   int
	InsecureSkipVerify    bool
	DisableHTTP2          bool
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	ReadIdleTimeout       time.Duration
	PingTimeout           time.Duration
}

func NewRoundTripper(cfg *TransportOptions) (http.RoundTripper, error) {
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ResponseHeaderTimeout <= 0 {
		cfg.ResponseHeaderTimeout = 10 * time.Second
	}
	if cfg.IdleConnTimeout <= 0 {
		cfg.IdleConnTimeout = 59 * time.Second
	}
	if cfg.ReadIdleTimeout <= 0 {
		cfg.ReadIdleTimeout = cfg.IdleConnTimeout
	}
	if cfg.MaxIdleConnsPerHost <= 0 {
		cfg.MaxIdleConnsPerHost = 10
	}

	dialer := &net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: 25 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,

		ReadBufferSize:  64 * 1024,
		WriteBufferSize: 64 * 1024,
	}

	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
	}

	if cfg.DisableHTTP2 {
		return transport, nil
	}

	transportHTTP1 := transport.Clone()
	transportHTTP2, err := http2.ConfigureTransports(transport)
	if err != nil {
		return nil, err
	}

	transportHTTP2.ReadIdleTimeout = cfg.ReadIdleTimeout
	transportHTTP2.PingTimeout = cfg.PingTimeout

	transportH2C := &h2cTransportWrapper{
		Transport: &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			AllowHTTP: true,
		},
	}

	transportH2C.ReadIdleTimeout = cfg.ReadIdleTimeout
	transportH2C.PingTimeout = cfg.PingTimeout
	transport.RegisterProtocol("h2c", transportH2C)

	return &mixRoundTripper{
		http2: transport,
		http:  transportHTTP1,
	}, nil
}

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

type mixRoundTripper struct {
	http2 *http.Transport
	http  *http.Transport
}

func (m *mixRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Don't use HTTP/2 if it is a connection upgrade
	if httpguts.HeaderValuesContainsToken(req.Header.Values("connection"), "Upgrade") {
		return m.http.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
