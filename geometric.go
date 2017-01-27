package hex

import "fmt"
import "math"

// GeoCoord is a geometric coordinate (a general coordinate on the 2D plane,
// not restricted to the hex grid).
type GeoCoord struct {
	X, Y float64
}

// String returns a human-readable string form of a GeoCoord.
func (p GeoCoord) String() string {
	return fmt.Sprintf("GeoCoord[%f,%f]", p.X, p.Y)
}

// AngularInterval is a connected interval of angles.
type AngularInterval struct {
	Empty, Full bool
	Rad0, Rad1  float64
}

func radiansToDegrees(x float64) float64 { return x * 180.0 / math.Pi }

// String returns a human-readable string form of an AngularInterval.
func (n AngularInterval) String() string {
	switch {
	case n.Empty:
		return "AngularInterval[empty]"
	case n.Full:
		return "AngularInterval[full]"
	default:
		return fmt.Sprintf("AngularInterval[%0.1f deg (%f rad), %0.1f deg (%f rad)]", radiansToDegrees(n.Rad0), n.Rad0, radiansToDegrees(n.Rad1), n.Rad1)
	}
}

// Size returns the size (in radians) of an angular interval.
func (n AngularInterval) Size() float64 {
	if n.Full {
		return 2 * math.Pi
	}
	if n.Rad1 >= n.Rad0 {
		return n.Rad1 - n.Rad0
	}
	return n.Rad1 + (2*math.Pi - n.Rad0)
}

var (
	// FullAngularInterval is the AngularInterval that contains all angles.
	FullAngularInterval = AngularInterval{Full: true}

	// EmptyAngularInterval is the AngularInterval that contains no angles.
	EmptyAngularInterval = AngularInterval{Empty: true}
)

// NewAngularInterval constructs a new angular interval -- neither empty nor full.
func NewAngularInterval(a0, a1 float64) AngularInterval {
	a0 = transformAngle(a0, 0.0)
	a1 = transformAngle(a1, 0.0)
	return AngularInterval{Rad0: a0, Rad1: a1}
}

// Intersection computes the intersection of two AngularIntervals.
func (n AngularInterval) Intersection(x AngularInterval) AngularInterval {
	switch {
	case n.Empty || x.Empty:
		return EmptyAngularInterval
	case n.Full:
		return x
	case x.Full:
		return n
	default:
		var rad0, rad1 float64
		switch {
		case n.Contains(x.Rad0):
			rad0 = x.Rad0
		case x.Contains(n.Rad0):
			rad0 = n.Rad0
		default:
			return EmptyAngularInterval
		}
		switch {
		case n.Contains(x.Rad1):
			rad1 = x.Rad1
		case x.Contains(n.Rad1):
			rad1 = n.Rad1
		default:
			return EmptyAngularInterval
		}
		return NewAngularInterval(rad0, rad1)
	}
}

// Angle converts a GeoCoord to an angle (from the origin).
func (g GeoCoord) Angle() float64 {
	rv := math.Atan2(g.Y, g.X)
	if rv < 0 {
		rv += 2 * math.Pi
	}
	return rv
}

// Add adds two GeoCoords.
func (g GeoCoord) Add(g2 GeoCoord) GeoCoord {
	return GeoCoord{g.X + g2.X, g.Y + g2.Y}
}

// Sub subtracts two GeoCoords.
func (g GeoCoord) Sub(g2 GeoCoord) GeoCoord {
	return GeoCoord{g.X - g2.X, g.Y - g2.Y}
}

// SquareLength computes the squared length of a GeoCoord (the squared distance
// from the origin).
func (g GeoCoord) SquareLength() float64 {
	return g.X*g.X + g.Y*g.Y
}

// Length computes the distance of a GeoCoord from the origin.
func (g GeoCoord) Length() float64 {
	return math.Sqrt(g.SquareLength())
}

// DistanceTo computes the distance between two GeoCoords.
func (g GeoCoord) DistanceTo(g2 GeoCoord) float64 {
	return g.Sub(g2).Length()
}

// Geo finds the GeoCoord representing the center of the given HexCoord.
func (c HexCoord) Geo() GeoCoord {
	x := 3 * hexHalfSideLength * float64(c.X)
	return GeoCoord{x, float64(c.Y)}
}

// Vertex finds the GeoCoord at the i'th corner vertex of the hexagon.
// i=0 is the east vertex (k,0) from the center, and the vertices proceed
// counterclockwise around the hexagon.
func (c HexCoord) Vertex(i int) GeoCoord {
	cx := 3 * hexHalfSideLength * float64(c.X)
	cy := float64(c.Y)
	k := hexHalfSideLength
	var dx, dy float64
	switch i % 6 {
	case 0:
		dx = 2 * k
	case 1:
		dx = k
		dy = 1
	case 2:
		dx = -k
		dy = 1
	case 3:
		dx = -2 * k
	case 4:
		dx = -k
		dy = -1
	case 5:
		dx = k
		dy = -1
	}
	return GeoCoord{cx + dx, cy + dy}
}

func transformAngle(a, offset float64) float64 {
	tp := 2 * math.Pi
	a = math.Mod(a+offset, tp)
	return math.Mod(a+tp, tp)
}

// ExtremeAngles computes the AngularInterval of this hexagon seen from
// the origin.
func (c HexCoord) ExtremeAngles() AngularInterval {
	if c.IsZero() {
		return FullAngularInterval
	}
	branchCut := 0.0
	if c.X > 0 {
		branchCut = math.Pi
	}
	a0 := transformAngle(c.Vertex(0).Angle(), -branchCut)
	a1 := a0
	for i := 1; i <= 5; i++ {
		a := transformAngle(c.Vertex(i).Angle(), -branchCut)
		if a < a0 {
			a0 = a
		}
		if a > a1 {
			a1 = a
		}
	}
	a0 = transformAngle(a0, branchCut)
	a1 = transformAngle(a1, branchCut)
	return NewAngularInterval(a0, a1)
}

// Contains tests whether another AngularInterval is contained in this one.
func (n AngularInterval) Contains(a float64) bool {
	if n.Empty {
		return false
	}
	if n.Full {
		return true
	}
	a = transformAngle(a, 0)
	if n.Rad0 <= n.Rad1 {
		return n.Rad0 <= a && a <= n.Rad1
	}
	return n.Rad0 <= a || a <= n.Rad1
}

// ContainsRay tests whether the hex would intersect a given ray from
// the origin.
func (c HexCoord) ContainsRay(dx, dy float64) bool {
	p := GeoCoord{dx, dy}
	return c.ExtremeAngles().Contains(p.Angle())
}

// NewGeoPolar creates a GeoCoord, specified with polar coordinates.
func NewGeoPolar(r float64, a float64) GeoCoord {
	return GeoCoord{X: r * math.Cos(a), Y: r * math.Sin(a)}
}

// Scaled produces a new GeoCoord that is a scaled version of the old one.
func (c GeoCoord) Scaled(s float64) GeoCoord {
	return GeoCoord{s * c.X, s * c.Y}
}

// Normalized produces a new GeoCoord that is a normalized
// version of the old one.
func (c GeoCoord) Normalized() GeoCoord {
	return c.Scaled(1.0 / c.Length())
}

// GeoCoord computes the cross product of this GeoCoord with another.
func (c GeoCoord) Cross(v GeoCoord) float64 {
	return c.X*v.X + c.Y*v.Y
}

// NewAngularSextant computes an AngularInterval corresponding to a given
// sextant (i.e. 60-degree section of the plane corresponding to the hex
// directions outwards from the origin).
func NewAngularSextant(section HexDir) AngularInterval {
	base := Directions[section]
	nextDir := OrderedDirections[(int(section)+1)%6]
	rad0 := base.Geo().Angle()
	rad1 := Directions[nextDir].Geo().Angle()
	return NewAngularInterval(rad0, rad1)
}
