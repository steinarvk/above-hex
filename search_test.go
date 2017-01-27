package hex

import (
	"testing"
)

func TestBasicAStar(t *testing.T) {
	goal := NewHex(0, 4)
	origin := NewHexSetSingleton(NewHex(0, 0))
	isGoal := func(p HexCoord) bool {
		return p == goal
	}
	stepCost := func(a, b HexCoord) (float64, bool) {
		if b.X == 0 && b.Y == 2 {
			return 0, false
		}
		return 2, true
	}
	heuristic := func(p HexCoord) float64 {
		return p.Geo().Sub(goal.Geo()).Length()
	}

	result, err := AStar(&AStarParams{
		Start:     origin,
		IsGoal:    isGoal,
		Cost:      stepCost,
		Heuristic: heuristic,
		MaxCost:   0,
	})

	if err != nil {
		t.Errorf("A* failed: %v", err)
		t.FailNow()
	}

	if len(result.Path) != 4 {
		t.Errorf("expected path of length 4, got: %v", result.Path)
		t.FailNow()
	}

	if result.Path[0] != NewHex(0, 0) {
		t.Errorf("expected to start at (0,0), started at: %v", result.Path[0])
	}

	if result.Path[3] != NewHex(0, 4) {
		t.Errorf("expected to end at (0,4), ended at: %v", result.Path[3])
	}
}

func TestBasicAStarFailing(t *testing.T) {
	goal := NewHex(0, 20)
	origin := NewHexSetSingleton(NewHex(0, 0))
	isGoal := func(p HexCoord) bool {
		return p == goal
	}
	stepCost := func(a, b HexCoord) (float64, bool) {
		if b.Radius() == 5 {
			return 0, false
		}
		return 2, true
	}
	heuristic := func(p HexCoord) float64 {
		return p.Geo().Sub(goal.Geo()).Length()
	}

	_, err := AStar(&AStarParams{
		Start:     origin,
		IsGoal:    isGoal,
		Cost:      stepCost,
		Heuristic: heuristic,
		MaxCost:   0,
	})

	if err == nil {
		t.Errorf("expected failure, got success")
	}
}
