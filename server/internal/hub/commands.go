package hub

import "paintwar/server/internal/ws"

// command is anything the hub goroutine processes. All state mutations flow
// through these so the hub owns its state without mutexes.
type command interface{ isCommand() }

// registerCmd admits a new client. reply receives nil on success or
// ws.ErrUsernameTaken if the username is already connected.
type registerCmd struct {
	client *ws.Client
	reply  chan error
}

// unregisterCmd removes a client after its connection closes.
type unregisterCmd struct {
	client *ws.Client
}

// messageCmd carries a decoded inbound message from a client.
type messageCmd struct {
	client *ws.Client
	env    ws.Envelope
}

// leaderboardCmd updates the cached leaderboard from the persistence worker.
type leaderboardCmd struct {
	entries []ws.LeaderEntry
}

func (registerCmd) isCommand()    {}
func (unregisterCmd) isCommand()  {}
func (messageCmd) isCommand()     {}
func (leaderboardCmd) isCommand() {}
