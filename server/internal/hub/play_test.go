package hub

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coder/websocket"
	"paintwar/server/internal/game"
	"paintwar/server/internal/ws"
)

func sendType(t *testing.T, c *websocket.Conn, msgType string, data any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	b, _ := ws.Encode(msgType, data)
	if err := c.Write(ctx, websocket.MessageText, b); err != nil {
		t.Fatalf("write %s: %v", msgType, err)
	}
}

func joinedID(t *testing.T, c *websocket.Conn) string {
	t.Helper()
	env := readUntil(t, c, ws.TypeJoined)
	var jd ws.JoinedData
	_ = json.Unmarshal(env.Data, &jd)
	return jd.UserID
}

type matchStart struct {
	Map    game.MapData `json:"map"`
	EndsAt int64        `json:"endsAt"`
}

func TestStartGameAuthorization(t *testing.T) {
	url := startHub(t)

	alice := dial(t, url)
	join(t, alice, "alice")
	_ = joinedID(t, alice)

	// Single player: not enough to start.
	sendType(t, alice, ws.TypeStartGame, nil)
	if env := readUntil(t, alice, ws.TypeError); errCode(env) != ws.ErrNotEnoughPlayers {
		t.Fatalf("expected NOT_ENOUGH_PLAYERS, got %s", errCode(env))
	}

	// Second player joins; non-leader cannot start.
	bob := dial(t, url)
	join(t, bob, "bob")
	_ = joinedID(t, bob)
	sendType(t, bob, ws.TypeStartGame, nil)
	if env := readUntil(t, bob, ws.TypeError); errCode(env) != ws.ErrNotLeader {
		t.Fatalf("expected NOT_LEADER, got %s", errCode(env))
	}
}

func errCode(env ws.Envelope) string {
	var ed ws.ErrorData
	_ = json.Unmarshal(env.Data, &ed)
	return ed.Code
}

func TestMatchStartStateAndInput(t *testing.T) {
	url := startHub(t)

	alice := dial(t, url)
	join(t, alice, "alice")
	aliceID := joinedID(t, alice)

	bob := dial(t, url)
	join(t, bob, "bob")
	_ = joinedID(t, bob)

	// Leader starts the match.
	sendType(t, alice, ws.TypeStartGame, nil)

	env := readUntil(t, alice, ws.TypeMatchStart)
	var ms matchStart
	_ = json.Unmarshal(env.Data, &ms)
	if len(ms.Map.Flags) != game.NumFlags {
		t.Fatalf("expected %d flags in match_start, got %d", game.NumFlags, len(ms.Map.Flags))
	}
	if ms.EndsAt == 0 {
		t.Error("match_start should carry endsAt")
	}

	// First state snapshot has both players.
	x0 := alicePos(t, alice, aliceID)

	// Move right; position should change.
	sendType(t, alice, ws.TypeInput, ws.InputData{DX: 1})
	var moved bool
	for i := 0; i < 6; i++ {
		if alicePos(t, alice, aliceID) > x0+5 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatalf("alice should have moved right from x=%.1f", x0)
	}
}

// alicePos reads the next state snapshot and returns the given player's X.
func alicePos(t *testing.T, c *websocket.Conn, id string) float64 {
	t.Helper()
	env := readUntil(t, c, ws.TypeState)
	var snap game.Snapshot
	_ = json.Unmarshal(env.Data, &snap)
	for _, p := range snap.Players {
		if p.ID == id {
			return p.X
		}
	}
	t.Fatalf("player %s not in snapshot", id)
	return 0
}
