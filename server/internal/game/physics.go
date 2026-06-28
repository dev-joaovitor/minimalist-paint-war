package game

import "math"

// Rect is an axis-aligned bounding box with top-left origin (x grows right, y
// grows down).
type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// Intersects reports whether two rectangles overlap.
func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W && r.X+r.W > o.X && r.Y < o.Y+o.H && r.Y+r.H > o.Y
}

// Expand returns the rectangle grown by m on every side.
func (r Rect) Expand(m float64) Rect {
	return Rect{X: r.X - m, Y: r.Y - m, W: r.W + 2*m, H: r.H + 2*m}
}

// Center returns the rectangle's center point.
func (r Rect) Center() (float64, float64) {
	return r.X + r.W/2, r.Y + r.H/2
}

// ContainsPoint reports whether (px,py) lies inside the rectangle.
func (r Rect) ContainsPoint(px, py float64) bool {
	return px >= r.X && px <= r.X+r.W && py >= r.Y && py <= r.Y+r.H
}

// segRectT returns the entry parameter t in [0,1] where the segment p0->p1 first
// enters rect, and whether such an intersection exists. Uses the slab method.
func segRectT(x0, y0, x1, y1 float64, r Rect) (float64, bool) {
	dx := x1 - x0
	dy := y1 - y0
	tmin := 0.0
	tmax := 1.0

	// X slab.
	if dx == 0 {
		if x0 < r.X || x0 > r.X+r.W {
			return 0, false
		}
	} else {
		t1 := (r.X - x0) / dx
		t2 := (r.X + r.W - x0) / dx
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		tmin = math.Max(tmin, t1)
		tmax = math.Min(tmax, t2)
		if tmin > tmax {
			return 0, false
		}
	}

	// Y slab.
	if dy == 0 {
		if y0 < r.Y || y0 > r.Y+r.H {
			return 0, false
		}
	} else {
		t1 := (r.Y - y0) / dy
		t2 := (r.Y + r.H - y0) / dy
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		tmin = math.Max(tmin, t1)
		tmax = math.Min(tmax, t2)
		if tmin > tmax {
			return 0, false
		}
	}

	return tmin, true
}

// segCircleT returns the entry parameter t in [0,1] where segment p0->p1 first
// touches the circle (cx,cy,radius), and whether it does.
func segCircleT(x0, y0, x1, y1, cx, cy, radius float64) (float64, bool) {
	dx := x1 - x0
	dy := y1 - y0
	fx := x0 - cx
	fy := y0 - cy

	a := dx*dx + dy*dy
	if a == 0 {
		if fx*fx+fy*fy <= radius*radius {
			return 0, true
		}
		return 0, false
	}
	b := 2 * (fx*dx + fy*dy)
	c := fx*fx + fy*fy - radius*radius
	disc := b*b - 4*a*c
	if disc < 0 {
		return 0, false
	}
	disc = math.Sqrt(disc)
	t := (-b - disc) / (2 * a)
	if t < 0 {
		t = (-b + disc) / (2 * a)
	}
	if t < 0 || t > 1 {
		return 0, false
	}
	return t, true
}
