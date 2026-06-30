package game

import "math"

// Step advances the simulation by one tick. nowMs is the current wall-clock time
// in milliseconds, used for cooldowns, respawns, and the match timer.
func (w *World) Step(nowMs int64) {
	w.Tick++
	w.fireProjectiles(nowMs)
	w.movePlayers()
	w.advanceProjectiles(nowMs)
	w.respawn(nowMs)
}

// fireProjectiles spawns shots for alive players that requested one and are off
// cooldown.
func (w *World) fireProjectiles(nowMs int64) {
	for _, id := range w.order {
		p := w.Players[id]
		if p.Dead || !p.input.Shoot {
			continue
		}
		if nowMs-p.lastShotMs < FireCooldownMs {
			continue
		}
		p.lastShotMs = nowMs
		dir := faceDir[p.Face]
		w.nextProjID++
		w.Projectiles = append(w.Projectiles, &Projectile{
			ID:      w.nextProjID,
			X:       p.X + dir[0]*MuzzleOffset,
			Y:       p.Y + dir[1]*MuzzleOffset,
			VX:      dir[0] * ProjSpeed,
			VY:      dir[1] * ProjSpeed,
			Team:    p.Team,
			OwnerID: p.ID,
		})
	}
}

// movePlayers integrates movement with axis-separated sliding against obstacles
// and world bounds.
func (w *World) movePlayers() {
	for _, id := range w.order {
		p := w.Players[id]
		if p.Dead {
			continue
		}
		if f := faceFromInput(p.input.DX, p.input.DY); f >= 0 {
			p.Face = f
		}
		if p.input.DX == 0 && p.input.DY == 0 {
			continue
		}
		nx := float64(p.input.DX)
		ny := float64(p.input.DY)
		if p.input.DX != 0 && p.input.DY != 0 {
			nx *= s
			ny *= s
		}
		vx := nx * PlayerSpeed * Dt
		vy := ny * PlayerSpeed * Dt

		if cand := p.X + vx; !w.playerBlocked(cand, p.Y) {
			p.X = cand
		}
		if cand := p.Y + vy; !w.playerBlocked(p.X, cand) {
			p.Y = cand
		}
	}
}

// playerBlocked reports whether a player centered at (x,y) would leave the world
// or overlap an obstacle.
func (w *World) playerBlocked(x, y float64) bool {
	if x-PlayerHalf < 0 || x+PlayerHalf > w.W || y-PlayerHalf < 0 || y+PlayerHalf > w.H {
		return true
	}
	box := Rect{X: x - PlayerHalf, Y: y - PlayerHalf, W: 2 * PlayerHalf, H: 2 * PlayerHalf}
	for _, o := range w.Obstacles {
		if box.Intersects(o) {
			return true
		}
	}
	return false
}

// advanceProjectiles moves each projectile and resolves the nearest collision
// along its path (obstacle, enemy player, or flag), using swept tests.
func (w *World) advanceProjectiles(nowMs int64) {
	kept := w.Projectiles[:0]
	for _, pr := range w.Projectiles {
		x1 := pr.X + pr.VX*Dt
		y1 := pr.Y + pr.VY*Dt

		bestT := math.Inf(1)
		var onHit func()

		// Obstacles.
		for _, o := range w.Obstacles {
			if t, ok := segRectT(pr.X, pr.Y, x1, y1, o.Expand(ProjRadius)); ok && t < bestT {
				bestT = t
				onHit = func() {} // absorbed
			}
		}
		// Enemy players (alive).
		for _, id := range w.order {
			tp := w.Players[id]
			if tp.Dead || tp.Team == pr.Team {
				continue
			}
			box := Rect{X: tp.X - PlayerHalf, Y: tp.Y - PlayerHalf, W: 2 * PlayerHalf, H: 2 * PlayerHalf}
			if t, ok := segRectT(pr.X, pr.Y, x1, y1, box.Expand(ProjRadius)); ok && t < bestT {
				bestT = t
				victim := tp
				onHit = func() { w.damage(victim, nowMs) }
			}
		}
		// Flags (capturable when not already owned by the shooter's team).
		for _, f := range w.Flags {
			if f.Team == pr.Team {
				continue
			}
			if t, ok := segCircleT(pr.X, pr.Y, x1, y1, f.X, f.Y, FlagHitR+ProjRadius); ok && t < bestT {
				bestT = t
				flag := f
				team := pr.Team
				onHit = func() { flag.Team = team }
			}
		}

		if onHit != nil {
			onHit()
			continue // projectile consumed
		}

		// No collision: advance, drop if it left the world.
		pr.X, pr.Y = x1, y1
		if pr.X < 0 || pr.X > w.W || pr.Y < 0 || pr.Y > w.H {
			continue
		}
		kept = append(kept, pr)
	}
	// Clear the tail of the reused slice to avoid retaining dropped projectiles.
	for i := len(kept); i < len(w.Projectiles); i++ {
		w.Projectiles[i] = nil
	}
	w.Projectiles = kept
}

// damage applies one hit; at 0 HP the player dies and schedules a respawn.
// Players are immune while their spawn-protection window is active.
func (w *World) damage(p *Player, nowMs int64) {
	if p.Dead || nowMs < p.InvulnUntil {
		return
	}
	p.HP--
	if p.HP <= 0 {
		p.HP = 0
		p.Dead = true
		p.RespawnAt = nowMs + RespawnMs
		p.input = Input{}
	}
}

// respawn revives dead players whose timer has elapsed.
func (w *World) respawn(nowMs int64) {
	for _, id := range w.order {
		p := w.Players[id]
		if p.Dead && nowMs >= p.RespawnAt {
			w.spawn(p, nowMs)
		}
	}
}
