package game

// World dimensions (pixels). Aspect roughly matches the design prototype.
const (
	WorldW = 1400.0
	WorldH = 1040.0
)

// Simulation timing.
const (
	TickHz = 20
	Dt     = 1.0 / TickHz // seconds per tick
)

// Player tuning.
const (
	PlayerSpeed = 200.0 // px/sec
	PlayerHalf  = 13.0  // half of the 26px square
	MaxHP       = 3
	RespawnMs   = 3000
	// SpawnProtectMs is the invulnerability window after (re)spawning, giving
	// players a moment to move off the base before they can be damaged.
	SpawnProtectMs = 2000
)

// Projectile tuning.
const (
	ProjSpeed      = 600.0 // px/sec
	ProjRadius     = 5.0
	FireCooldownMs = 400
	MuzzleOffset   = PlayerHalf + ProjRadius + 2
)

// Map composition.
const (
	NumFlags     = 9  // odd to reduce ties
	NumObstacles = 16 // upper bound; mapgen may place fewer
	FlagHitR     = 16.0
)

// s is sin/cos of 45deg, used for diagonal unit vectors.
const s = 0.70710678118

// faceDir maps an 8-direction facing index to a unit vector. Index order:
// 0=E 1=SE 2=S 3=SW 4=W 5=NW 6=N 7=NE (y grows downward).
var faceDir = [8][2]float64{
	{1, 0}, {s, s}, {0, 1}, {-s, s}, {-1, 0}, {-s, -s}, {0, -1}, {s, -s},
}

// faceFromInput returns the facing index for a movement input, or -1 when there
// is no movement (caller keeps the previous facing).
func faceFromInput(dx, dy int) int {
	switch {
	case dx > 0 && dy == 0:
		return 0
	case dx > 0 && dy > 0:
		return 1
	case dx == 0 && dy > 0:
		return 2
	case dx < 0 && dy > 0:
		return 3
	case dx < 0 && dy == 0:
		return 4
	case dx < 0 && dy < 0:
		return 5
	case dx == 0 && dy < 0:
		return 6
	case dx > 0 && dy < 0:
		return 7
	default:
		return -1
	}
}
