/**
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var defaultDialierHandshakeTimeoutInSeconds = 45 * time.Second

type WebSocketBufferPool interface {
	// Get gets a value from the pool or returns nil if the pool is empty.
	Get() interface{}
	// Put adds a value to the pool.
	Put(interface{})
}

type DialerConfig struct {
	Url string

	RequestHeader http.Header

	// NetDial specifies the dial function for creating TCP connections. If
	// NetDial is nil, net.Dialer DialContext is used.
	NetDial func(network, addr string) (net.Conn, error)

	// NetDialContext specifies the dial function for creating TCP connections. If
	// NetDialContext is nil, NetDial is used.
	NetDialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	// NetDialTlsContext specifies the dial function for creating TLS/TCP connections. If
	// NetDialTlsContext is nil, NetDialContext is used.
	// If NetDialTlsContext is set, Dial assumes the TLS handshake is done there and
	// TLSClientConfig is ignored.
	NetDialTlsContext func(ctx context.Context, network, addr string) (net.Conn, error)

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	Proxy func(*http.Request) (*url.URL, error)

	// TlsClientConfig specifies the TLS configuration to use with tls.Client.
	// If nil, the default configuration is used.
	// If either NetDialTLS or NetDialTLSContext are set, Dial assumes the TLS handshake
	// is done there and TLSClientConfig is ignored.
	TlsClientConfig *tls.Config

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then a useful default size is used. The I/O buffer sizes
	// do not limit the size of the messages that can be sent or received.
	ReadBufferSize, WriteBufferSize int

	// WriteBufferPool is a pool of buffers for write operations. If the value
	// is not set, then write buffers are allocated to the connection for the
	// lifetime of the connection.
	//
	// A pool is most useful when the application has a modest volume of writes
	// across a large number of connections.
	//
	// Applications should use a single pool for each unique value of
	// WriteBufferSize.
	WriteBufferPool WebSocketBufferPool

	// Subprotocols specifies the client's requested subprotocols.
	Subprotocols []string

	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool

	// Jar specifies the cookie jar.
	// If Jar is nil, cookies are not sent in requests and ignored
	// in responses.
	Jar http.CookieJar
}

func DialWebSocket(ctx context.Context, config DialerConfig) (*websocket.Conn, error) {

	u := url.URL{Scheme: "wss", Host: config.Url}

	var dialer = &websocket.Dialer{
		NetDial:           config.NetDial,
		NetDialContext:    config.NetDialContext,
		NetDialTLSContext: config.NetDialTlsContext,
		Proxy:             config.Proxy,
		HandshakeTimeout:  config.HandshakeTimeout,
		TLSClientConfig:   config.TlsClientConfig,
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		WriteBufferPool:   config.WriteBufferPool,
		Subprotocols:      config.Subprotocols,
		EnableCompression: config.EnableCompression,
		Jar:               config.Jar,
	}

	if dialer.Proxy == nil {
		dialer.Proxy = http.ProxyFromEnvironment
	}

	if dialer.HandshakeTimeout == 0 {
		dialer.HandshakeTimeout = defaultDialierHandshakeTimeoutInSeconds
	}

	c, _, err := dialer.DialContext(ctx, u.String(), config.RequestHeader)
	if err != nil {
		return nil, err
	}

	return c, nil
}
