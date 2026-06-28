package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/coder/websocket"
)

// ErrUsernameTaken is returned by a Registrar when a username is already active.
var ErrUsernameTaken = errors.New("username already connected")

// Registrar owns client lifecycle and inbound message handling. The hub
// implements this; a WebSocket connection knows nothing about game rules.
type Registrar interface {
	// Register admits a client or returns ErrUsernameTaken. On success the
	// registrar is responsible for any welcome messages (joined, lobby_state).
	Register(c *Client) error
	// Unregister removes a client after its connection closes.
	Unregister(c *Client)
	// HandleMessage processes an inbound, decoded message.
	HandleMessage(c *Client, env Envelope)
}

var usernameRe = regexp.MustCompile(`^[a-z]{3,16}$`)

// validUsername enforces lowercase letters only, length 3-16.
func validUsername(s string) bool {
	return usernameRe.MatchString(s)
}

const joinTimeout = 10 * time.Second

// NewHandler returns an http.HandlerFunc that upgrades to WebSocket, performs
// the join handshake (username validation + registration), then runs the pumps.
func NewHandler(reg Registrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			return
		}

		parent := context.Background()

		// First message must be a join within the handshake timeout.
		joinCtx, cancelJoin := context.WithTimeout(parent, joinTimeout)
		_, data, err := conn.Read(joinCtx)
		cancelJoin()
		if err != nil {
			_ = conn.CloseNow()
			return
		}

		var env Envelope
		if json.Unmarshal(data, &env) != nil || env.Type != TypeJoin {
			writeSync(parent, conn, TypeError, ErrorData{ErrInvalidUsername, "expected join message"})
			_ = conn.Close(websocket.StatusPolicyViolation, "bad handshake")
			return
		}

		var jd JoinData
		_ = json.Unmarshal(env.Data, &jd)
		if !validUsername(jd.Username) {
			writeSync(parent, conn, TypeError, ErrorData{ErrInvalidUsername, "username must be 3-16 lowercase letters"})
			_ = conn.Close(websocket.StatusPolicyViolation, "invalid username")
			return
		}

		client := newClient(parent, conn, jd.Username)
		if err := reg.Register(client); err != nil {
			writeSync(parent, conn, TypeError, ErrorData{ErrUsernameInUse, "username is already connected"})
			_ = conn.Close(websocket.StatusPolicyViolation, "username in use")
			return
		}

		go client.writePump()
		client.readPump(reg) // blocks until the connection closes
		reg.Unregister(client)
	}
}

// writeSync writes a single message synchronously (used before pumps start).
func writeSync(ctx context.Context, conn *websocket.Conn, msgType string, data any) {
	b, err := encode(msgType, data)
	if err != nil {
		return
	}
	wctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	_ = conn.Write(wctx, websocket.MessageText, b)
}
