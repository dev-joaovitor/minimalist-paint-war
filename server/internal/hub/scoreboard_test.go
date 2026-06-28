package hub

import (
	"encoding/json"
	"strings"
	"testing"

	"paintwar/server/internal/ws"
)

func TestScoreboardAndReturnToLobby(t *testing.T) {
	url := startHubMs(t, 200) // short match so it times out quickly

	alice := dial(t, url)
	join(t, alice, "alice")
	_ = joinedID(t, alice)

	bob := dial(t, url)
	join(t, bob, "bob")
	_ = joinedID(t, bob)

	sendType(t, alice, ws.TypeStartGame, nil)

	// Match times out -> scoreboard.
	env := readUntil(t, alice, ws.TypeScoreboard)
	var sb scoreboardData
	_ = json.Unmarshal(env.Data, &sb)
	if !strings.Contains(sb.ScoreText, " x ") {
		t.Errorf("unexpected scoreText %q", sb.ScoreText)
	}
	if len(sb.PerPlayer) != 2 {
		t.Errorf("expected 2 player results, got %d", len(sb.PerPlayer))
	}
	if sb.Winner == "" {
		t.Error("scoreboard winner should be set (RED/GREEN/DRAW)")
	}

	// Returning to lobby re-pools players onto balanced teams.
	sendType(t, alice, ws.TypeReturnToLobby, nil)
	ls := lobbyState(t, alice, 2)
	for _, p := range ls.Players {
		if p.Role != "LOBBY_PLAYER" {
			t.Errorf("expected LOBBY_PLAYER after return, got %s", p.Role)
		}
		if p.Team != "RED" && p.Team != "GREEN" {
			t.Errorf("expected a team after return, got %q", p.Team)
		}
	}
}
