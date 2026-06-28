// Package model holds small domain types shared across hub, game, and ws so
// those packages can reference teams/roles/states without import cycles.
package model

// Team identifies a side. Empty means unassigned (e.g. spectators).
type Team string

const (
	TeamNone  Team = ""
	TeamRed   Team = "RED"
	TeamGreen Team = "GREEN"
)

// Role is a member's participation level in the room.
type Role string

const (
	RoleLobbyPlayer Role = "LOBBY_PLAYER"
	RolePlayer      Role = "PLAYER"
	RoleSpectator   Role = "SPECTATOR"
)

// PlayerResult is one player's outcome in a finished match.
type PlayerResult struct {
	Username string
	Team     Team
	Result   string // "win", "loss", or "draw"
}

// MatchResult summarizes a finished match for the scoreboard and persistence.
type MatchResult struct {
	Seed       int64
	Red        int
	Green      int
	Winner     Team // TeamNone means a draw
	DurationMs int64
	Players    []PlayerResult
}

// RoomState is the room's lifecycle phase.
type RoomState string

const (
	StateLobby      RoomState = "LOBBY"
	StatePlaying    RoomState = "PLAYING"
	StateScoreboard RoomState = "SCOREBOARD"
)
