package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"paintwar/server/internal/model"
	"paintwar/server/internal/ws"
)

func TestBalanceTeamsAlternates(t *testing.T) {
	h := &Hub{}
	for i := 0; i < 5; i++ {
		h.members = append(h.members, &member{
			client: &ws.Client{ID: fmt.Sprintf("u%d", i)},
			role:   model.RoleLobbyPlayer,
		})
	}
	h.balanceTeams()

	var red, green int
	for _, m := range h.members {
		switch m.team {
		case model.TeamRed:
			red++
		case model.TeamGreen:
			green++
		default:
			t.Fatalf("member %s has no team", m.client.ID)
		}
	}
	if diff := red - green; diff < -1 || diff > 1 {
		t.Errorf("unbalanced teams: red=%d green=%d", red, green)
	}
}

// --- integration helpers ---

func startHub(t *testing.T) string {
	t.Helper()
	h := New(120000)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.NewHandler(h))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

func join(t *testing.T, c *websocket.Conn, username string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	b, _ := ws.Encode(ws.TypeJoin, ws.JoinData{Username: username})
	if err := c.Write(ctx, websocket.MessageText, b); err != nil {
		t.Fatalf("write join: %v", err)
	}
}

// readUntil reads frames until one matches msgType, failing on timeout.
func readUntil(t *testing.T, c *websocket.Conn, msgType string) ws.Envelope {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		_, data, err := c.Read(ctx)
		cancel()
		if err != nil {
			t.Fatalf("waiting for %q: %v", msgType, err)
		}
		var env ws.Envelope
		if json.Unmarshal(data, &env) != nil {
			continue
		}
		if env.Type == msgType {
			return env
		}
	}
}

func lobbyState(t *testing.T, c *websocket.Conn, wantPlayers int) ws.LobbyStateData {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		env := readUntil(t, c, ws.TypeLobbyState)
		var ls ws.LobbyStateData
		_ = json.Unmarshal(env.Data, &ls)
		if len(ls.Players) == wantPlayers {
			return ls
		}
	}
	t.Fatalf("never saw lobby_state with %d players", wantPlayers)
	return ws.LobbyStateData{}
}

func TestLeaderElectionAndHandoff(t *testing.T) {
	url := startHub(t)

	// alice joins first -> becomes leader.
	alice := dial(t, url)
	join(t, alice, "alice")
	if env := readUntil(t, alice, ws.TypeYouAreLeader); env.Type != ws.TypeYouAreLeader {
		t.Fatal("alice should be leader")
	}
	ls := lobbyState(t, alice, 1)
	aliceID := ls.Players[0].ID
	if ls.LeaderID != aliceID {
		t.Fatalf("expected leader %s, got %s", aliceID, ls.LeaderID)
	}

	// bob joins -> two balanced players, alice still leader.
	bob := dial(t, url)
	join(t, bob, "bob")
	ls = lobbyState(t, bob, 2)
	if ls.LeaderID != aliceID {
		t.Errorf("alice should remain leader, got %s", ls.LeaderID)
	}
	if ls.Players[0].Team == ls.Players[1].Team {
		t.Errorf("teams should differ: %s vs %s", ls.Players[0].Team, ls.Players[1].Team)
	}
	bobID := ls.Players[1].ID

	// alice (leader) disconnects -> bob is promoted.
	_ = alice.Close(websocket.StatusNormalClosure, "bye")
	if env := readUntil(t, bob, ws.TypeYouAreLeader); env.Type != ws.TypeYouAreLeader {
		t.Fatal("bob should be promoted to leader")
	}
	ls = lobbyState(t, bob, 1)
	if ls.LeaderID != bobID {
		t.Errorf("expected bob %s as leader, got %s", bobID, ls.LeaderID)
	}
}
