package game

import (
	"testing"

	"paintwar/server/internal/model"
)

func baseMap() MapData {
	return MapData{
		Width:     WorldW,
		Height:    WorldH,
		RedBase:   Rect{X: 0, Y: WorldH - 150, W: 150, H: 150},
		GreenBase: Rect{X: WorldW - 150, Y: 0, W: 150, H: 150},
	}
}

func TestPlayerMoves(t *testing.T) {
	md := baseMap()
	w := NewWorld(md, []PlayerSpec{{ID: "p1", Team: model.TeamRed}}, 0, 120000)
	p := w.Players["p1"]
	p.X, p.Y = 500, 400

	w.SetInput("p1", Input{DX: 1})
	w.Step(1000)

	if p.X <= 500 {
		t.Fatalf("expected player to move right, x=%.1f", p.X)
	}
	if p.Face != 0 {
		t.Errorf("expected facing east (0), got %d", p.Face)
	}
}

func TestObstacleBlocksMovement(t *testing.T) {
	md := baseMap()
	md.Obstacles = []Rect{{X: 120, Y: 380, W: 60, H: 60}}
	w := NewWorld(md, []PlayerSpec{{ID: "p1", Team: model.TeamRed}}, 0, 120000)
	p := w.Players["p1"]
	p.X, p.Y = 100, 405 // just left of the obstacle

	w.SetInput("p1", Input{DX: 1})
	w.Step(1000)

	if p.X+PlayerHalf > 120 {
		t.Fatalf("player should be blocked by obstacle, x=%.1f", p.X)
	}
}

func TestProjectileCapturesFlag(t *testing.T) {
	md := baseMap()
	md.Flags = []Flag{{ID: 0, X: 160, Y: 100, Team: model.TeamNone}}
	w := NewWorld(md, []PlayerSpec{{ID: "p1", Team: model.TeamRed}}, 0, 120000)
	p := w.Players["p1"]
	p.X, p.Y = 100, 100
	p.Face = 0 // east, toward the flag

	w.SetInput("p1", Input{Shoot: true})
	w.Step(1000)

	if w.Flags[0].Team != model.TeamRed {
		t.Fatalf("expected flag captured by RED, got %q", w.Flags[0].Team)
	}
}

func TestPlayerDamageDeathRespawn(t *testing.T) {
	md := baseMap()
	w := NewWorld(md, []PlayerSpec{
		{ID: "red", Team: model.TeamRed},
		{ID: "grn", Team: model.TeamGreen},
	}, 0, 120000)
	red := w.Players["red"]
	grn := w.Players["grn"]
	red.X, red.Y = 100, 100
	red.Face = 0
	grn.X, grn.Y = 160, 100

	// Damage only lands after the spawn-protection window has elapsed.
	now := int64(SpawnProtectMs + 1)
	for i := 0; i < MaxHP; i++ {
		w.SetInput("red", Input{Shoot: true})
		w.Step(now)
		now += FireCooldownMs
	}
	if !grn.Dead || grn.HP != 0 {
		t.Fatalf("expected green dead at 0 HP, dead=%v hp=%d", grn.Dead, grn.HP)
	}

	// Before respawn time: still dead.
	w.Step(grn.RespawnAt - 1)
	if !grn.Dead {
		t.Fatal("green should still be dead before respawn time")
	}
	// At respawn time: revived at base with full HP.
	respawnAt := grn.RespawnAt
	w.Step(respawnAt)
	if grn.Dead || grn.HP != MaxHP {
		t.Fatalf("expected green respawned, dead=%v hp=%d", grn.Dead, grn.HP)
	}
	bx, by := md.GreenBase.Center()
	if grn.X != bx || grn.Y != by {
		t.Errorf("expected respawn at base center (%.0f,%.0f), got (%.0f,%.0f)", bx, by, grn.X, grn.Y)
	}
}

func TestSpawnProtectionBlocksDamage(t *testing.T) {
	md := baseMap()
	w := NewWorld(md, []PlayerSpec{
		{ID: "red", Team: model.TeamRed},
		{ID: "grn", Team: model.TeamGreen},
	}, 0, 120000)
	red := w.Players["red"]
	grn := w.Players["grn"]
	red.X, red.Y = 100, 100
	red.Face = 0
	grn.X, grn.Y = 160, 100

	// Within the spawn-protection window: hits are ignored.
	now := int64(500)
	for i := 0; i < MaxHP; i++ {
		w.SetInput("red", Input{Shoot: true})
		w.Step(now)
		now += FireCooldownMs
	}
	if grn.Dead || grn.HP != MaxHP {
		t.Fatalf("expected green unharmed while protected, dead=%v hp=%d", grn.Dead, grn.HP)
	}
}

func TestWinnerByFlagMajority(t *testing.T) {
	md := baseMap()
	md.Flags = []Flag{
		{ID: 0, Team: model.TeamRed},
		{ID: 1, Team: model.TeamGreen},
		{ID: 2, Team: model.TeamGreen},
	}
	w := NewWorld(md, nil, 0, 120000)
	if got := w.Winner(); got != model.TeamGreen {
		t.Fatalf("expected GREEN winner, got %q", got)
	}

	md.Flags = []Flag{{ID: 0, Team: model.TeamRed}, {ID: 1, Team: model.TeamGreen}}
	w = NewWorld(md, nil, 0, 120000)
	if got := w.Winner(); got != model.TeamNone {
		t.Fatalf("expected tie (none), got %q", got)
	}
}
