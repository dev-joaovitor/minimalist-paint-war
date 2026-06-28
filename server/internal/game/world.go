package game

import "paintwar/server/internal/model"

// Input is a player's latest intent, applied once per tick (last-write-wins).
type Input struct {
	DX    int
	DY    int
	Shoot bool
}

// Player is a controllable avatar.
type Player struct {
	ID   string
	Team model.Team
	X    float64
	Y    float64
	Face int
	HP   int
	Dead bool

	RespawnAt  int64 // ms; valid while Dead
	lastShotMs int64
	input      Input
}

// Projectile is a straight-line shot owned by a team.
type Projectile struct {
	ID      int
	X, Y    float64
	VX, VY  float64
	Team    model.Team
	OwnerID string
}

// Flag is a static capture point; Team is its current owner (TeamNone = neutral).
type Flag struct {
	ID   int
	X, Y float64
	Team model.Team
}

// MapData is the static geometry produced by mapgen and frozen for a match.
type MapData struct {
	Seed      int64   `json:"seed"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Obstacles []Rect  `json:"obstacles"`
	Flags     []Flag  `json:"flags"`
	RedBase   Rect    `json:"redBase"`
	GreenBase Rect    `json:"greenBase"`
}

// PlayerSpec describes a player to place when a match starts.
type PlayerSpec struct {
	ID   string
	Team model.Team
}

// World is the full mutable match state. It is owned by a single goroutine; it
// performs no locking of its own.
type World struct {
	W, H      float64
	Obstacles []Rect
	Flags     []*Flag
	RedBase   Rect
	GreenBase Rect

	Players     map[string]*Player
	order       []string // stable player iteration order
	Projectiles []*Projectile

	Tick       int64
	startMs    int64
	endsAtMs   int64
	nextProjID int
	Seed       int64
}

// NewWorld builds a world from map data and the players to spawn.
func NewWorld(md MapData, specs []PlayerSpec, startMs, matchMs int64) *World {
	w := &World{
		W:         md.Width,
		H:         md.Height,
		Obstacles: md.Obstacles,
		RedBase:   md.RedBase,
		GreenBase: md.GreenBase,
		Players:   make(map[string]*Player, len(specs)),
		startMs:   startMs,
		endsAtMs:  startMs + matchMs,
		Seed:      md.Seed,
	}
	for i := range md.Flags {
		f := md.Flags[i]
		w.Flags = append(w.Flags, &f)
	}
	for _, sp := range specs {
		p := &Player{ID: sp.ID, Team: sp.Team, HP: MaxHP}
		w.spawn(p)
		w.Players[sp.ID] = p
		w.order = append(w.order, sp.ID)
	}
	return w
}

// SetInput records a player's latest input for the next tick.
func (w *World) SetInput(id string, in Input) {
	if p, ok := w.Players[id]; ok {
		p.input = in
	}
}

// baseFor returns the spawn rectangle for a team.
func (w *World) baseFor(t model.Team) Rect {
	if t == model.TeamGreen {
		return w.GreenBase
	}
	return w.RedBase
}

// spawn places a player at its team base center and resets vitals.
func (w *World) spawn(p *Player) {
	cx, cy := w.baseFor(p.Team).Center()
	p.X, p.Y = cx, cy
	p.HP = MaxHP
	p.Dead = false
	p.input = Input{}
}

// TimeLeftMs returns remaining match time, clamped at zero.
func (w *World) TimeLeftMs(nowMs int64) int64 {
	if r := w.endsAtMs - nowMs; r > 0 {
		return r
	}
	return 0
}

// Ended reports whether the match timer has expired.
func (w *World) Ended(nowMs int64) bool {
	return nowMs >= w.endsAtMs
}

// FlagCount tallies flags by team.
func (w *World) FlagCount() (red, green int) {
	for _, f := range w.Flags {
		switch f.Team {
		case model.TeamRed:
			red++
		case model.TeamGreen:
			green++
		}
	}
	return
}

// Winner returns the team holding the flag majority, or TeamNone on a tie.
func (w *World) Winner() model.Team {
	red, green := w.FlagCount()
	switch {
	case red > green:
		return model.TeamRed
	case green > red:
		return model.TeamGreen
	default:
		return model.TeamNone
	}
}
