package hex

import (
	"fmt"

	"github.com/oleiade/lane"
)

// AStarParams provides parameters for an A* search.
type AStarParams struct {
	Start     *HexSet
	IsGoal    func(HexCoord) bool
	Cost      func(HexCoord, HexCoord) (float64, bool)
	Heuristic func(HexCoord) float64
	MaxCost   float64
}

// AStarResult represents the result of an A* search.
type AStarResult struct {
	Path []HexCoord
	Cost float64
}

type aStarNode struct {
	point     HexCoord
	cost      float64
	heuristic float64
}

// AStar performs an A* search.
func AStar(params *AStarParams) (*AStarResult, error) {
	closed := NewHexSet()
	open := lane.NewPQueue(lane.MINPQ)
	openMap := map[HexCoord]*aStarNode{}
	trail := map[HexCoord]*HexCoord{}

	insertNode := func(n *aStarNode) {
		k := int(10000 * (n.cost + n.heuristic))
		open.Push(n, k)
		openMap[n.point] = n
	}

	for _, p := range params.Start.ToList() {
		trail[p] = nil
		node := aStarNode{
			point:     p,
			cost:      0,
			heuristic: params.Heuristic(p),
		}
		insertNode(&node)
	}

	for open.Size() > 0 {
		currentNode, _ := open.Pop()
		current := currentNode.(*aStarNode)

		if params.MaxCost != 0 && current.cost > params.MaxCost {
			return nil, fmt.Errorf("no path found within cost limit")
		}

		if params.IsGoal(current.point) {
			var rpath []HexCoord
			for p := &current.point; p != nil; p = trail[*p] {
				rpath = append(rpath, *p)
			}

			result := AStarResult{}
			result.Cost = current.cost
			for i := len(rpath) - 1; i >= 0; i-- {
				result.Path = append(result.Path, rpath[i])
			}

			return &result, nil
		}

		closed.Add(current.point)

		for _, neighbour := range current.point.Neighbours() {
			if closed.Contains(neighbour) {
				continue
			}

			stepCost, ok := params.Cost(current.point, neighbour)
			if !ok {
				continue
			}

			node := aStarNode{
				point: neighbour,
				cost:  current.cost + stepCost,
			}
			prevNode, present := openMap[node.point]
			if present {
				if node.cost >= prevNode.cost {
					continue
				}
				node.heuristic = prevNode.heuristic
			} else {
				node.heuristic = params.Heuristic(node.point)
			}

			insertNode(&node)
			trail[node.point] = &current.point
		}
	}

	return nil, fmt.Errorf("no path found")
}
