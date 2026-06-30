package game

// PlayerState is a player's per-tick wire representation.
type PlayerState struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Face   int     `json:"face"`
	HP     int     `json:"hp"`
	Team   string  `json:"team"`
	Dead   bool    `json:"dead"`
	Invuln bool    `json:"invuln"`
}

// ProjectileState is a projectile's per-tick wire representation.
type ProjectileState struct {
	ID   int     `json:"id"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Team string  `json:"team"`
}

// FlagState carries only the dynamic flag ownership (positions ship in MapData).
type FlagState struct {
	ID   int    `json:"id"`
	Team string `json:"team"`
}

// Snapshot is the full 20Hz world state sent to every client during a match.
type Snapshot struct {
	Tick        int64             `json:"tick"`
	TimeLeftMs  int64             `json:"timeLeftMs"`
	FlagCount   map[string]int    `json:"flagCount"`
	Players     []PlayerState     `json:"players"`
	Projectiles []ProjectileState `json:"projectiles"`
	Flags       []FlagState       `json:"flags"`
}

// Snapshot builds the wire state for the current tick.
func (w *World) Snapshot(nowMs int64) Snapshot {
	red, green := w.FlagCount()
	snap := Snapshot{
		Tick:        w.Tick,
		TimeLeftMs:  w.TimeLeftMs(nowMs),
		FlagCount:   map[string]int{"RED": red, "GREEN": green},
		Players:     make([]PlayerState, 0, len(w.order)),
		Projectiles: make([]ProjectileState, 0, len(w.Projectiles)),
		Flags:       make([]FlagState, 0, len(w.Flags)),
	}
	for _, id := range w.order {
		p := w.Players[id]
		snap.Players = append(snap.Players, PlayerState{
			ID: p.ID, X: round(p.X), Y: round(p.Y), Face: p.Face,
			HP: p.HP, Team: string(p.Team), Dead: p.Dead,
			Invuln: p.InvulnUntil > nowMs,
		})
	}
	for _, pr := range w.Projectiles {
		snap.Projectiles = append(snap.Projectiles, ProjectileState{
			ID: pr.ID, X: round(pr.X), Y: round(pr.Y), Team: string(pr.Team),
		})
	}
	for _, f := range w.Flags {
		snap.Flags = append(snap.Flags, FlagState{ID: f.ID, Team: string(f.Team)})
	}
	return snap
}

// round trims to one decimal to keep snapshots compact.
func round(v float64) float64 {
	return float64(int(v*10)) / 10
}
