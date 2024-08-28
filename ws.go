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
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketConnetion struct {
	conn *websocket.Conn
}

func (c *WebSocketConnetion) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WebSocketConnetion) SetReadLimit(limit int64) {
	c.conn.SetReadLimit(limit)
}

func (c *WebSocketConnetion) CloseHandler() func(code int, text string) error {
	return c.conn.CloseHandler()
}

func (c *WebSocketConnetion) SetCloseHandler(h func(code int, text string) error) {
	c.SetCloseHandler(h)
}

func (c *WebSocketConnetion) NetConn() net.Conn {
	return c.NetConn()
}

func (c *WebSocketConnetion) ReadMessage() (messageType int, p []byte, err error) {
	return c.ReadMessage()
}

func (c *WebSocketConnetion) Close() error {
	return c.Close()
}

func (c *WebSocketConnetion) RemoteAddr() net.Addr {
	return c.RemoteAddr()
}

func (c *WebSocketConnetion) Subprotocol() string {
	return c.Subprotocol()
}

type OnWebSocketBinaryMessage func(message []byte) bool

func ListenForWebSocketMessages(c *WebSocketConnetion, messageHandler OnWebSocketBinaryMessage) error {
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		switch messageType {
		case websocket.BinaryMessage:
			messageHandler(message)
		case websocket.CloseMessage:
			return nil
		}
	}
}
