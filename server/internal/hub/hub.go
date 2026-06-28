package hub

import (
	"encoding/json"
	"log"
	"time"

	"paintwar/server/internal/game"
	"paintwar/server/internal/mapgen"
	"paintwar/server/internal/model"
	"paintwar/server/internal/ws"
)

const commandBuffer = 256

// Persister records players and match results asynchronously. Calls must not
// block the hub goroutine. A nil Persister disables persistence.
type Persister interface {
	UpsertPlayer(username string)
	SaveMatch(r model.MatchResult)
}

// member is a connected client plus its room role and team.
type member struct {
	client *ws.Client
	role   model.Role
	team   model.Team
}

// Hub is the single owner of all room state. Exactly one goroutine (run) reads
// and mutates its fields; everything else communicates via the commands channel
// or the tick timer. It implements ws.Registrar.
type Hub struct {
	commands  chan command
	matchMs   int64
	persister Persister

	state      model.RoomState
	members    []*member          // ordered by join time; members[0] is the leader
	byUsername map[string]*member // active usernames -> member

	world   *game.World
	curMap  game.MapData
	endsAt  int64
	pending model.MatchResult // last finished match, awaiting persistence (M7)

	leaderboard []ws.LeaderEntry // cached; populated in milestone 7
}

// New creates a hub in the LOBBY state and starts its goroutine. persister may
// be nil to run without a database.
func New(matchMs int64, persister Persister) *Hub {
	h := &Hub{
		commands:   make(chan command, commandBuffer),
		matchMs:    matchMs,
		persister:  persister,
		state:      model.StateLobby,
		byUsername: make(map[string]*member),
	}
	go h.run()
	return h
}

// run is the single-owner loop: it serializes commands and tick processing.
func (h *Hub) run() {
	ticker := time.NewTicker(time.Second / game.TickHz)
	defer ticker.Stop()
	for {
		select {
		case cmd := <-h.commands:
			switch c := cmd.(type) {
			case registerCmd:
				c.reply <- h.handleRegister(c.client)
			case unregisterCmd:
				h.handleUnregister(c.client)
			case messageCmd:
				h.handleMessage(c.client, c.env)
			case leaderboardCmd:
				h.leaderboard = c.entries
				if h.state == model.StateLobby {
					h.broadcastLobbyState()
				}
			}
		case <-ticker.C:
			h.tick()
		}
	}
}

func nowMs() int64 { return time.Now().UnixMilli() }

// --- ws.Registrar implementation (called from connection goroutines) ---

func (h *Hub) Register(c *ws.Client) error {
	reply := make(chan error, 1)
	h.commands <- registerCmd{client: c, reply: reply}
	return <-reply
}

func (h *Hub) Unregister(c *ws.Client) { h.commands <- unregisterCmd{client: c} }

func (h *Hub) HandleMessage(c *ws.Client, env ws.Envelope) {
	h.commands <- messageCmd{client: c, env: env}
}

// UpdateLeaderboard is called by the persistence worker (off the hub goroutine)
// to publish a refreshed leaderboard. It maps domain rows to wire rows and hands
// off to the hub goroutine.
func (h *Hub) UpdateLeaderboard(entries []model.LeaderEntry) {
	wire := make([]ws.LeaderEntry, 0, len(entries))
	for _, e := range entries {
		wire = append(wire, ws.LeaderEntry{Username: e.Username, Wins: e.Wins, Losses: e.Losses})
	}
	h.commands <- leaderboardCmd{entries: wire}
}

// --- command handlers (hub goroutine only) ---

func (h *Hub) handleRegister(c *ws.Client) error {
	if _, exists := h.byUsername[c.Username]; exists {
		return ws.ErrUsernameTaken
	}

	playing := h.state == model.StatePlaying
	role := model.RoleLobbyPlayer
	if playing {
		role = model.RoleSpectator
	}

	m := &member{client: c, role: role}
	h.members = append(h.members, m)
	h.byUsername[c.Username] = m

	if h.persister != nil {
		h.persister.UpsertPlayer(c.Username)
	}

	c.SendMsg(ws.TypeJoined, ws.JoinedData{
		UserID: c.ID, Username: c.Username, Role: string(role), Team: string(m.team),
	})

	if h.leaderID() == c.ID {
		c.SendMsg(ws.TypeYouAreLeader, nil)
	}

	if playing {
		// Late joiner spectates the match in progress.
		c.SendMsg(ws.TypeYouAreSpectator, nil)
		c.SendMsg(ws.TypeMatchStart, matchStartData{Map: h.curMap, EndsAt: h.endsAt})
	} else {
		h.balanceTeams()
		h.broadcastLobbyState()
	}
	return nil
}

func (h *Hub) handleUnregister(c *ws.Client) {
	m, ok := h.byUsername[c.Username]
	if !ok || m.client != c {
		return
	}

	prevLeader := h.leaderID()
	delete(h.byUsername, c.Username)
	for i, mm := range h.members {
		if mm == m {
			h.members = append(h.members[:i], h.members[i+1:]...)
			break
		}
	}
	if h.world != nil {
		h.world.RemovePlayer(c.ID)
	}

	// Empty room: reset to a clean lobby so the next joiner starts fresh.
	if len(h.members) == 0 {
		h.state = model.StateLobby
		h.world = nil
		return
	}

	// Promote a new leader if needed.
	if newLeader := h.leaderID(); newLeader != "" && newLeader != prevLeader {
		h.members[0].client.SendMsg(ws.TypeYouAreLeader, nil)
	}

	if h.state == model.StateLobby {
		h.balanceTeams()
		h.broadcastLobbyState()
	}
}

func (h *Hub) handleMessage(c *ws.Client, env ws.Envelope) {
	switch env.Type {
	case ws.TypePing:
		c.SendMsg(ws.TypePong, env.Data)

	case ws.TypeStartGame:
		h.handleStartGame(c)

	case ws.TypeReturnToLobby:
		if h.state == model.StateScoreboard {
			h.backToLobby()
		}

	case ws.TypeInput:
		if h.state != model.StatePlaying || h.world == nil {
			return
		}
		m := h.byUsername[c.Username]
		if m == nil || m.role != model.RolePlayer {
			return
		}
		var in ws.InputData
		if json.Unmarshal(env.Data, &in) != nil {
			return
		}
		h.world.SetInput(c.ID, game.Input{
			DX: clampDir(in.DX), DY: clampDir(in.DY), Shoot: in.Shoot,
		})
	}
}

func (h *Hub) handleStartGame(c *ws.Client) {
	if h.leaderID() != c.ID {
		c.SendMsg(ws.TypeError, ws.ErrorData{Code: ws.ErrNotLeader, Message: "only the leader can start"})
		return
	}
	if h.state != model.StateLobby {
		return
	}
	if h.lobbyPlayerCount() < minPlayers {
		c.SendMsg(ws.TypeError, ws.ErrorData{Code: ws.ErrNotEnoughPlayers, Message: "need at least 2 players"})
		return
	}
	h.startMatch()
}

// --- match lifecycle ---

func (h *Hub) startMatch() {
	h.balanceTeams() // freeze balanced teams for the match

	specs := make([]game.PlayerSpec, 0, len(h.members))
	for _, m := range h.members {
		if m.role == model.RoleSpectator {
			continue
		}
		m.role = model.RolePlayer
		specs = append(specs, game.PlayerSpec{ID: m.client.ID, Team: m.team})
	}

	seed := time.Now().UnixNano()
	h.curMap = mapgen.Generate(seed)
	start := nowMs()
	h.endsAt = start + h.matchMs
	h.world = game.NewWorld(h.curMap, specs, start, h.matchMs)
	h.state = model.StatePlaying

	h.broadcastMsg(ws.TypeMatchStart, matchStartData{Map: h.curMap, EndsAt: h.endsAt})
}

// finishMatch transitions PLAYING -> SCOREBOARD: it tallies the result, freezes
// it for persistence (M7), and broadcasts the scoreboard. Roles stay frozen
// until return_to_lobby so spectators are only re-pooled at the lobby.
func (h *Hub) finishMatch() {
	h.pending = h.buildMatchResult()
	if h.persister != nil {
		h.persister.SaveMatch(h.pending)
	}
	h.state = model.StateScoreboard
	h.world = nil
	h.broadcastMsg(ws.TypeScoreboard, scoreboardFrom(h.pending))
}

// backToLobby transitions SCOREBOARD -> LOBBY, re-pooling spectators onto teams.
func (h *Hub) backToLobby() {
	h.state = model.StateLobby
	h.world = nil
	for _, m := range h.members {
		m.role = model.RoleLobbyPlayer
	}
	h.balanceTeams()
	h.broadcastLobbyState()
}

func (h *Hub) tick() {
	if h.state != model.StatePlaying || h.world == nil {
		return
	}
	now := nowMs()
	h.world.Step(now)
	if h.world.Ended(now) {
		h.finishMatch()
		return
	}
	h.broadcastMsg(ws.TypeState, h.world.Snapshot(now))
}

// --- helpers ---

func (h *Hub) leaderID() string {
	if len(h.members) == 0 {
		return ""
	}
	return h.members[0].client.ID
}

func (h *Hub) broadcastLobbyState() {
	players := make([]ws.LobbyPlayer, 0, len(h.members))
	for _, m := range h.members {
		players = append(players, ws.LobbyPlayer{
			ID: m.client.ID, Username: m.client.Username,
			Team: string(m.team), Role: string(m.role),
		})
	}
	h.broadcastMsg(ws.TypeLobbyState, ws.LobbyStateData{
		LeaderID: h.leaderID(), Players: players, Leaderboard: h.leaderboard,
	})
}

// broadcastMsg encodes once and sends to every member.
func (h *Hub) broadcastMsg(msgType string, data any) {
	b, err := ws.Encode(msgType, data)
	if err != nil {
		log.Printf("%s encode: %v", msgType, err)
		return
	}
	for _, m := range h.members {
		m.client.Send(b)
	}
}

// matchStartData is the match_start payload (static map + match end time).
type matchStartData struct {
	Map    game.MapData `json:"map"`
	EndsAt int64        `json:"endsAt"`
}

// clampDir constrains a direction component to {-1, 0, 1}.
func clampDir(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}
