package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestValidUsername(t *testing.T) {
	ok := []string{"abc", "johndoe", "abcdefghijklmnop"}
	bad := []string{"ab", "John", "john doe", "john1", "john_doe", "", "abcdefghijklmnopq", "joão"}
	for _, s := range ok {
		if !validUsername(s) {
			t.Errorf("expected %q valid", s)
		}
	}
	for _, s := range bad {
		if validUsername(s) {
			t.Errorf("expected %q invalid", s)
		}
	}
}

// dialAndJoin connects, sends a join with username, and returns the first reply.
func dialAndJoin(t *testing.T, url, username string) Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(url, "http")
	c, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })

	join, _ := encode(TypeJoin, JoinData{Username: username})
	if err := c.Write(ctx, websocket.MessageText, join); err != nil {
		t.Fatalf("write join: %v", err)
	}

	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal reply: %v", err)
	}
	return env
}

func newTestServer() *httptest.Server {
	reg := NewMemRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", NewHandler(reg))
	return httptest.NewServer(mux)
}

func TestJoinSuccess(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	env := dialAndJoin(t, srv.URL, "johndoe")
	if env.Type != TypeJoined {
		t.Fatalf("expected joined, got %q", env.Type)
	}
	var jd JoinedData
	_ = json.Unmarshal(env.Data, &jd)
	if jd.Username != "johndoe" {
		t.Errorf("expected username johndoe, got %q", jd.Username)
	}
}

func TestJoinInvalidUsername(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	env := dialAndJoin(t, srv.URL, "John1")
	if env.Type != TypeError {
		t.Fatalf("expected error, got %q", env.Type)
	}
	var ed ErrorData
	_ = json.Unmarshal(env.Data, &ed)
	if ed.Code != ErrInvalidUsername {
		t.Errorf("expected %s, got %s", ErrInvalidUsername, ed.Code)
	}
}

func TestJoinDuplicateRejected(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	// First connection holds the username.
	if env := dialAndJoin(t, srv.URL, "johndoe"); env.Type != TypeJoined {
		t.Fatalf("first join failed: %q", env.Type)
	}

	// Second connection with the same username must be rejected.
	env := dialAndJoin(t, srv.URL, "johndoe")
	if env.Type != TypeError {
		t.Fatalf("expected error, got %q", env.Type)
	}
	var ed ErrorData
	_ = json.Unmarshal(env.Data, &ed)
	if ed.Code != ErrUsernameInUse {
		t.Errorf("expected %s, got %s", ErrUsernameInUse, ed.Code)
	}
}
