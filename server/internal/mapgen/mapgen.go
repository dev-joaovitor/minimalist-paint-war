// Package mapgen produces randomized, non-overlapping match maps: obstacles,
// flags, and the two team bases never conflict with one another.
package mapgen

import (
	"math/rand"

	"paintwar/server/internal/game"
)

const (
	edgeMargin  = 28.0  // keep shapes away from the world border
	baseW       = 150.0 // team base size
	baseH       = 150.0
	spawnClear  = 36.0 // reserved breathing room around each base
	obstacleGap = 22.0 // minimum gap between obstacles (so players fit through)
	flagClear   = 34.0 // open space required around a flag
	maxTries    = 80
	lProbPct    = 35 // percent chance an obstacle is an L-shape
)

// Generate builds a deterministic map for the given seed.
func Generate(seed int64) game.MapData {
	rng := rand.New(rand.NewSource(seed))

	// Bases: red bottom-left, green top-right (matching the prototype).
	redBase := game.Rect{X: edgeMargin, Y: game.WorldH - edgeMargin - baseH, W: baseW, H: baseH}
	greenBase := game.Rect{X: game.WorldW - edgeMargin - baseW, Y: edgeMargin, W: baseW, H: baseH}

	// Reserved zones that nothing else may overlap.
	reserved := []game.Rect{redBase.Expand(spawnClear), greenBase.Expand(spawnClear)}

	var obstacles []game.Rect
	for i := 0; i < game.NumObstacles; i++ {
		shape := tryPlaceObstacle(rng, reserved)
		if shape == nil {
			continue // couldn't fit; place fewer rather than fail
		}
		for _, r := range shape {
			obstacles = append(obstacles, r)
			reserved = append(reserved, r)
		}
	}

	flags := make([]game.Flag, 0, game.NumFlags)
	for i := 0; i < game.NumFlags; i++ {
		if x, y, ok := tryPlaceFlag(rng, reserved); ok {
			flags = append(flags, game.Flag{ID: i, X: x, Y: y, Team: ""})
			reserved = append(reserved, game.Rect{
				X: x - flagClear, Y: y - flagClear, W: 2 * flagClear, H: 2 * flagClear,
			})
		}
	}

	return game.MapData{
		Seed:      seed,
		Width:     game.WorldW,
		Height:    game.WorldH,
		Obstacles: obstacles,
		Flags:     flags,
		RedBase:   redBase,
		GreenBase: greenBase,
	}
}

// tryPlaceObstacle returns the rectangles of a non-conflicting obstacle (one for
// a bar, two for an L-shape), or nil if no placement was found.
func tryPlaceObstacle(rng *rand.Rand, reserved []game.Rect) []game.Rect {
	for t := 0; t < maxTries; t++ {
		var shape []game.Rect
		if rng.Intn(100) < lProbPct {
			shape = randomL(rng)
		} else {
			shape = []game.Rect{randomBar(rng)}
		}
		if shape == nil || conflicts(shape, reserved) {
			continue
		}
		return shape
	}
	return nil
}

// randomBar makes a single random rectangle within the playable bounds.
func randomBar(rng *rand.Rand) game.Rect {
	w := 30 + rng.Float64()*150
	h := 20 + rng.Float64()*130
	return game.Rect{
		X: edgeMargin + rng.Float64()*(game.WorldW-2*edgeMargin-w),
		Y: edgeMargin + rng.Float64()*(game.WorldH-2*edgeMargin-h),
		W: w,
		H: h,
	}
}

// randomL makes an L-shape from two rectangles sharing a corner.
func randomL(rng *rand.Rand) []game.Rect {
	thick := 24 + rng.Float64()*16
	long := 90 + rng.Float64()*90
	x := edgeMargin + rng.Float64()*(game.WorldW-2*edgeMargin-long)
	y := edgeMargin + rng.Float64()*(game.WorldH-2*edgeMargin-long)
	// The two arms only share an edge (no area overlap), so the L still reads as
	// one shape while keeping all obstacles pairwise non-overlapping.
	horizontal := game.Rect{X: x, Y: y, W: long, H: thick}
	vertical := game.Rect{X: x, Y: y + thick, W: thick, H: long - thick}
	return []game.Rect{horizontal, vertical}
}

// conflicts reports whether any rect in shape (grown by the obstacle gap) hits a
// reserved rect or leaves the world.
func conflicts(shape, reserved []game.Rect) bool {
	for _, r := range shape {
		if r.X < edgeMargin || r.Y < edgeMargin ||
			r.X+r.W > game.WorldW-edgeMargin || r.Y+r.H > game.WorldH-edgeMargin {
			return true
		}
		grown := r.Expand(obstacleGap)
		for _, res := range reserved {
			if grown.Intersects(res) {
				return true
			}
		}
	}
	return false
}

// tryPlaceFlag finds an open point with flagClear space around it.
func tryPlaceFlag(rng *rand.Rand, reserved []game.Rect) (float64, float64, bool) {
	for t := 0; t < maxTries; t++ {
		x := edgeMargin + flagClear + rng.Float64()*(game.WorldW-2*(edgeMargin+flagClear))
		y := edgeMargin + flagClear + rng.Float64()*(game.WorldH-2*(edgeMargin+flagClear))
		zone := game.Rect{X: x - flagClear, Y: y - flagClear, W: 2 * flagClear, H: 2 * flagClear}
		clear := true
		for _, res := range reserved {
			if zone.Intersects(res) {
				clear = false
				break
			}
		}
		if clear {
			return x, y, true
		}
	}
	return 0, 0, false
}
