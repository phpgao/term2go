package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestCellStyle_Faint tests CellStyle.Faint method for detecting faint text attribute.
func TestCellStyle_Faint(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{Faint: boolPtr(true)}}
	assert.True(t, c.Faint())
	assert.False(t, (&CellStyle{raw: &iterm2.CellStyle{}}).Faint())
}

// TestCellStyle_Blink tests CellStyle.Blink method for detecting blink text attribute.
func TestCellStyle_Blink(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{Blink: boolPtr(true)}}
	assert.True(t, c.Blink())
	assert.False(t, (&CellStyle{raw: &iterm2.CellStyle{}}).Blink())
}

// TestCellStyle_Invisible tests CellStyle.Invisible method for detecting invisible text attribute.
func TestCellStyle_Invisible(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{Invisible: boolPtr(true)}}
	assert.True(t, c.Invisible())
	assert.False(t, (&CellStyle{raw: &iterm2.CellStyle{}}).Invisible())
}

// TestCellStyle_Inverse tests CellStyle.Inverse method for detecting inverse text attribute.
func TestCellStyle_Inverse(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{Inverse: boolPtr(true)}}
	assert.True(t, c.Inverse())
	assert.False(t, (&CellStyle{raw: &iterm2.CellStyle{}}).Inverse())
}

// TestCellStyle_Guarded tests CellStyle.Guarded method for detecting guarded text attribute.
func TestCellStyle_Guarded(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{Guarded: boolPtr(true)}}
	assert.True(t, c.Guarded())
	assert.False(t, (&CellStyle{raw: &iterm2.CellStyle{}}).Guarded())
}

// TestCellStyle_BlockID tests CellStyle.BlockID method for retrieving block ID.
func TestCellStyle_BlockID(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{BlockID: strPtr("my-block")}}
	assert.Equal(t, "my-block", c.BlockID())
	assert.Equal(t, "", (&CellStyle{raw: &iterm2.CellStyle{}}).BlockID())
}

// TestCellStyle_FGAlternate tests CellStyle.FGAlternate method for retrieving foreground alternate color.
func TestCellStyle_FGAlternate(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{}}
	_, ok := c.FGAlternate()
	assert.False(t, ok)
}

// TestCellStyle_FGPlacementX tests CellStyle.FGPlacementX method for retrieving foreground X placement.
func TestCellStyle_FGPlacementX(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{}}
	_, ok := c.FGPlacementX()
	assert.False(t, ok)
}

// TestCellStyle_BGAlternate tests CellStyle.BGAlternate method for retrieving background alternate color.
func TestCellStyle_BGAlternate(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{}}
	_, ok := c.BGAlternate()
	assert.False(t, ok)
}

// TestCellStyle_BGRGB tests CellStyle.BGRGB method for retrieving background RGB color.
func TestCellStyle_BGRGB(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{}}
	_, ok := c.BGRGB()
	assert.False(t, ok)
}

// TestCellStyle_BGPlacementY tests CellStyle.BGPlacementY method for retrieving background Y placement.
func TestCellStyle_BGPlacementY(t *testing.T) {
	c := &CellStyle{raw: &iterm2.CellStyle{}}
	_, ok := c.BGPlacementY()
	assert.False(t, ok)
}

func boolPtr(b bool) *bool { return &b }
