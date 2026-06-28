package ws

import "encoding/json"

// Envelope is the wire format for every message in both directions.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Client -> Server message types.
const (
	TypeJoin          = "join"
	TypeInput         = "input"
	TypeStartGame     = "start_game"
	TypeReturnToLobby = "return_to_lobby"
	TypePing          = "ping"
)

// Server -> Client message types.
const (
	TypeJoined          = "joined"
	TypeError           = "error"
	TypeYouAreLeader    = "you_are_leader"
	TypeYouAreSpectator = "you_are_spectator"
	TypeLobbyState      = "lobby_state"
	TypeMatchStart      = "match_start"
	TypeState           = "state"
	TypeScoreboard      = "scoreboard"
	TypePong            = "pong"
)

// Error codes.
const (
	ErrInvalidUsername  = "INVALID_USERNAME"
	ErrUsernameInUse    = "USERNAME_IN_USE"
	ErrNotLeader        = "NOT_LEADER"
	ErrNotEnoughPlayers = "NOT_ENOUGH_PLAYERS"
)

// JoinData is the payload of a join message.
type JoinData struct {
	Username string `json:"username"`
}

// InputData is the payload of an input message. dx/dy are in {-1, 0, 1}.
type InputData struct {
	DX    int  `json:"dx"`
	DY    int  `json:"dy"`
	Shoot bool `json:"shoot"`
}

// PingData carries a client timestamp echoed back in pong.
type PingData struct {
	T int64 `json:"t"`
}

// JoinedData acknowledges a successful join.
type JoinedData struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Team     string `json:"team"`
}

// ErrorData describes a rejected request.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// LobbyPlayer is one roster entry in lobby_state.
type LobbyPlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Team     string `json:"team"`
	Role     string `json:"role"`
}

// LeaderEntry is one leaderboard row.
type LeaderEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
}

// LobbyStateData is broadcast whenever the lobby roster or leader changes.
// Clients identify themselves by matching their own id (from the joined ack).
type LobbyStateData struct {
	LeaderID    string        `json:"leaderId"`
	Players     []LobbyPlayer `json:"players"`
	Leaderboard []LeaderEntry `json:"leaderboard"`
}

// Encode marshals a typed payload into an Envelope's JSON bytes.
func Encode(msgType string, data any) ([]byte, error) {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	return json.Marshal(Envelope{Type: msgType, Data: raw})
}
