package term2go

import (
	"context"
	"sync"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// ScreenContents
// ============================================================================

// ScreenContents wraps a GetBufferResponse with convenience accessors.  It
// represents the visible region of a terminal session at a point in time.
type ScreenContents struct {
	raw *iterm2.GetBufferResponse
}

// NewScreenContents creates a ScreenContents from a proto response.
func NewScreenContents(raw *iterm2.GetBufferResponse) *ScreenContents {
	return &ScreenContents{raw: raw}
}

// Cursor returns the cursor position, or nil.
func (s *ScreenContents) Cursor() *Coord {
	if s.raw == nil {
		return nil
	}
	c := s.raw.GetCursor()
	if c == nil {
		return nil
	}
	return &Coord{X: c.GetX(), Y: int32(c.GetY())}
}

// LineCount returns the number of lines.
func (s *ScreenContents) LineCount() int {
	if s.raw == nil {
		return 0
	}
	return len(s.raw.GetContents())
}

// Lines returns all lines as LineContent wrappers.
func (s *ScreenContents) Lines() []*LineContent {
	if s.raw == nil {
		return nil
	}
	rawLines := s.raw.GetContents()
	result := make([]*LineContent, len(rawLines))
	for i, rl := range rawLines {
		result[i] = newLineContent(rl)
	}
	return result
}

// Raw returns the underlying proto response.
func (s *ScreenContents) Raw() *iterm2.GetBufferResponse { return s.raw }

// ============================================================================
// LineContent
// ============================================================================

// LineContent wraps a LineContents proto, pre-computing per-cell offsets so
// callers can do random-access lookups by column.
type LineContent struct {
	raw     *iterm2.LineContents
	text    string
	hardEOL bool
	cells   []cellInfo
}

type cellInfo struct {
	start  int        // byte offset into text
	length int        // byte length of this cell's rune(s) in text
	style  *CellStyle // nil if no style for this cell
}

func newLineContent(raw *iterm2.LineContents) *LineContent {
	if raw == nil {
		return nil
	}
	text := raw.GetText()
	codes := raw.GetCodePointsPerCell()
	styles := expandStyles(raw.GetStyle())

	// Pre-compute cell boundaries from code_points_per_cell RLE.
	totalCells := 0
	for _, cpc := range codes {
		totalCells += int(cpc.GetRepeats())
	}

	cells := make([]cellInfo, 0, totalCells)
	pos := 0
	for _, cpc := range codes {
		n := int(cpc.GetNumCodePoints())
		repeats := int(cpc.GetRepeats())
		for i := 0; i < repeats; i++ {
			cellLen := 0
			for j := 0; j < n && pos+cellLen < len(text); j++ {
				_, size := utf8.DecodeRuneInString(text[pos+cellLen:])
				cellLen += size
			}
			ci := cellInfo{start: pos, length: cellLen}
			pos += cellLen
			cells = append(cells, ci)
		}
	}

	// Assign styles. Each style entry in the expanded list covers one cell
	// (RLE is already expanded by expandStyles).
	for i := 0; i < len(cells) && i < len(styles); i++ {
		cells[i].style = styles[i]
	}

	return &LineContent{
		raw:     raw,
		text:    text,
		hardEOL: raw.GetContinuation() == iterm2.LineContents_CONTINUATION_HARD_EOL,
		cells:   cells,
	}
}

// Text returns the raw text content of the line.
func (l *LineContent) Text() string {
	if l == nil {
		return ""
	}
	return l.text
}

// RuneAt returns the first rune at column col and its byte length, or (0,0).
func (l *LineContent) RuneAt(col int) (rune, int) {
	if l == nil || col < 0 || col >= len(l.cells) {
		return 0, 0
	}
	cell := l.cells[col]
	if cell.start+cell.length > len(l.text) {
		return 0, 0
	}
	r, s := utf8.DecodeRuneInString(l.text[cell.start:])
	return r, s
}

// StyleAt returns the cell style at column col, or nil if no style info
// or col is out of bounds.
func (l *LineContent) StyleAt(col int) *CellStyle {
	if l == nil || col < 0 || col >= len(l.cells) {
		return nil
	}
	return l.cells[col].style
}

// Len returns the number of cells in this line.
func (l *LineContent) Len() int {
	if l == nil {
		return 0
	}
	return len(l.cells)
}

// HardEOL reports whether the line ends with a hard (explicit) newline.
func (l *LineContent) HardEOL() bool {
	if l == nil {
		return false
	}
	return l.hardEOL
}

// expandStyles expands the RLE-encoded style list into per-cell entries.
func expandStyles(styles []*iterm2.CellStyle) []*CellStyle {
	var result []*CellStyle
	for _, s := range styles {
		repeats := int(s.GetRepeats())
		if repeats == 0 {
			repeats = 1
		}
		cs := &CellStyle{raw: s}
		for i := 0; i < repeats; i++ {
			result = append(result, cs)
		}
	}
	return result
}

// ============================================================================
// CellStyle
// ============================================================================

// CellStyle wraps a proto CellStyle with convenience accessors.
type CellStyle struct {
	raw *iterm2.CellStyle
}

// --- text attributes ---------------------------------------------------------

func (c *CellStyle) Bold() bool          { return c.raw.GetBold() }
func (c *CellStyle) Faint() bool         { return c.raw.GetFaint() }
func (c *CellStyle) Italic() bool        { return c.raw.GetItalic() }
func (c *CellStyle) Blink() bool         { return c.raw.GetBlink() }
func (c *CellStyle) Underline() bool     { return c.raw.GetUnderline() }
func (c *CellStyle) Strikethrough() bool { return c.raw.GetStrikethrough() }
func (c *CellStyle) Invisible() bool     { return c.raw.GetInvisible() }
func (c *CellStyle) Inverse() bool       { return c.raw.GetInverse() }
func (c *CellStyle) Guarded() bool       { return c.raw.GetGuarded() }

// Image returns the image placeholder type.
func (c *CellStyle) Image() iterm2.ImagePlaceholderType {
	return c.raw.GetImage()
}

// UnderlineRGB returns the underline color if set.
func (c *CellStyle) UnderlineRGB() (*iterm2.RGBColor, bool) {
	uc := c.raw.GetUnderlineColor()
	return uc, uc != nil
}

// URL returns the URL and identifier if set.
func (c *CellStyle) URL() (url, identifier string, ok bool) {
	u := c.raw.GetUrl()
	if u == nil {
		return "", "", false
	}
	return u.GetUrl(), u.GetIdentifier(), true
}

// BlockID returns the block ID string.
func (c *CellStyle) BlockID() string { return c.raw.GetBlockID() }

// --- foreground color --------------------------------------------------------

// FGStandard returns the standard (palette-indexed) foreground color.
func (c *CellStyle) FGStandard() (uint32, bool) {
	v, ok := c.raw.FgColor.(*iterm2.CellStyle_FgStandard)
	if !ok {
		return 0, false
	}
	return v.FgStandard, true
}

// FGAlternate returns the alternate (semantic) foreground color.
func (c *CellStyle) FGAlternate() (iterm2.AlternateColor, bool) {
	v, ok := c.raw.FgColor.(*iterm2.CellStyle_FgAlternate)
	if !ok {
		return 0, false
	}
	return v.FgAlternate, true
}

// FGRGB returns the RGB foreground color.
func (c *CellStyle) FGRGB() (*iterm2.RGBColor, bool) {
	v, ok := c.raw.FgColor.(*iterm2.CellStyle_FgRgb)
	if !ok {
		return nil, false
	}
	return v.FgRgb, true
}

// FGPlacementX returns the alternate-placement-x foreground value.
func (c *CellStyle) FGPlacementX() (uint32, bool) {
	v, ok := c.raw.FgColor.(*iterm2.CellStyle_FgAlternatePlacementX)
	if !ok {
		return 0, false
	}
	return v.FgAlternatePlacementX, true
}

// HasFG reports whether foreground color is set.
func (c *CellStyle) HasFG() bool { return c.raw.FgColor != nil }

// --- background color --------------------------------------------------------

// BGStandard returns the standard background color.
func (c *CellStyle) BGStandard() (uint32, bool) {
	v, ok := c.raw.BgColor.(*iterm2.CellStyle_BgStandard)
	if !ok {
		return 0, false
	}
	return v.BgStandard, true
}

// BGAlternate returns the alternate background color.
func (c *CellStyle) BGAlternate() (iterm2.AlternateColor, bool) {
	v, ok := c.raw.BgColor.(*iterm2.CellStyle_BgAlternate)
	if !ok {
		return 0, false
	}
	return v.BgAlternate, true
}

// BGRGB returns the RGB background color.
func (c *CellStyle) BGRGB() (*iterm2.RGBColor, bool) {
	v, ok := c.raw.BgColor.(*iterm2.CellStyle_BgRgb)
	if !ok {
		return nil, false
	}
	return v.BgRgb, true
}

// BGPlacementY returns the alternate-placement-y background value.
func (c *CellStyle) BGPlacementY() (uint32, bool) {
	v, ok := c.raw.BgColor.(*iterm2.CellStyle_BgAlternatePlacementY)
	if !ok {
		return 0, false
	}
	return v.BgAlternatePlacementY, true
}

// HasBG reports whether background color is set.
func (c *CellStyle) HasBG() bool { return c.raw.BgColor != nil }

// ============================================================================
// ScreenStreamer
// ============================================================================

// ScreenStreamer streams terminal screen contents on each update.
// Create one with NewScreenStreamer, iterate over Chan(), and call Close()
// when finished.
type ScreenStreamer struct {
	conn            *Connection
	sessionID       string
	ch              chan *ScreenContents
	done            chan struct{}
	token           NotificationToken
	once            sync.Once
	dispatchHandler NotificationHandler // registered on conn to call Dispatch
	// notify is a signal channel that the notification callback uses to
	// wake the run goroutine. The callback MUST NOT call GetBuffer directly
	// because it runs inside dispatchLoop; calling GetBuffer would block
	// dispatchLoop waiting for a response that it needs to deliver.
	notify chan struct{}
}

// NewScreenStreamer subscribes to screen-update notifications for sessionID
// and starts fetching screen contents on each update.
//
// Usage:
//
//	s, err := NewScreenStreamer(conn, sessionID)
//	if err != nil { ... }
//	defer s.Close()
//	for sc := range s.Chan() {
//	    for _, line := range sc.Lines() { ... }
//	}
func NewScreenStreamer(conn *Connection, sessionID string) (*ScreenStreamer, error) {
	s := &ScreenStreamer{
		conn:      conn,
		sessionID: sessionID,
		ch:        make(chan *ScreenContents, 8),
		done:      make(chan struct{}),
		notify:    make(chan struct{}, 1),
	}

	// The callback only signals; GetBuffer happens in the run goroutine so
	// we don't deadlock inside dispatchLoop.
	token, err := SubscribeScreenUpdate(context.Background(), conn, conn, func(caller Caller, n *iterm2.ScreenUpdateNotification) {
		select {
		case s.notify <- struct{}{}:
		default:
		}
	}, sessionID)
	if err != nil {
		return nil, err
	}
	s.token = token

	// Register a Dispatch handler on the Connection.  SubscribeScreenUpdate
	// stores the callback in conn.notifyMap, but dispatchLoop only calls
	// handlers registered via RegisterHandler — we need a bridge.
	s.dispatchHandler = func(msg *iterm2.ServerOriginatedMessage) bool {
		conn.Dispatch(msg)
		return false
	}
	conn.RegisterHandler(s.dispatchHandler)

	// Auto-close the streamer when the connection is lost.
	conn.OnDisconnect(func() { s.Close() })

	go s.run()
	return s, nil
}

func (s *ScreenStreamer) run() {
	defer close(s.ch)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-s.done
		cancel()
	}()
	for {
		select {
		case <-s.notify:
			buf, err := GetBuffer(ctx, s.conn, s.sessionID, &iterm2.LineRange{
				ScreenContentsOnly: proto.Bool(true),
			})
			if err != nil {
				continue
			}
			sc := NewScreenContents(buf)
			select {
			case s.ch <- sc:
			case <-s.done:
				return
			}
		case <-s.done:
			return
		}
	}
}

// Chan returns a receive-only channel of screen contents.
// The channel is closed when the streamer is closed.
func (s *ScreenStreamer) Chan() <-chan *ScreenContents { return s.ch }

// Close stops the streamer and unsubscribes from notifications.
// It is safe to call multiple times.
func (s *ScreenStreamer) Close() {
	s.once.Do(func() {
		close(s.done)
		s.conn.Unsubscribe(s.token)
		if s.dispatchHandler != nil {
			s.conn.UnregisterHandler(s.dispatchHandler)
		}
	})
}

// ============================================================================
// EnumerateRanges — selection utility
// ============================================================================

// EnumerateRanges iterates over a selected range, calling fn for each
// line-contiguous sub-selection.
func EnumerateRanges(
	sel *iterm2.Selection,
	fn func(start, end Coord) error,
) error {
	for _, sub := range sel.GetSubSelections() {
		wcr := sub.GetWindowedCoordRange()
		if wcr == nil {
			continue
		}
		r := wcr.GetCoordRange()
		if r == nil {
			continue
		}
		start := CoordFromProto(r.GetStart())
		end := CoordFromProto(r.GetEnd())
		if err := fn(start, end); err != nil {
			return err
		}
	}
	return nil
}
