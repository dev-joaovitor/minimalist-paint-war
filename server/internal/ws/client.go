package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

const (
	sendBuffer   = 32
	writeTimeout = 5 * time.Second
)

var clientSeq atomic.Int64

// Client is a single WebSocket connection. It owns one read pump and one write
// pump; outbound messages are queued on a buffered channel so callers (e.g. the
// hub) never block on the socket.
type Client struct {
	ID       string
	Username string

	conn   *websocket.Conn
	send   chan []byte
	ctx    context.Context
	cancel context.CancelFunc
	closed sync.Once
}

func newClient(parent context.Context, conn *websocket.Conn, username string) *Client {
	ctx, cancel := context.WithCancel(parent)
	return &Client{
		ID:       "u" + strconv.FormatInt(clientSeq.Add(1), 10),
		Username: username,
		conn:     conn,
		send:     make(chan []byte, sendBuffer),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Send queues a pre-encoded message. It never blocks: if the buffer is full the
// consumer is too slow and the connection is dropped.
func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	case <-c.ctx.Done():
	default:
		c.close()
	}
}

// SendMsg encodes a typed payload and queues it.
func (c *Client) SendMsg(msgType string, data any) {
	b, err := Encode(msgType, data)
	if err != nil {
		return
	}
	c.Send(b)
}

func (c *Client) close() {
	c.closed.Do(func() {
		c.cancel()
		_ = c.conn.CloseNow()
	})
}

// readPump reads frames until the connection closes, dispatching each to reg.
func (c *Client) readPump(reg Registrar) {
	defer c.close()
	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			return
		}
		var env Envelope
		if json.Unmarshal(data, &env) != nil {
			continue
		}
		reg.HandleMessage(c, env)
	}
}

// writePump drains the send channel to the socket.
func (c *Client) writePump() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.send:
			wctx, cancel := context.WithTimeout(c.ctx, writeTimeout)
			err := c.conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				c.close()
				return
			}
		}
	}
}
