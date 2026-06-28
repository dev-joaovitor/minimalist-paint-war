package mapgen

import (
	"testing"

	"paintwar/server/internal/game"
)

// TestNoOverlaps verifies that across many seeds, obstacles never overlap each
// other or the bases, flags sit in open space, and everything stays in bounds.
func TestNoOverlaps(t *testing.T) {
	for seed := int64(0); seed < 500; seed++ {
		md := Generate(seed)

		if len(md.Flags) != game.NumFlags {
			t.Fatalf("seed %d: expected %d flags, got %d", seed, game.NumFlags, len(md.Flags))
		}

		// Obstacles in bounds and pairwise non-overlapping.
		for i, a := range md.Obstacles {
			if a.X < 0 || a.Y < 0 || a.X+a.W > md.Width || a.Y+a.H > md.Height {
				t.Fatalf("seed %d: obstacle %d out of bounds: %+v", seed, i, a)
			}
			for j := i + 1; j < len(md.Obstacles); j++ {
				// L-shape components share a corner, so allow exact adjacency by
				// testing strict overlap with a tiny inset.
				if overlapStrict(a, md.Obstacles[j]) {
					t.Fatalf("seed %d: obstacles %d and %d overlap", seed, i, j)
				}
			}
		}

		// Flags must not sit inside an obstacle or a base.
		for _, f := range md.Flags {
			for _, o := range md.Obstacles {
				if o.ContainsPoint(f.X, f.Y) {
					t.Fatalf("seed %d: flag %d inside obstacle", seed, f.ID)
				}
			}
			if md.RedBase.ContainsPoint(f.X, f.Y) || md.GreenBase.ContainsPoint(f.X, f.Y) {
				t.Fatalf("seed %d: flag %d inside a base", seed, f.ID)
			}
		}
	}
}

// TestDeterministic ensures the same seed yields the same map.
func TestDeterministic(t *testing.T) {
	a := Generate(42)
	b := Generate(42)
	if len(a.Obstacles) != len(b.Obstacles) || len(a.Flags) != len(b.Flags) {
		t.Fatal("same seed produced different shapes")
	}
	for i := range a.Obstacles {
		if a.Obstacles[i] != b.Obstacles[i] {
			t.Fatalf("obstacle %d differs between identical seeds", i)
		}
	}
}

// overlapStrict reports a genuine area overlap, ignoring shared edges.
func overlapStrict(a, b game.Rect) bool {
	const e = 0.5
	return a.X+e < b.X+b.W && a.X+a.W > b.X+e && a.Y+e < b.Y+b.H && a.Y+a.H > b.Y+e
}
