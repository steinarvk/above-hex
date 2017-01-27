package hex

import (
	"fmt"

	pb "github.com/steinarvk/above-hex/hexpb"
)

// HexDir represents one of the six directions on a hex grid.
type HexDir int32

const (
	North     HexDir = iota
	Northwest        = iota
	Southwest        = iota
	South            = iota
	Southeast        = iota
	Northeast        = iota
)

var (
	OrderedDirections = []HexDir{
		North, Northwest, Southwest, South, Southeast, Northeast,
	}

	OrthogonalCCW = map[HexDir]HexDir{
		North:     Southwest,
		Northwest: South,
		Southwest: Southeast,
		South:     Northeast,
		Southeast: North,
		Northeast: Northwest,
	}

	Directions = map[HexDir]HexCoord{
		North:     HexCoord{0, 2},
		Northwest: HexCoord{-1, 1},
		Southwest: HexCoord{-1, -1},
		South:     HexCoord{0, -2},
		Southeast: HexCoord{1, -1},
		Northeast: HexCoord{1, 1},
	}

	Origin = HexCoord{0, 0}
)

// DirectionToProto converts a HexDir to a pb.Direction.
func DirectionToProto(d HexDir) pb.Direction {
	switch d {
	case North:
		return pb.Direction_NORTH
	case South:
		return pb.Direction_SOUTH
	case Northwest:
		return pb.Direction_NORTHWEST
	case Southwest:
		return pb.Direction_SOUTHWEST
	case Northeast:
		return pb.Direction_NORTHEAST
	case Southeast:
		return pb.Direction_SOUTHEAST
	default:
		panic(fmt.Errorf("unknown direction from proto: %v", d))
	}
}

// DirectionFromProto converts a pb.Direction to a HexDir.
func DirectionFromProto(d pb.Direction) HexDir {
	switch d {
	case pb.Direction_NORTH:
		return North
	case pb.Direction_SOUTH:
		return South
	case pb.Direction_NORTHWEST:
		return Northwest
	case pb.Direction_SOUTHWEST:
		return Southwest
	case pb.Direction_NORTHEAST:
		return Northeast
	case pb.Direction_SOUTHEAST:
		return Southeast
	default:
		panic(fmt.Errorf("unknown direction to proto: %v", d))
	}
}
