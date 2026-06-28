package hub

import (
	"fmt"

	"paintwar/server/internal/model"
)

// scorePair carries the two team flag totals.
type scorePair struct {
	Red   int `json:"RED"`
	Green int `json:"GREEN"`
}

// playerResult is one scoreboard row.
type playerResult struct {
	Username string `json:"username"`
	Team     string `json:"team"`
	Result   string `json:"result"`
}

// scoreboardData is the scoreboard message payload.
type scoreboardData struct {
	Score     scorePair      `json:"score"`
	Winner    string         `json:"winner"`
	ScoreText string         `json:"scoreText"`
	PerPlayer []playerResult `json:"perPlayer"`
}

// buildMatchResult tallies the finished match from the world and current roster.
// It must run before roles are reset back to lobby players.
func (h *Hub) buildMatchResult() model.MatchResult {
	red, green := h.world.FlagCount()
	winner := h.world.Winner()

	res := model.MatchResult{
		Seed:       h.world.Seed,
		Red:        red,
		Green:      green,
		Winner:     winner,
		DurationMs: h.matchMs,
	}
	for _, m := range h.members {
		if m.role != model.RolePlayer {
			continue
		}
		res.Players = append(res.Players, model.PlayerResult{
			Username: m.client.Username,
			Team:     m.team,
			Result:   outcome(m.team, winner),
		})
	}
	return res
}

// scoreboardFrom converts a MatchResult into the broadcast payload.
func scoreboardFrom(r model.MatchResult) scoreboardData {
	winner := string(r.Winner)
	if r.Winner == model.TeamNone {
		winner = "DRAW"
	}
	sb := scoreboardData{
		Score:     scorePair{Red: r.Red, Green: r.Green},
		Winner:    winner,
		ScoreText: fmt.Sprintf("Red %d x %d Green", r.Red, r.Green),
	}
	for _, p := range r.Players {
		sb.PerPlayer = append(sb.PerPlayer, playerResult{
			Username: p.Username, Team: string(p.Team), Result: p.Result,
		})
	}
	return sb
}

// outcome maps a team + winner to "win" / "loss" / "draw" (relative to winner).
func outcome(team, winner model.Team) string {
	switch {
	case winner == model.TeamNone:
		return "draw"
	case team == winner:
		return "win"
	default:
		return "loss"
	}
}
