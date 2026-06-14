package term2go

import iterm2 "github.com/phpgao/term2go/proto"

// Coord represents a terminal coordinate (column, line).
type Coord struct {
	X, Y int32
}

// CoordRange represents a range of coordinates.
type CoordRange struct {
	Start, End Coord
}

// CoordFromProto converts a proto Coord to a native Coord.
// Note: proto Y is int64 (line numbers can exceed int32 range for long scrollback),
// but native Coord uses int32 for both fields. Values beyond int32 range are truncated.
func CoordFromProto(c *iterm2.Coord) Coord {
	return Coord{X: c.GetX(), Y: int32(c.GetY())}
}

// CoordRangeFromProto converts a proto CoordRange to a native CoordRange.
func CoordRangeFromProto(cr *iterm2.CoordRange) CoordRange {
	return CoordRange{
		Start: CoordFromProto(cr.GetStart()),
		End:   CoordFromProto(cr.GetEnd()),
	}
}

// Point represents an origin coordinate in iTerm2 (pixels).
type Point struct {
	X, Y int32
}

// Size represents dimensions (width, height).
type Size struct {
	Width, Height int32
}

// WindowFrame stores both origin and size of a window.
type WindowFrame struct {
	Origin Point
	Size   Size
}
