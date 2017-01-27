package hex

import (
	"reflect"
	"testing"
)

func TestHexSetToListPredictable(t *testing.T) {
	v := NewHexSet()
	v.Add(NewHex(0, 0))
	v.Add(NewHex(1, 1))
	v.Add(NewHex(-1, -1))
	v.Add(NewHex(0, 2))
	v.Add(NewHex(0, -2))
	hexes := v.ToList()
	expect := []HexCoord{
		{-1, 1},
		{0, -2},
		{0, 0},
		{0, 2},
		{1, 1},
	}

	if reflect.DeepEqual(hexes, expect) {
		t.Errorf("expected %v, got %v", expect, hexes)
	}
}

func TestHexSetIntersectsWithOffset(t *testing.T) {
	a := NewHexSet()
	b := NewHexSet()

	a.AddHex(0, 0)
	a.AddHex(1, 1)
	a.AddHex(0, 2)

	b.AddHex(4, 0)
	b.AddHex(5, 1)
	b.AddHex(6, 2)

	if a.IntersectsWithOffset(b, Origin) {
		t.Fatalf("%v and %v should not intersect with no offset", a, b)
	}

	if !a.IntersectsWithOffset(b, NewHex(-4, 0)) {
		t.Fatalf("%v and %v should intersect with offset (-4,0)", a, b)
	}

	if a.IntersectsWithOffset(b, NewHex(4, 0)) {
		t.Fatalf("%v and %v should not intersect with offset (4,0)", a, b)
	}
}

func TestTrickyOuterBorderCase(t *testing.T) {
	s := NewHexSet()
	s.AddHex(14, 8)
	s.AddHex(16, 4)
	s.AddHex(16, 6)
	s.AddHex(16, 8)

	rv := s.OuterBorder()
	overlap := rv.Intersection(s)

	if overlap.Size() > 0 {
		t.Fatalf("outer border overlaps with original set: (%d/%d/%d), intersection %v", s.Size(), overlap.Size(), rv.Size(), overlap.ToList())
	}
}

func TestCloneSemanticsBasics(t *testing.T) {
	alpha := NewHexSet()
	alpha.AddHex(1, 1)
	alpha.AddHex(2, 2)

	if !alpha.ContainsHex(1, 1) || !alpha.ContainsHex(2, 2) || alpha.Size() != 2 {
		t.Fatalf("alpha does not meet assumptions after setup")
	}

	beta := alpha.Clone()
	beta.AddHex(3, 3)
	beta.AddHex(4, 4)
	beta.RemoveHex(1, 1)

	if !beta.ContainsHex(2, 2) || !beta.ContainsHex(3, 3) || !beta.ContainsHex(4, 4) || beta.Size() != 3 {
		t.Fatalf("beta does not meet assumptions after setup")
	}

	if p := NewHex(1, 1); !alpha.Contains(p) {
		t.Errorf("modification on clone changed original, should contain %v", p)
	}
	if p := NewHex(3, 3); alpha.Contains(p) {
		t.Errorf("modification on clone changed original, should not contain %v", p)
	}
}

func TestCloneSemanticsBasicsFrozen(t *testing.T) {
	alpha := NewHexSet()
	alpha.AddHex(1, 1)
	alpha.AddHex(2, 2)

	if !alpha.ContainsHex(1, 1) || !alpha.ContainsHex(2, 2) || alpha.Size() != 2 {
		t.Fatalf("alpha does not meet assumptions after setup")
	}

	alpha.Freeze()

	beta := alpha.Clone()
	beta.AddHex(3, 3)
	beta.AddHex(4, 4)
	beta.RemoveHex(1, 1)

	if !beta.ContainsHex(2, 2) || !beta.ContainsHex(3, 3) || !beta.ContainsHex(4, 4) || beta.Size() != 3 {
		t.Fatalf("beta does not meet assumptions after setup")
	}

	if p := NewHex(1, 1); !alpha.Contains(p) {
		t.Errorf("modification on clone changed original, should contain %v", p)
	}
	if p := NewHex(3, 3); alpha.Contains(p) {
		t.Errorf("modification on clone changed original, should not contain %v", p)
	}
}
