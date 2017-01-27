package hex

import "testing"
import "math"
import "math/rand"

func TestIntervalSweepCoversExtremeRays(t *testing.T) {
	rand.Seed(123)

	for i := 0; i < 100; i++ {
		a := rand.Float64() * math.Pi * 2
		rr := rand.Float64() * 3.0
		r := rand.Int()%100 + 1
		interval := NewAngularInterval(a-rr, a+rr)
		rayA := NewGeoPolar(1.0, interval.Rad0)
		rayB := NewGeoPolar(1.0, interval.Rad1)

		traced := map[HexCoord]bool{}

		for _, d := range OrderedDirections {
			mark := func(p HexCoord) {
				traced[p] = true
			}
			sweepIntervalInSextant(interval, d, r, mark)
		}

		for _, p := range HexCircle(r) {
			onRay := p.ContainsRay(rayA.X, rayA.Y) || p.ContainsRay(rayB.X, rayB.Y)
			if onRay && !traced[p] {
				t.Errorf("expected %v to be traced for %v on hex radius %v, but it was not", p, interval, r)
				t.FailNow()
			}
		}
	}
}

func TestSweepThinConeIsRay(t *testing.T) {
	rand.Seed(123)

	r := 30
	w := 0.00000001

	for i := 0; i < 100; i++ {
		a := rand.Float64() * math.Pi * 2
		interval := NewAngularInterval(a-w, a+w)

		traced := map[HexCoord]bool{}

		for n := 0; n <= r; n++ {
			for _, d := range OrderedDirections {
				mark := func(p HexCoord) {
					traced[p] = true
				}
				sweepIntervalInSextant(interval, d, n, mark)
			}
		}

		for n := 0; n <= r; n++ {
			for _, p := range HexCircle(n) {
				if !p.ExtremeAngles().Contains(a) {
					continue
				}

				if !traced[p] {
					t.Errorf("tracing thin cone %v, expected %v but it was not lit", interval, p)
					t.FailNow()
				}

				delete(traced, p)
			}
		}

		if len(traced) > 0 {
			t.Errorf("tracing thin cone %v, superflous elements were traced: %v", interval, traced)
			for p, _ := range traced {
				t.Errorf("  hex %v: %v", p, p.ExtremeAngles())
			}
		}
	}
}

func TestSimpleFovCalculation(t *testing.T) {
	lit := map[HexCoord]bool{}
	obstruct := func(p HexCoord) bool { return p.X == 0 && p.Y == 2 }
	addLight := func(p HexCoord, _ AngularInterval) {
		lit[p] = true
	}
	Origin.CalculateFov(FullAngularInterval, 10, obstruct, addLight)

	if !lit[Origin] {
		t.Errorf("expected origin to be lit")
	}

	if !lit[NewHex(0, 2)] {
		t.Errorf("expected wall to be lit")
	}

	if lit[NewHex(0, 4)] {
		t.Errorf("expected tile beyond wall to not be lit")
	}

	if !lit[NewHex(0, -4)] {
		t.Errorf("expected south tile to be lit")
	}

	if lit[NewHex(1, 9)] {
		t.Errorf("expected tile diagonally behind wall to not be lit")
	}

	if !lit[NewHex(1, -9)] {
		t.Errorf("expected far south diagonal tile to be lit")
	}
}

func TestFovCannotSeeThroughDiagonalWall(t *testing.T) {
	lit := map[HexCoord]bool{}
	hs := NewHexSet()
	hs.Add(NewHex(1, 1))
	hs.Add(NewHex(0, 2))
	obstruct := func(p HexCoord) bool { return hs.Contains(p) }
	addLight := func(p HexCoord, _ AngularInterval) {
		lit[p] = true
	}
	Origin.CalculateFov(FullAngularInterval, 2, obstruct, addLight)

	if lit[NewHex(1, 3)] {
		t.Errorf("expected tile beyond wall to not be lit")
	}
	if !lit[NewHex(-1, -3)] {
		t.Errorf("expected control tile to be lit")
	}
}

func BenchmarkSimpleFov(b *testing.B) {
	walls := map[HexCoord]bool{}

	r := 40
	torch := NewHex(5, 5)
	for _, p := range HexDisk(r) {
		if p.X == torch.X && p.Y == torch.Y {
			continue
		}
		walls[p] = rand.Float64() > 0.85
	}

	isWall := func(p HexCoord) bool { return walls[p] }
	doNothing := func(p HexCoord, n AngularInterval) {}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		torch.CalculateFov(FullAngularInterval, 10, isWall, doNothing)
	}
}

func TestHoleFillingTrickySet(t *testing.T) {
	s := NewHexSet()
	s.AddHex(14, 8)
	s.AddHex(16, 4)

	rv := HexSetWithHolesFilled(s)
	toomuch := s.Difference(rv)
	if toomuch.Size() > 0 {
		t.Fatalf("set with holes filled does not contain the original set: contains %v, missing %v", rv.ToList(), toomuch.ToList())
	}
}

func TestHoleFillingNonConvexSet(t *testing.T) {
	s := NewHexSetAround(Origin, 1)
	s.RemoveHex(0, 0)

	q := HexSetWithHolesFilled(s)
	if !q.ContainsHex(0, 0) {
		t.Fatalf("hole not filled as expected")
	}

	s.RemoveHex(1, 1)
	r := HexSetWithHolesFilled(s)
	if r.ContainsHex(0, 0) {
		t.Fatalf("non-hole (0,0) was filled: (0,0) is outside set because of (1,1)")
	}
}
