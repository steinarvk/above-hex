package hex

import (
	"fmt"
	"sort"

	"github.com/oleiade/lane"
)

// HexCircumference computes the circumference (in number of cells) of a
// hex "circle" of radius r, where with the degenerate case of r=0 representing
// a single cell.
func HexCircumference(r int) int {
	if r == 0 {
		return 1
	}

	return 6 * r
}

// RayIntersection finds the first intersection of a ray (where the direction
// is cell-aligned) with a HexSet. An error is returned if no intersection
// exists.
func RayIntersection(start, delta HexCoord, set *HexSet) (HexCoord, error) {
	sanityLimit := set.MaxRadius()
	pos := start

	for {
		if set.Contains(pos) {
			return pos, nil
		}

		lastRadius := pos.Radius()
		pos = pos.AddDelta(delta)
		radiusDelta := pos.Radius() - lastRadius

		if radiusDelta > 0 && lastRadius > sanityLimit {
			return Origin, fmt.Errorf("ray %s + %s * t will never intersect set", start, delta)
		}
	}
}

// BreadthFirstSearch runs a breadth-first search through a hex grid. The hex grid is
// specified through the two functions "isSteppable" (which determines whether a step from
// one coordinate to another is legal) and "isGoal" (which determines whether the goal
// has been reached).
func BreadthFirstSearch(start HexCoord, isSteppable func(HexCoord, HexCoord) bool, isGoal func(HexCoord) bool) ([]HexCoord, error) {
	return BreadthFirstSearchFromMultiple(NewHexSetSingleton(start), isSteppable, isGoal)
}

// BreadthFirstSearchFromMultiple performs a breadth-first search (like BreadthFirstSearch),
// except starting from multiple coordinates at once.
func BreadthFirstSearchFromMultiple(frontierSet *HexSet, isSteppable func(HexCoord, HexCoord) bool, isGoal func(HexCoord) bool) ([]HexCoord, error) {
	frontier := frontierSet.ToList()
	trail := map[HexCoord]*HexCoord{}
	for _, p := range frontier {
		trail[p] = nil
	}

	for {
		if len(frontier) == 0 {
			return nil, fmt.Errorf("no path to goal")
		}

		parent := frontier[0]
		frontier = frontier[1:]

		if isGoal(parent) {
			path := []HexCoord{}
			node := parent

			for {
				path = append([]HexCoord{node}, path...)
				if trail[node] == nil {
					return path, nil
				}
				node = *trail[node]
			}
		}

		for _, nb := range parent.Neighbours() {
			_, seen := trail[nb]
			if seen || !isSteppable(parent, nb) {
				continue
			}
			trail[nb] = &parent
			frontier = append(frontier, nb)
		}
	}
}

// HexSetWithHolesFilled returns a HexSet with any internal holes filled.
func HexSetWithHolesFilled(s *HexSet) *HexSet {
	potentialHoleHexes := []HexCoord{}

	var current *HexCoord

	for _, p := range s.ToOrderedList() {
		if current != nil {
			if current.X == p.X {
				if p.Y <= current.Y {
					panic(fmt.Errorf("ToOrderedList() returned %v after %v", p, current))
				}
				for y := current.Y + 2; y < p.Y; y += 2 {
					potentialHoleHexes = append(potentialHoleHexes, NewHex(current.X, y))
				}
			} else {
				if p.X <= current.X {
					panic(fmt.Errorf("ToOrderedList() returned %v after %v", p, current))
				}
			}
			*current = p
		} else {
			rrv := NewHex(p.X, p.Y)
			current = &rrv
		}
	}

	maxR := s.MaxRadius()
	shownLake := NewHexSet()
	shownOcean := NewHexSet()

	for _, p := range potentialHoleHexes {
		if shownLake.Contains(p) || shownOcean.Contains(p) {
			continue
		}

		body := NewHexSet()
		isOcean := false

		q := lane.NewPQueue(lane.MAXPQ)
		q.Push(p, p.Radius())

		for !isOcean && q.Size() > 0 {
			headUncast, _ := q.Pop()
			head := headUncast.(HexCoord)

			body.Add(head)

			for _, nb := range head.Neighbours() {
				if body.Contains(nb) || s.Contains(nb) {
					continue
				}
				r := nb.Radius()
				if r > maxR || shownOcean.Contains(nb) {
					isOcean = true
					break
				}
				q.Push(nb, r)
			}
		}

		if isOcean {
			shownOcean.InPlaceUnion(body)
		} else {
			shownLake.InPlaceUnion(body)
		}
	}

	return s.Union(shownLake)
}

// IsFullyConnected checks whether a HexSet is fully connected.
func IsFullyConnected(s *HexSet) bool {
	if s.Size() == 0 {
		return true
	}

	any, err := s.PickArbitrary()
	if err != nil {
		panic(fmt.Errorf("unexpected error in PickArbitrary(): %v", err))
	}

	component, err := FindConnectedComponent(any, s.Predicate())
	if err != nil {
		panic(fmt.Errorf("unexpected error in FindConnectedComponent(): %v", err))
	}

	return component.Size() == s.Size()
}

// FindConnectedComponent determines, given a representative coordinate and
// a membership-test function, the connected component of the representative.
func FindConnectedComponent(start HexCoord, isContained func(HexCoord) bool) (*HexSet, error) {
	rv := NewHexSetSingleton(start)
	q := []HexCoord{start}

	for len(q) > 0 {
		h := q[0]
		q = q[1:]

		for _, nb := range h.Neighbours() {
			if !rv.Contains(nb) && isContained(nb) {
				rv.Add(nb)
				q = append(q, nb)
			}
		}
	}

	return rv, nil
}

// DepthFirstSearch performs a depth-first search on a hex grid. Like BreadthFirstSearch, the
// hex grid is specified through callback functions "isSteppable" and "isGoal".
func DepthFirstSearch(start HexCoord, isSteppable func(HexCoord, HexCoord) bool, isGoal func(HexCoord) bool) ([]HexCoord, error) {
	trail := map[HexCoord]*HexCoord{start: nil}
	frontier := []HexCoord{start}

	for {
		if len(frontier) == 0 {
			return nil, fmt.Errorf("no path to goal")
		}

		parent := frontier[0]
		frontier = frontier[1:]

		if isGoal(parent) {
			path := []HexCoord{}
			node := parent

			for {
				path = append([]HexCoord{node}, path...)
				if trail[node] == nil {
					return path, nil
				}
				node = *trail[node]
			}
		}

		for _, nb := range parent.Neighbours() {
			_, seen := trail[nb]
			if seen || !isSteppable(parent, nb) {
				continue
			}
			trail[nb] = &parent
			frontier = append([]HexCoord{nb}, frontier...)
		}
	}
}

func sweepIntervalInSextant(n AngularInterval, section HexDir, r int, visit func(HexCoord)) error {
	sext := NewAngularSextant(section)

	n = n.Intersection(sext)
	if n.Empty {
		return nil
	}
	if r == 0 {
		visit(Origin)
		return nil
	}
	base := Origin.AddMultDelta(r, Directions[section])
	delta := Directions[OrthogonalCCW[section]]

	k := r
	if k == 0 {
		k = 1
	}

	i := sort.Search(k, func(i int) bool {
		p := base.AddMultDelta(i, delta)
		interval := NewAngularInterval(sext.Rad0, p.ExtremeAngles().Rad1)
		return interval.Contains(n.Rad0)
	})

	for {
		p := base.AddMultDelta(i, delta)
		angles := p.ExtremeAngles()
		unseen := NewAngularInterval(angles.Rad0, sext.Rad1)
		first := sext.Intersection(angles).Size() > 0
		second := unseen.Contains(n.Rad1)
		if !first || !second {
			break
		}
		visit(p)
		i++
	}

	return nil
}

// CalculateFov calculates shadowcasting field-of-view from the origin. FOV may be restricted
// by angle or by radius (use -1) to not restrict by radius.
// Obstructions are specified through the "obstruct" callback.
// Output happens with the "addLight" callback which adds light to a certain HexCoord.
// Note that addLight can be called multiple times for one HexCoord.
func (c HexCoord) CalculateFov(cone AngularInterval, maxR int, obstruct func(HexCoord) bool, addLight func(HexCoord, AngularInterval)) {
	epsilon := 0.000001

	for _, section := range OrderedDirections {
		restricted := NewAngularSextant(section).Intersection(cone)

		if restricted.Empty {
			continue
		}

		currentGen := []AngularInterval{restricted}

		for r := 0; maxR < 0 || r <= maxR; r++ {
			var nextGen []AngularInterval

			if len(currentGen) == 0 {
				break
			}

			for _, n := range currentGen {
				outputs := []AngularInterval{n}

				visit := func(p HexCoord) {
					realp := p.AddDelta(c)

					addLight(realp, n)

					if !obstruct(realp) {
						return
					}

					if len(outputs) < 1 {
						return
					}

					ba := p.ExtremeAngles()
					prefix := outputs[:len(outputs)-1]
					lastOutput := outputs[len(outputs)-1]
					containsStart := lastOutput.Contains(ba.Rad0)
					containsEnd := lastOutput.Contains(ba.Rad1)

					switch {
					case containsStart && containsEnd:
						// Split
						more := []AngularInterval{
							NewAngularInterval(lastOutput.Rad0, ba.Rad0),
							NewAngularInterval(ba.Rad1, lastOutput.Rad1),
						}
						outputs = append(prefix, more...)

					case containsStart:
						// Blocked at the end
						outputs = append(prefix, NewAngularInterval(lastOutput.Rad0, ba.Rad0))

					case containsEnd:
						// Blocked at the beginning
						outputs = append(prefix, NewAngularInterval(ba.Rad1, lastOutput.Rad1))

					default:
						// Fully blocked
						outputs = prefix
					}
				}

				sweepIntervalInSextant(n, section, r, visit)

				nextGen = append(nextGen, outputs...)
			}

			currentGen = nil
			for _, g := range nextGen {
				if g.Size() > epsilon {
					currentGen = append(currentGen, g)
				}
			}
		}
	}
}
