package term2go

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestCoord tests Coord struct basic properties.
func TestCoord(t *testing.T) {
	c := Coord{X: 10, Y: 20}
	assert.Equal(t, int32(10), c.X)
	assert.Equal(t, int32(20), c.Y)
}

// TestCoordRange tests CoordRange struct basic properties.
func TestCoordRange(t *testing.T) {
	cr := CoordRange{
		Start: Coord{X: 1, Y: 2},
		End:   Coord{X: 3, Y: 4},
	}
	assert.Equal(t, Coord{X: 1, Y: 2}, cr.Start)
	assert.Equal(t, Coord{X: 3, Y: 4}, cr.End)
}

// TestCoordFromProto tests CoordFromProto function.
func TestCoordFromProto(t *testing.T) {
	pc := &iterm2.Coord{X: proto.Int32(5), Y: proto.Int64(10)}
	c := CoordFromProto(pc)
	assert.Equal(t, Coord{X: 5, Y: 10}, c)
}

// TestCoordFromProto_Nil tests CoordFromProto handling nil input.
func TestCoordFromProto_Nil(t *testing.T) {
	c := CoordFromProto(nil)
	assert.Equal(t, Coord{}, c)
}

// TestCoordRangeFromProto tests CoordRangeFromProto function.
func TestCoordRangeFromProto(t *testing.T) {
	pcr := &iterm2.CoordRange{
		Start: &iterm2.Coord{X: proto.Int32(1), Y: proto.Int64(2)},
		End:   &iterm2.Coord{X: proto.Int32(3), Y: proto.Int64(4)},
	}
	cr := CoordRangeFromProto(pcr)
	assert.Equal(t, Coord{X: 1, Y: 2}, cr.Start)
	assert.Equal(t, Coord{X: 3, Y: 4}, cr.End)
}

// TestCoordRangeFromProto_Nil tests CoordRangeFromProto handling nil input.
func TestCoordRangeFromProto_Nil(t *testing.T) {
	cr := CoordRangeFromProto(nil)
	assert.Equal(t, CoordRange{}, cr)
}

// TestPoint tests Point struct basic properties.
func TestPoint(t *testing.T) {
	p := Point{X: 100, Y: 200}
	assert.Equal(t, Point{X: 100, Y: 200}, p)
}

// TestSize tests Size struct basic properties.
func TestSize(t *testing.T) {
	s := Size{Width: 800, Height: 600}
	assert.Equal(t, Size{Width: 800, Height: 600}, s)
}

// TestWindowFrame tests WindowFrame struct basic properties.
func TestWindowFrame(t *testing.T) {
	wf := WindowFrame{
		Origin: Point{X: 10, Y: 20},
		Size:   Size{Width: 640, Height: 480},
	}
	assert.Equal(t, Point{X: 10, Y: 20}, wf.Origin)
	assert.Equal(t, Size{Width: 640, Height: 480}, wf.Size)
}
