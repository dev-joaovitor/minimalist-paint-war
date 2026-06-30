package hub

import "paintwar/server/internal/model"

// minPlayers is the minimum number of lobby players required to start a match.
const minPlayers = 2

// lobbyPlayerCount returns how many members are eligible to play (not spectators).
func (h *Hub) lobbyPlayerCount() int {
	n := 0
	for _, m := range h.members {
		if m.role != model.RoleSpectator {
			n++
		}
	}
	return n
}

// activePlayerCount returns how many members are playing the current match.
func (h *Hub) activePlayerCount() int {
	n := 0
	for _, m := range h.members {
		if m.role == model.RolePlayer {
			n++
		}
	}
	return n
}
