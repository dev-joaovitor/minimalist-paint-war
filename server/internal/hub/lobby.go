package hub

import "paintwar/server/internal/model"

// balanceTeams reassigns lobby players to Red/Green alternately by join order,
// guaranteeing |Red - Green| <= 1. Spectators keep no team. Called on any roster
// change while in the lobby. Players cannot choose their team.
func (h *Hub) balanceTeams() {
	next := model.TeamRed
	for _, m := range h.members {
		if m.role == model.RoleSpectator {
			m.team = model.TeamNone
			continue
		}
		m.team = next
		if next == model.TeamRed {
			next = model.TeamGreen
		} else {
			next = model.TeamRed
		}
	}
}
