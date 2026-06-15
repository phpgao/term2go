package term2go

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Commonly used iTerm2 profile property keys.
// These match the human-readable key names used internally by iTerm2.
const (
	ProfileKeyForegroundColor         = "Foreground Color"
	ProfileKeyBackgroundColor         = "Background Color"
	ProfileKeyBoldColor               = "Bold Color"
	ProfileKeyLinkColor               = "Link Color"
	ProfileKeySelectionColor          = "Selection Color"
	ProfileKeySelectedTextColor       = "Selected Text Color"
	ProfileKeyCursorColor             = "Cursor Color"
	ProfileKeyCursorTextColor         = "Cursor Text Color"
	ProfileKeyBadgeText               = "Badge Text"
	ProfileKeyUseBrightBold           = "Use Bright Bold"
	ProfileKeyCursorType              = "Cursor Type"
	ProfileKeyBlinkingCursor          = "Blinking Cursor"
	ProfileKeyUseLigatures            = "Use Ligatures"
	ProfileKeyCustomCommand           = "Custom Command"
	ProfileKeyCommand                 = "Command"
	ProfileKeyFont                    = "Font"
	ProfileKeyTransparency            = "Transparency"
	ProfileKeyBlend                   = "Blend"
	ProfileKeyBlur                    = "Blur"
	ProfileKeyBackgroundImageLocation = "Background Image Location"
	ProfileKeyBoldIsBright            = "Bold Brightens Color"
	ProfileKeyScrollbackLines         = "Scrollback Lines"
	ProfileKeyUnlimitedScrollback     = "Unlimited Scrollback"
	ProfileKeySilenceBell             = "Silence Bell"
	ProfileKeyColumns                 = "Columns"
	ProfileKeyRows                    = "Rows"
	// ANSI color keys
	ProfileKeyAnsi0Color  = "Ansi 0 Color"
	ProfileKeyAnsi1Color  = "Ansi 1 Color"
	ProfileKeyAnsi2Color  = "Ansi 2 Color"
	ProfileKeyAnsi3Color  = "Ansi 3 Color"
	ProfileKeyAnsi4Color  = "Ansi 4 Color"
	ProfileKeyAnsi5Color  = "Ansi 5 Color"
	ProfileKeyAnsi6Color  = "Ansi 6 Color"
	ProfileKeyAnsi7Color  = "Ansi 7 Color"
	ProfileKeyAnsi8Color  = "Ansi 8 Color"
	ProfileKeyAnsi9Color  = "Ansi 9 Color"
	ProfileKeyAnsi10Color = "Ansi 10 Color"
	ProfileKeyAnsi11Color = "Ansi 11 Color"
	ProfileKeyAnsi12Color = "Ansi 12 Color"
	ProfileKeyAnsi13Color = "Ansi 13 Color"
	ProfileKeyAnsi14Color = "Ansi 14 Color"
	ProfileKeyAnsi15Color = "Ansi 15 Color"
	// Font keys
	ProfileKeyFontAntialias   = "ASCII Anti Aliased"
	ProfileKeyFontWeight      = "Normal Font Weight"
	ProfileKeyBoldFontWeight  = "Bold Font Weight"
	ProfileKeyUseNonAsciiFont = "Use a different font for non-ASCII text"
	ProfileKeyNonAsciiFont    = "Non-ASCII Font"
	ProfileKeyUseItalicFont   = "Use Italic Font"
	ProfileKeyUseThinStrokes  = "Thin Strokes"
	// Background image keys
	ProfileKeyBackgroundImageTiling    = "Background Image Mode"
	ProfileKeyBackgroundImageAlignment = "Background Image Alignment"
	ProfileKeyBackgroundImageBehavior  = "Background Image Behavior"
)

// CustomCommandEnabled is the value to use with ProfileKeyCustomCommand when enabled.
const CustomCommandEnabled = "Yes"

// LocalWriteOnlyProfile holds profile property changes that affect only
// the current session, without modifying the underlying profile.
type LocalWriteOnlyProfile struct {
	// values maps profile key → JSON-encoded value (as stored in proto)
	values map[string]string
}

// NewLocalWriteOnlyProfile creates an empty write-only profile.
func NewLocalWriteOnlyProfile() *LocalWriteOnlyProfile {
	return &LocalWriteOnlyProfile{values: make(map[string]string)}
}

// Values returns the raw key→JSON-value map for RPC use.
func (p *LocalWriteOnlyProfile) Values() map[string]string {
	return p.values
}

// Assignments returns the profile property assignments for the RPC.
func (p *LocalWriteOnlyProfile) Assignments() []*iterm2.SetProfilePropertyRequest_Assignment {
	var result []*iterm2.SetProfilePropertyRequest_Assignment
	for k, v := range p.values {
		result = append(result, &iterm2.SetProfilePropertyRequest_Assignment{
			Key:       proto.String(k),
			JsonValue: proto.String(v),
		})
	}
	return result
}

func (p *LocalWriteOnlyProfile) set(key, jsonValue string) {
	p.values[key] = jsonValue
}

// SetForegroundColor sets the foreground text color.
func (p *LocalWriteOnlyProfile) SetForegroundColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyForegroundColor, c.ToProfileJSON())
	}
	return p
}

// SetBackgroundColor sets the background color.
func (p *LocalWriteOnlyProfile) SetBackgroundColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyBackgroundColor, c.ToProfileJSON())
	}
	return p
}

// SetBoldColor sets the bold text color.
func (p *LocalWriteOnlyProfile) SetBoldColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyBoldColor, c.ToProfileJSON())
	}
	return p
}

// SetLinkColor sets the link color.
func (p *LocalWriteOnlyProfile) SetLinkColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyLinkColor, c.ToProfileJSON())
	}
	return p
}

// SetSelectionColor sets the selection background color.
func (p *LocalWriteOnlyProfile) SetSelectionColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeySelectionColor, c.ToProfileJSON())
	}
	return p
}

// SetSelectedTextColor sets the selected text color.
func (p *LocalWriteOnlyProfile) SetSelectedTextColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeySelectedTextColor, c.ToProfileJSON())
	}
	return p
}

// SetCursorColor sets the cursor color.
func (p *LocalWriteOnlyProfile) SetCursorColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyCursorColor, c.ToProfileJSON())
	}
	return p
}

// SetCursorTextColor sets the cursor text color.
func (p *LocalWriteOnlyProfile) SetCursorTextColor(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyCursorTextColor, c.ToProfileJSON())
	}
	return p
}

func jsonBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// SetUseBrightBold controls whether bold text is rendered with bright colors.
func (p *LocalWriteOnlyProfile) SetUseBrightBold(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUseBrightBold, jsonBool(v))
	return p
}

// SetBoldIsBright controls whether bold text is Brightened.
func (p *LocalWriteOnlyProfile) SetBoldIsBright(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyBoldIsBright, jsonBool(v))
	return p
}

// SetBlinkingCursor controls cursor blink.
func (p *LocalWriteOnlyProfile) SetBlinkingCursor(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyBlinkingCursor, jsonBool(v))
	return p
}

// SetUseLigatures controls font ligatures.
func (p *LocalWriteOnlyProfile) SetUseLigatures(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUseLigatures, jsonBool(v))
	return p
}

// SetSilenceBell controls the terminal bell.
func (p *LocalWriteOnlyProfile) SetSilenceBell(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeySilenceBell, jsonBool(v))
	return p
}

// SetUnlimitedScrollback enables or disables unlimited scrollback.
func (p *LocalWriteOnlyProfile) SetUnlimitedScrollback(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUnlimitedScrollback, jsonBool(v))
	return p
}

// SetScrollbackLines sets the max scrollback lines.
func (p *LocalWriteOnlyProfile) SetScrollbackLines(n int) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(n)
	p.set(ProfileKeyScrollbackLines, string(v))
	return p
}

// SetTransparency sets the transparency (0–1).
func (p *LocalWriteOnlyProfile) SetTransparency(v float64) *LocalWriteOnlyProfile {
	b, _ := json.Marshal(v)
	p.set(ProfileKeyTransparency, string(b))
	return p
}

// SetBlend sets the background image blend (0–1).
func (p *LocalWriteOnlyProfile) SetBlend(v float64) *LocalWriteOnlyProfile {
	b, _ := json.Marshal(v)
	p.set(ProfileKeyBlend, string(b))
	return p
}

// SetBlur sets the background blur (0–1).
func (p *LocalWriteOnlyProfile) SetBlur(v float64) *LocalWriteOnlyProfile {
	b, _ := json.Marshal(v)
	p.set(ProfileKeyBlur, string(b))
	return p
}

// SetFont sets the font name.
func (p *LocalWriteOnlyProfile) SetFont(name string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(name)
	p.set(ProfileKeyFont, string(v))
	return p
}

// SetCustomCommand enables custom command mode and sets the command.
func (p *LocalWriteOnlyProfile) SetCustomCommand(cmd string) *LocalWriteOnlyProfile {
	p.set(ProfileKeyCustomCommand, jsonBool(true))
	v, _ := json.Marshal(cmd)
	p.set(ProfileKeyCommand, string(v))
	return p
}

// SetCommand sets the command to run.
func (p *LocalWriteOnlyProfile) SetCommand(cmd string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(cmd)
	p.set(ProfileKeyCommand, string(v))
	return p
}

// SetCursorType sets the cursor type. Valid values: "xterm", "block", "underline", "vertical".
func (p *LocalWriteOnlyProfile) SetCursorType(typ string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(typ)
	p.set(ProfileKeyCursorType, string(v))
	return p
}

// SetBadgeText sets the badge text displayed in the session.
func (p *LocalWriteOnlyProfile) SetBadgeText(text string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(text)
	p.set(ProfileKeyBadgeText, string(v))
	return p
}

// SetBackgroundImageLocation sets the background image URL or path.
func (p *LocalWriteOnlyProfile) SetBackgroundImageLocation(path string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(path)
	p.set(ProfileKeyBackgroundImageLocation, string(v))
	return p
}

// SetBackgroundImageTiling sets how the background image tiles (0=stretch, 1=tile, 2=aspect fill, 3=aspect fit).
func (p *LocalWriteOnlyProfile) SetBackgroundImageTiling(mode int) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(mode)
	p.set(ProfileKeyBackgroundImageTiling, string(v))
	return p
}

// SetBackgroundImageAlignment sets the background image alignment.
func (p *LocalWriteOnlyProfile) SetBackgroundImageAlignment(alignment string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(alignment)
	p.set(ProfileKeyBackgroundImageAlignment, string(v))
	return p
}

// SetBackgroundImageBehavior sets the background image behavior.
func (p *LocalWriteOnlyProfile) SetBackgroundImageBehavior(behavior string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(behavior)
	p.set(ProfileKeyBackgroundImageBehavior, string(v))
	return p
}

// SetFontAntialias controls font anti-aliasing.
func (p *LocalWriteOnlyProfile) SetFontAntialias(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyFontAntialias, jsonBool(v))
	return p
}

// SetFontWeight sets the normal font weight (e.g., "normal", "bold").
func (p *LocalWriteOnlyProfile) SetFontWeight(w string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(w)
	p.set(ProfileKeyFontWeight, string(v))
	return p
}

// SetBoldFontWeight sets the bold font weight.
func (p *LocalWriteOnlyProfile) SetBoldFontWeight(w string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(w)
	p.set(ProfileKeyBoldFontWeight, string(v))
	return p
}

// SetUseNonAsciiFont controls whether a separate non-ASCII font is used.
func (p *LocalWriteOnlyProfile) SetUseNonAsciiFont(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUseNonAsciiFont, jsonBool(v))
	return p
}

// SetNonAsciiFont sets the non-ASCII font name.
func (p *LocalWriteOnlyProfile) SetNonAsciiFont(name string) *LocalWriteOnlyProfile {
	v, _ := json.Marshal(name)
	p.set(ProfileKeyNonAsciiFont, string(v))
	return p
}

// SetUseItalicFont controls italic font usage.
func (p *LocalWriteOnlyProfile) SetUseItalicFont(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUseItalicFont, jsonBool(v))
	return p
}

// SetUseThinStrokes controls thin stroke rendering.
func (p *LocalWriteOnlyProfile) SetUseThinStrokes(v bool) *LocalWriteOnlyProfile {
	p.set(ProfileKeyUseThinStrokes, jsonBool(v))
	return p
}

// ============================================================================
// ANSI Color Setters
// ============================================================================

// SetAnsi0Color sets the ANSI 0 (Black) color.
func (p *LocalWriteOnlyProfile) SetAnsi0Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi0Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi1Color sets the ANSI 1 (Red) color.
func (p *LocalWriteOnlyProfile) SetAnsi1Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi1Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi2Color sets the ANSI 2 (Green) color.
func (p *LocalWriteOnlyProfile) SetAnsi2Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi2Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi3Color sets the ANSI 3 (Yellow) color.
func (p *LocalWriteOnlyProfile) SetAnsi3Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi3Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi4Color sets the ANSI 4 (Blue) color.
func (p *LocalWriteOnlyProfile) SetAnsi4Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi4Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi5Color sets the ANSI 5 (Magenta) color.
func (p *LocalWriteOnlyProfile) SetAnsi5Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi5Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi6Color sets the ANSI 6 (Cyan) color.
func (p *LocalWriteOnlyProfile) SetAnsi6Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi6Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi7Color sets the ANSI 7 (White) color.
func (p *LocalWriteOnlyProfile) SetAnsi7Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi7Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi8Color sets the ANSI 8 (Bright Black) color.
func (p *LocalWriteOnlyProfile) SetAnsi8Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi8Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi9Color sets the ANSI 9 (Bright Red) color.
func (p *LocalWriteOnlyProfile) SetAnsi9Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi9Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi10Color sets the ANSI 10 (Bright Green) color.
func (p *LocalWriteOnlyProfile) SetAnsi10Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi10Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi11Color sets the ANSI 11 (Bright Yellow) color.
func (p *LocalWriteOnlyProfile) SetAnsi11Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi11Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi12Color sets the ANSI 12 (Bright Blue) color.
func (p *LocalWriteOnlyProfile) SetAnsi12Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi12Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi13Color sets the ANSI 13 (Bright Magenta) color.
func (p *LocalWriteOnlyProfile) SetAnsi13Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi13Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi14Color sets the ANSI 14 (Bright Cyan) color.
func (p *LocalWriteOnlyProfile) SetAnsi14Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi14Color, c.ToProfileJSON())
	}
	return p
}

// SetAnsi15Color sets the ANSI 15 (Bright White) color.
func (p *LocalWriteOnlyProfile) SetAnsi15Color(c *Color) *LocalWriteOnlyProfile {
	if c != nil {
		p.set(ProfileKeyAnsi15Color, c.ToProfileJSON())
	}
	return p
}

// SetProfileProperties applies LocalWriteOnlyProfile changes to the session.
func (s *Session) SetProfileProperties(ctx context.Context, p *LocalWriteOnlyProfile) error {
	req := &iterm2.SetProfilePropertyRequest{
		Target: &iterm2.SetProfilePropertyRequest_Session{
			Session: s.ID,
		},
		Assignments: p.Assignments(),
	}
	msg := newRequest()
	msg.Submessage = &iterm2.ClientOriginatedMessage_SetProfilePropertyRequest{
		SetProfilePropertyRequest: req,
	}
	resp, err := s.caller.Call(ctx, msg)
	if err != nil {
		return fmt.Errorf("set profile properties: %w", err)
	}
	return checkError(resp)
}

// SetProfileProperty sets a single profile property on this session.
func (s *Session) SetProfileProperty(ctx context.Context, key, jsonValue string) error {
	return SetProfileProperty(ctx, s.caller, s.ID, key, jsonValue)
}

// GetProfile retrieves all profile properties for this session.
func (s *Session) GetProfile(ctx context.Context) (map[string]string, error) {
	resp, err := GetProfileProperty(ctx, s.caller, s.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	result := make(map[string]string, len(resp.GetProperties()))
	for _, prop := range resp.GetProperties() {
		result[prop.GetKey()] = prop.GetJsonValue()
	}
	return result, nil
}

// GetProfileProperty retrieves a specific profile property.
func (s *Session) GetProfileProperty(ctx context.Context, key string) (string, error) {
	resp, err := GetProfileProperty(ctx, s.caller, s.ID, []string{key})
	if err != nil {
		return "", fmt.Errorf("get profile property %s: %w", key, err)
	}
	props := resp.GetProperties()
	if len(props) == 0 {
		return "", nil
	}
	return props[0].GetJsonValue(), nil
}
