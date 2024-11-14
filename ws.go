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
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var defaultDialierHandshakeTimeoutInSeconds = 10 * time.Second

// ListenForWebSocketTextMessages is a blocking call that listens for messges. If there is an
// error, it exits and the error is returned. If a close message is received, the function
// exits and returns nil.
func ListenForWebSocketTextMessages(c *WebSocketConnection, messageHandler OnWebSocketTextMessage) error {
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		switch messageType {
		case websocket.TextMessage:
			messageHandler(message)
		case websocket.CloseMessage:
			return nil
		}
	}
}

func DefaultDialerConfig(url string) DialerConfig {
	return DialerConfig{
		Url:              url,
		HandshakeTimeout: 5 * time.Second,
		Proxy:            http.ProxyFromEnvironment,
	}
}

func DialWebSocket(ctx context.Context, config DialerConfig) (*WebSocketConnection, error) {

	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

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

	if dialer.HandshakeTimeout <= 0 {
		dialer.HandshakeTimeout = defaultDialierHandshakeTimeoutInSeconds
	}

	c, _, err := dialer.DialContext(ctx, u.String(), config.RequestHeader)
	if err != nil {
		return nil, err
	}

	return &WebSocketConnection{conn: c}, nil
}
