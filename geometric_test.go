package hex

import "testing"
import "math"

func TestAngularIntervalSize(t *testing.T) {
	if NewAngularInterval(1, 5).Size() != 4.0 {
		t.Errorf("expected [1,5] to have size 4")
	}

	s := NewAngularInterval(5, 1).Size()
	if s < 2.27 || s > 2.29 {
		t.Errorf("expected [5,1] to have size 2.28.., got: %v", s)
	}
}

func TestAngularIntervalIntersection(t *testing.T) {
	tolerance := 0.0001
	checkSize := func(a, b AngularInterval, expect float64) {
		got := a.Intersection(b)
		if math.Abs(got.Size()-expect) > tolerance {
			t.Errorf("expected %v intersect %v to have size %v, got %v (size %v)", a, b, expect, got, got.Size())
		}
	}
	checkContains := func(a, b AngularInterval, expect float64) {
		got := a.Intersection(b)
		if !got.Contains(expect) {
			t.Errorf("expected %v intersect %v to contain %v, got %v", a, b, expect, got)
		}
	}
	checkNotContains := func(a, b AngularInterval, expect float64) {
		got := a.Intersection(b)
		if got.Contains(expect) {
			t.Errorf("expected %v intersect %v to not contain %v, got %v", a, b, expect, got)
		}
	}

	a := NewAngularInterval(1, 5)
	b := NewAngularInterval(2, 4)

	checkSize(a, b, 2.0)
	checkSize(b, a, 2.0)

	c := NewAngularInterval(1, 3)

	checkSize(b, c, 1.0)
	checkContains(b, c, 2.5)
	checkNotContains(b, c, 3.5)

	d := NewAngularInterval(3, 5)
	checkSize(b, d, 1.0)
	checkNotContains(b, d, 2.5)
	checkContains(b, d, 3.5)

	e := NewAngularInterval(4, 2)

	checkContains(a, e, 0.0)
	checkNotContains(a, e, 3.0)
}
