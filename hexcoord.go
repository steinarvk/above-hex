package hex

import "fmt"

import (
	"math/rand"

	pb "github.com/steinarvk/above-hex/hexpb"
)

// HexCoord is a coordinate on a hex grid.
// (X,Y) is a valid hex coordinate if and only if X and Y are of equal parity.
type HexCoord struct {
	X, Y int
}

const (
	sqrt3             = 1.7320508075688772
	hexHalfSideLength = 1.0 / sqrt3
)

// NewHex creates a new HexCoord, without checking that it is valid.
func NewHex(x, y int) HexCoord {
	return HexCoord{x, y}
}

// TryNewHex creates a new HexCoord, returning an error if the two integers do
// not specify a valid coordinate (i.e. are not of equal parity).
func TryNewHex(x, y int) (HexCoord, error) {
	xEven := (x % 2) == 0
	yEven := (y % 2) == 0
	if xEven != yEven {
		return Origin, fmt.Errorf("no such hex")
	}
	return NewHex(x, y), nil
}

// HexDisk computes a slice of all the HexCoords at or below a certain radius.
func HexDisk(r int) []HexCoord {
	var rv []HexCoord
	for i := 0; i <= r; i++ {
		rv = append(rv, HexCircle(i)...)
	}
	return rv
}

// RandomHexOnCircle returns a random hex on a "hex circle".
func RandomHexOnCircle(r *rand.Rand, radius int) HexCoord {
	coords := HexCircle(radius)
	return coords[r.Intn(len(coords))]
}

// HexCircle computes a slice of all the HexCoords at a certain radius.
func HexCircle(r int) []HexCoord {
	var rv []HexCoord

	n := 6 * r
	if r == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		rv = append(rv, NewHexPolar(r, i))
	}

	return rv
}

// NewHexPolar computes a HexCoord using integral polar coordinates.
func NewHexPolar(r, step int) HexCoord {
	if r == 0 {
		return NewHex(0, 0)
	}

	circum := HexCircumference(r)
	if step > circum {
		step = step % circum
	}

	section := step / r
	substep := step % r

	direction := OrderedDirections[section]
	baseDelta := Directions[direction]
	orthoDelta := Directions[OrthogonalCCW[direction]]

	base := Origin.AddMultDelta(r, baseDelta)
	return base.AddMultDelta(substep, orthoDelta)
}

// AbsX returns the absolute value of the HexCoord's X coordinate.
func (c *HexCoord) AbsX() int {
	if c.X >= 0 {
		return c.X
	}
	return -c.X
}

// AbsY returns the absolute value of the HexCoord's Y coordinate.
func (c *HexCoord) AbsY() int {
	if c.Y >= 0 {
		return c.Y
	}
	return -c.Y
}

// Radius computes the radius of the HexCoord.
func (c *HexCoord) Radius() int {
	rv := c.AbsX() + c.AbsY()
	if rv%2 != 0 {
		panic(fmt.Errorf("programming error: sum of hex coordinates (%v) is odd", c))
	}
	smallSteps := c.AbsX()
	bigSteps := (c.AbsY() - c.AbsX()) / 2
	if bigSteps < 0 {
		bigSteps = 0
	}
	return bigSteps + smallSteps
}

// Negation computes the negation of the HexCoord (i.e. scaled by -1).
func (c HexCoord) Negation() HexCoord {
	return Origin.AddMultDelta(-1, c)
}

// AddDelta adds two HexCoords.
func (c *HexCoord) AddDelta(d HexCoord) HexCoord {
	return c.AddMultDelta(1, d)
}

// Move computes a HexCoord that results after taking an number of steps
// from an initial HexCoord.
func (c *HexCoord) Move(ds ...HexDir) HexCoord {
	rv := *c
	for _, d := range ds {
		dc := Directions[d]
		rv = HexCoord{c.X + dc.X, c.Y + dc.Y}
	}
	return rv
}

// NeighboursSet computes the HexSet of a HexCoord's immediate neighbours.
func (c *HexCoord) NeighboursSet() *HexSet {
	rv := NewHexSet()
	for _, d := range OrderedDirections {
		rv.Add(c.AddDelta(Directions[d]))
	}
	return rv
}

// Neighbours computes a slice of the HexCoord's immediate neighbours.
func (c *HexCoord) Neighbours() []HexCoord {
	rv := make([]HexCoord, 6)
	for i, d := range OrderedDirections {
		rv[i] = c.AddDelta(Directions[d])
	}
	return rv
}

// String computes a human-readable string form of a HexCoord.
func (c HexCoord) String() string {
	return fmt.Sprintf("Hex[%d,%d]", c.X, c.Y)
}

// AddMultDelta adds a multiple of a HexCoord; i.e. c + m * d.
func (c *HexCoord) AddMultDelta(m int, d HexCoord) HexCoord {
	return HexCoord{c.X + m*d.X, c.Y + m*d.Y}
}

// IsZero tests whether a HexCoord is the origin.
func (c HexCoord) IsZero() bool {
	return c.X == 0 && c.Y == 0
}

// ToProto converts a HexCoord to a proto.
func (c HexCoord) ToProto() *pb.HexCoord {
	return &pb.HexCoord{X: int32(c.X), Y: int32(c.Y)}
}

// HexFromProto converts a proto to a HexCoord.
func HexFromProto(p *pb.HexCoord) (HexCoord, error) {
	return TryNewHex(int(p.X), int(p.Y))
}

// Minus subtracts two HexCoords.
func (c HexCoord) Minus(p HexCoord) HexCoord {
	return c.AddMultDelta(-1, p)
}

// Less computes a lexicographic ordering of HexCoords.
func (a HexCoord) Less(b HexCoord) bool {
	// c < x
	if a.X < b.X {
		return true
	}
	if a.X > b.X {
		return false
	}
	if a.Y < b.Y {
		return true
	}
	return false
}
