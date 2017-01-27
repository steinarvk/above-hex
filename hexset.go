package hex

import "fmt"
import "math/rand"

import (
	"github.com/bradfitz/slice"
	"log"
	"math"
	"strings"

	pb "github.com/steinarvk/above-hex/hexpb"
)

type hexSetIntf interface {
	Add(p HexCoord)
	Remove(p HexCoord)
	Contains(p HexCoord) bool
	Enumerate() []HexCoord
	MaxRadius() int
	Size() int
}

type hexSetBottom struct {
	members         map[HexCoord]bool
	cachedMaxRadius int
}

type hexSetModifiedClone struct {
	frozenHexSet hexSetIntf
	mutated      bool
	additions    *HexSet
	removals     *HexSet
}

func (h *hexSetModifiedClone) ensureMutable() {
	if !h.mutated {
		h.mutated = true
		h.additions = NewHexSet()
		h.removals = NewHexSet()
	}
}

func (h *hexSetModifiedClone) Add(p HexCoord) {
	h.ensureMutable()
	if h.removals.Contains(p) {
		h.removals.Remove(p)
	}
	if h.frozenHexSet.Contains(p) {
		return
	}
	h.additions.Add(p)
}

func (h *hexSetModifiedClone) Remove(p HexCoord) {
	h.ensureMutable()
	if h.additions.Contains(p) {
		h.additions.Remove(p)
	}
	if !h.frozenHexSet.Contains(p) {
		return
	}
	h.removals.Add(p)
}

func (h *hexSetModifiedClone) Contains(p HexCoord) bool {
	if h.frozenHexSet.Contains(p) {
		return !h.mutated || !h.removals.Contains(p)
	} else {
		return h.mutated && h.additions.Contains(p)
	}
}

func (h *hexSetModifiedClone) Size() int {
	rv := h.frozenHexSet.Size()
	if h.mutated {
		rv += h.additions.Size() - h.removals.Size()
	}
	return rv
}

func (h *hexSetModifiedClone) Enumerate() []HexCoord {
	rv := []HexCoord{}
	for _, p := range h.frozenHexSet.Enumerate() {
		if !h.mutated || !h.removals.Contains(p) {
			rv = append(rv, p)
		}
	}
	if h.mutated {
		for _, p := range h.additions.Enumerate() {
			rv = append(rv, p)
		}
	}
	return rv
}

func (h *hexSetModifiedClone) MaxRadius() int {
	maxR := h.frozenHexSet.MaxRadius()
	if h.mutated {
		if r := h.removals.MaxRadius(); r == maxR {
			// Fall back to naive solution
			maxR = 0
			for _, p := range h.Enumerate() {
				if r := p.Radius(); r > maxR {
					maxR = r
				}
			}
			return maxR
		}
		if r := h.additions.MaxRadius(); r > maxR {
			maxR = r
		}
	}
	return maxR
}

func (h *hexSetBottom) Add(p HexCoord) {
	h.members[p] = true
	if r := p.Radius(); h.cachedMaxRadius != -1 && r > h.cachedMaxRadius {
		h.cachedMaxRadius = r
	}
}

func (h *hexSetBottom) Remove(p HexCoord) {
	delete(h.members, p)
	if h.cachedMaxRadius != -1 && p.Radius() == h.cachedMaxRadius {
		h.cachedMaxRadius = -1
	}
}

func (h *hexSetBottom) Enumerate() []HexCoord {
	rv := []HexCoord{}
	for p, _ := range h.members {
		rv = append(rv, p)
	}
	return rv
}

func (h *hexSetBottom) Size() int {
	return len(h.members)
}

func (h *hexSetBottom) Contains(p HexCoord) bool {
	_, present := h.members[p]
	return present
}

func (h *hexSetBottom) MaxRadius() int {
	switch {
	case h.cachedMaxRadius == 0:
		if h.Size() == 0 || (h.Size() == 1 && h.Contains(Origin)) {
			return 0
		}
		panic(fmt.Errorf("hexSetBottom not initialized properly"))
	case h.cachedMaxRadius != -1:
		return h.cachedMaxRadius
	default:
		rv := 0
		for p, _ := range h.members {
			r := p.Radius()
			if r > rv {
				rv = r
			}
		}
		h.cachedMaxRadius = rv
		return rv
	}
}

// HexSet is a set of HexCoords.
type HexSet struct {
	impl          hexSetIntf
	frozen        bool
	cloneChildren []*HexSet
}

// Freeze marks the HexSet as immutable.
func (s *HexSet) Freeze() {
	s.frozen = true
}

func (s *HexSet) explicitClone() *HexSet {
	rv := NewHexSet()
	for _, p := range s.Enumerate() {
		rv.Add(p)
	}
	return rv
}

func (s *HexSet) ensureThawed() {
	if s.frozen {
		panic("attempting to modify frozen HexSet")
	}
	if len(s.cloneChildren) != 0 {
		log.Printf("oh no! rescuing mutable set with frozen children")
		rv := s.explicitClone()
		rv.Freeze()
		for _, ch := range s.cloneChildren {
			ch.impl = &hexSetModifiedClone{
				frozenHexSet: rv,
			}
		}
		s.cloneChildren = nil
	}
}

// NewHexSet creates a new (empty and unfrozen) HexSet.
func NewHexSet() *HexSet {
	return &HexSet{
		impl: &hexSetBottom{
			members:         make(map[HexCoord]bool),
			cachedMaxRadius: -1,
		},
	}
}

// NewHexSetSingleton creates a new HexSet containing a single HexCoord.
func NewHexSetSingleton(p HexCoord) *HexSet {
	rv := NewHexSet()
	rv.Add(p)
	return rv
}

// NewHexSetAround creates a new HexSet containing one HexCoord and all
// HexCoords within a certain radius of it.
func NewHexSetAround(p HexCoord, r int) *HexSet {
	rv := NewHexSetSingleton(p)
	rv.Expand(r)
	return rv
}

// Enumerate converts a HexSet to a slice of HexCoords.
func (h *HexSet) Enumerate() []HexCoord {
	return h.impl.Enumerate()
}

// Random selects a random HexCoord from a HexSet.
func (h *HexSet) Random() (HexCoord, error) {
	return h.RandomFrom(rand.New(rand.NewSource(rand.Int63())))
}

// RandomFrom selects a random HexCoord from a HexSet, using a specified
// rand.Rand.
func (h *HexSet) RandomFrom(r *rand.Rand) (HexCoord, error) {
	elements := h.Enumerate()

	n := int32(len(elements))
	if n == 0 {
		return Origin, fmt.Errorf("random choice from empty set")
	}

	i := r.Int31n(n)
	for _, v := range elements {
		if i == 0 {
			return v, nil
		}

		i--
	}

	return Origin, fmt.Errorf("impossible choice was made in set selection")
}

// PickArbitrary picks an arbitrary HexCoord from a HexSet.
func (h *HexSet) PickArbitrary() (HexCoord, error) {
	if h.Size() == 0 {
		return Origin, fmt.Errorf("empty set")
	}
	return h.Enumerate()[0], nil
}

// ToList converts a HexSet to a list of HexCoords.
func (h *HexSet) ToList() []HexCoord {
	return h.ToOrderedList()
}

// ToOrderedList converts a HexSet to an ordered list of HexCoords.
func (h *HexSet) ToOrderedList() []HexCoord {
	rv := []HexCoord{}

	for _, p := range h.Enumerate() {
		rv = append(rv, p)
	}

	slice.Sort(rv[:], func(i, j int) bool {
		return rv[i].Less(rv[j])
	})

	return rv
}

// Contains checks whether a HexSet contains a certain HexCoord.
func (h *HexSet) Contains(x HexCoord) bool {
	return h.impl.Contains(x)
}

// ContainsSet checks whether a HexSet contains an entire other set.
func (h *HexSet) ContainsSet(xs *HexSet) bool {
	for _, p := range xs.Enumerate() {
		if !h.Contains(p) {
			return false
		}
	}
	return true
}

// Size returns the number of elements in the HexSet.
func (h *HexSet) Size() int {
	return h.impl.Size()
}

// Clone makes a clone of a HexSet.
func (h *HexSet) Clone() *HexSet {
	rv := &HexSet{
		impl: &hexSetModifiedClone{
			frozenHexSet: h,
		},
	}

	if !h.frozen {
		h.cloneChildren = append(h.cloneChildren, rv)
	}

	return rv
}

// Difference removes all elements in another HexSet from this HexSet.
func (h *HexSet) Difference(x *HexSet) *HexSet {
	rv := NewHexSet()

	for _, p := range h.Enumerate() {
		if x.Contains(p) {
			continue
		}
		rv.Add(p)
	}

	return rv
}

// IntersectsWithOffset checks whether a HexSet intersects with the result
// of mapping over another by (+offset), i.e. adding a certain HexCoord delta
// to each element.
func (h *HexSet) IntersectsWithOffset(x *HexSet, offset HexCoord) bool {
	negOffset := offset.Negation()
	if x.Size() < h.Size() {
		return x.IntersectsWithOffset(h, negOffset)
	}

	for _, p := range h.Enumerate() {
		pp := p.AddDelta(negOffset)
		if x.Contains(pp) {
			return true
		}
	}

	return false
}

// Intersects checks whether two HexSets intersect.
func (h *HexSet) Intersects(x *HexSet) bool {
	if x.Size() < h.Size() {
		return x.Intersects(h)
	}

	for _, p := range h.Enumerate() {
		if x.Contains(p) {
			return true
		}
	}

	return false
}

// Intersection computes the intersection of two HexSets.
func (h *HexSet) Intersection(x *HexSet) *HexSet {
	rv := NewHexSet()

	for _, p := range h.Enumerate() {
		if x.Contains(p) {
			rv.Add(p)
		}
	}

	return rv
}

// InPlaceUnion mutates the HexSet to take the union with another.
func (h *HexSet) InPlaceUnion(x *HexSet) {
	for _, p := range x.Enumerate() {
		h.impl.Add(p)
	}
}

// Except computes a new HexSet that is like this one, except not containing
// a certain HexCoord.
func (h *HexSet) Except(p HexCoord) *HexSet {
	rv := h.Clone()
	rv.Remove(p)
	return rv
}

// Union computes a new HexSet that is like this one union'd with another.
func (h *HexSet) Union(x *HexSet) *HexSet {
	rv := h.Clone()
	rv.InPlaceUnion(x)
	return rv
}

// Add adds a HexCoord to the HexSet.
func (h *HexSet) Add(x HexCoord) {
	h.ensureThawed()
	h.impl.Add(x)
}

// RemoveHex removes a certain HexCoord (specified explicitly with integers) from
// the HexSet.
func (h *HexSet) RemoveHex(i, j int) {
	h.Remove(NewHex(i, j))
}

// RemoveHex adds a certain HexCoord (specified explicitly with integers) to
// the HexSet.
func (h *HexSet) AddHex(i, j int) {
	h.Add(NewHex(i, j))
}

// ContainsHex checks whether a certain HexCoord (specified explicitly with
// integers) is in the HexSet.
func (h *HexSet) ContainsHex(i, j int) bool {
	return h.Contains(NewHex(i, j))
}

// Remove removes a HexCoord from the HexSet.
func (h *HexSet) Remove(x HexCoord) {
	h.ensureThawed()
	h.impl.Remove(x)
}

func (h *HexSet) expandOneFiltered(f func(HexCoord) bool) {
	done := map[HexCoord]bool{}

	for _, k := range h.ToList() {
		for _, n := range k.Neighbours() {
			if h.Contains(n) || done[n] {
				continue
			}

			if f(n) {
				h.Add(n)
			}

			done[n] = true
		}
	}
}

// ExpandFiltered expands the HexSet n steps outwards; except not proceeding
// into any hex p where f(p) return false.
func (h *HexSet) ExpandFiltered(n int, f func(HexCoord) bool) {
	for i := 0; i < n; i++ {
		h.expandOneFiltered(f)
	}
}

// Expand expands the HexSet n steps outward from each coordinate.
func (h *HexSet) Expand(n int) {
	always := func(_ HexCoord) bool { return true }
	h.ExpandFiltered(n, always)
}

// Expanded computes a new HexSet that is like the old one, except
// expanded by n steps.
func (h *HexSet) Expanded(n int) *HexSet {
	rv := h.Clone()
	rv.Expand(n)
	return rv
}

// Filtered computes a new HexSet that is like the old one, except
// only containing HexCoords p where f(p) returns true.
func (h *HexSet) Filtered(f func(HexCoord) bool) *HexSet {
	rv := h.Clone()
	rv.Filter(f)
	return rv
}

// Filter removes any point p from the HexSet where f(p) returns false.
func (h *HexSet) Filter(f func(HexCoord) bool) {
	for _, p := range h.Enumerate() {
		if !f(p) {
			h.Remove(p)
		}
	}
}

// Each runs a callback on each of the HexCoords in a HexSet.
func (h *HexSet) Each(f func(HexCoord)) {
	h.Filter(func(x HexCoord) bool {
		f(x)
		return true
	})
}

// MaxRadius returns an upper bound on the radius of the HexCoords in the set.
func (h *HexSet) MaxRadius() int {
	return h.impl.MaxRadius()
}

// Predicate converts the HexSet to a callback that checks for HexCoord membership in the set.
func (c *HexSet) Predicate() func(HexCoord) bool {
	return func(x HexCoord) bool {
		return c.Contains(x)
	}
}

// GeoDistanceHeuristic returns a function f(x) that, given a HexCoord x, computes
// the minimum geometric distance to any point p in the HexSet.
func (c *HexSet) GeoDistanceHeuristic() func(HexCoord) float64 {
	return func(x HexCoord) float64 {
		p := x.Geo()
		sqd := math.Inf(1)
		for _, y := range c.Enumerate() {
			q := y.Geo()
			sql := p.Sub(q).SquareLength()
			if sql < sqd {
				sqd = sql
			}
		}
		return math.Sqrt(sqd)
	}
}

// ToProto converts a HexSet to a pb.HexSet proto.
func (c *HexSet) ToProto() *pb.HexSet {
	rv := pb.HexSet{}
	for _, x := range c.Enumerate() {
		rv.Coords = append(rv.Coords, x.ToProto())
	}
	return &rv
}

// OuterBorder computes the outer border of the HexSet.
func (s *HexSet) OuterBorder() *HexSet {
	rv := NewHexSet()
	excluded := HexSetWithHolesFilled(s)
	for _, p := range s.Enumerate() {
		for _, nb := range p.Neighbours() {
			if !excluded.Contains(nb) {
				rv.Add(nb)
			}
		}
	}
	return rv
}

// HexSetFromProto converts a pb.HexSet proto to a HexSet.
func HexSetFromProto(p *pb.HexSet) (*HexSet, error) {
	rv := NewHexSet()
	if p != nil {
		for _, x := range p.Coords {
			coord, err := HexFromProto(x)
			if err != nil {
				return nil, err
			}

			rv.Add(coord)
		}
	}
	return rv, nil
}

// GoRepr returns an explicit Go representation of the HexSet (code that would
// produce it).
func (s *HexSet) GoRepr() string {
	lines := []string{"s := hex.NewHexSet()"}
	for _, p := range s.ToList() {
		line := fmt.Sprintf("s.AddHex(%d, %d)", p.X, p.Y)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
