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
	ErrInvalidUsername = "INVALID_USERNAME"
	ErrUsernameInUse   = "USERNAME_IN_USE"
	ErrNotLeader       = "NOT_LEADER"
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

// encode marshals a typed payload into an Envelope's JSON bytes.
func encode(msgType string, data any) ([]byte, error) {
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
