package hub

import (
	"log"

	"paintwar/server/internal/model"
	"paintwar/server/internal/ws"
)

const commandBuffer = 256

// member is a connected client plus its room role and team.
type member struct {
	client *ws.Client
	role   model.Role
	team   model.Team
}

// Hub is the single owner of all room state. Exactly one goroutine (Run) reads
// and mutates its fields; everything else communicates via the commands channel.
// It implements ws.Registrar.
type Hub struct {
	commands chan command

	state      model.RoomState
	members    []*member          // ordered by join time; members[0] is the leader
	byUsername map[string]*member // active usernames -> member

	leaderboard []ws.LeaderEntry // cached; populated in milestone 7
}

// New creates a hub in the LOBBY state and starts its goroutine.
func New() *Hub {
	h := &Hub{
		commands:   make(chan command, commandBuffer),
		state:      model.StateLobby,
		byUsername: make(map[string]*member),
	}
	go h.run()
	return h
}

// run is the single-owner command loop.
func (h *Hub) run() {
	for cmd := range h.commands {
		switch c := cmd.(type) {
		case registerCmd:
			c.reply <- h.handleRegister(c.client)
		case unregisterCmd:
			h.handleUnregister(c.client)
		case messageCmd:
			h.handleMessage(c.client, c.env)
		}
	}
}

// --- ws.Registrar implementation (called from connection goroutines) ---

// Register blocks until the hub goroutine decides whether to admit the client.
func (h *Hub) Register(c *ws.Client) error {
	reply := make(chan error, 1)
	h.commands <- registerCmd{client: c, reply: reply}
	return <-reply
}

// Unregister enqueues removal; it returns immediately.
func (h *Hub) Unregister(c *ws.Client) {
	h.commands <- unregisterCmd{client: c}
}

// HandleMessage enqueues a decoded message; it returns immediately.
func (h *Hub) HandleMessage(c *ws.Client, env ws.Envelope) {
	h.commands <- messageCmd{client: c, env: env}
}

// --- command handlers (run on the hub goroutine only) ---

func (h *Hub) handleRegister(c *ws.Client) error {
	if _, exists := h.byUsername[c.Username]; exists {
		return ws.ErrUsernameTaken
	}

	role := model.RoleLobbyPlayer
	if h.state == model.StatePlaying {
		role = model.RoleSpectator
	}

	m := &member{client: c, role: role}
	h.members = append(h.members, m)
	h.byUsername[c.Username] = m

	if role == model.RoleSpectator {
		c.SendMsg(ws.TypeYouAreSpectator, nil)
	} else {
		h.balanceTeams()
	}

	c.SendMsg(ws.TypeJoined, ws.JoinedData{
		UserID:   c.ID,
		Username: c.Username,
		Role:     string(m.role),
		Team:     string(m.team),
	})

	if h.leaderID() == c.ID {
		c.SendMsg(ws.TypeYouAreLeader, nil)
	}

	h.broadcastLobbyState()
	return nil
}

func (h *Hub) handleUnregister(c *ws.Client) {
	m, ok := h.byUsername[c.Username]
	if !ok || m.client != c {
		return // already replaced or never registered
	}

	prevLeader := h.leaderID()
	delete(h.byUsername, c.Username)
	for i, mm := range h.members {
		if mm == m {
			h.members = append(h.members[:i], h.members[i+1:]...)
			break
		}
	}

	h.balanceTeams()

	// Notify a newly promoted leader.
	if newLeader := h.leaderID(); newLeader != "" && newLeader != prevLeader {
		if lm := h.members[0]; lm.role != model.RoleSpectator {
			lm.client.SendMsg(ws.TypeYouAreLeader, nil)
		}
	}

	h.broadcastLobbyState()
}

func (h *Hub) handleMessage(c *ws.Client, env ws.Envelope) {
	switch env.Type {
	case ws.TypePing:
		c.SendMsg(ws.TypePong, env.Data)
	case ws.TypeStartGame:
		// Validated and wired in milestone 5.
		if h.leaderID() != c.ID {
			c.SendMsg(ws.TypeError, ws.ErrorData{Code: ws.ErrNotLeader, Message: "only the leader can start"})
		}
	}
}

// leaderID returns the id of the current leader (first member), or "".
func (h *Hub) leaderID() string {
	if len(h.members) == 0 {
		return ""
	}
	return h.members[0].client.ID
}

// broadcastLobbyState sends the current roster + leader + leaderboard to all.
func (h *Hub) broadcastLobbyState() {
	players := make([]ws.LobbyPlayer, 0, len(h.members))
	for _, m := range h.members {
		players = append(players, ws.LobbyPlayer{
			ID:       m.client.ID,
			Username: m.client.Username,
			Team:     string(m.team),
			Role:     string(m.role),
		})
	}
	data := ws.LobbyStateData{
		LeaderID:    h.leaderID(),
		Players:     players,
		Leaderboard: h.leaderboard,
	}
	b, err := ws.Encode(ws.TypeLobbyState, data)
	if err != nil {
		log.Printf("lobby_state encode: %v", err)
		return
	}
	for _, m := range h.members {
		m.client.Send(b)
	}
}
