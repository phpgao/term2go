package term2go

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// Color represents a terminal color with optional alpha and color space.
type Color struct {
	Red        float64
	Green      float64
	Blue       float64
	Alpha      float64
	ColorSpace string
}

// NewColor creates a Color with full opacity (alpha=255) in the sRGB color space.
func NewColor(r, g, b float64) *Color {
	return &Color{
		Red:        r,
		Green:      g,
		Blue:       b,
		Alpha:      255,
		ColorSpace: "sRGB",
	}
}

// NewColorWithAlpha creates a Color with the specified alpha in the sRGB color space.
func NewColorWithAlpha(r, g, b, a float64) *Color {
	return &Color{
		Red:        r,
		Green:      g,
		Blue:       b,
		Alpha:      a,
		ColorSpace: "sRGB",
	}
}

// NewColorWithColorSpace creates a Color with the specified color space and full opacity.
func NewColorWithColorSpace(r, g, b float64, cs string) *Color {
	return &Color{
		Red:        r,
		Green:      g,
		Blue:       b,
		Alpha:      255,
		ColorSpace: cs,
	}
}

// ColorPreset is a named collection of colors for terminal attributes.
type ColorPreset struct {
	Name   string
	Colors map[string]*Color
}

// ListColorPresets returns the names of all available color presets.
func ListColorPresets(ctx context.Context, caller Caller) ([]string, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ColorPresetRequest{
		ColorPresetRequest: &iterm2.ColorPresetRequest{
			Request: &iterm2.ColorPresetRequest_ListPresets_{
				ListPresets: &iterm2.ColorPresetRequest_ListPresets{},
			},
		},
	}

	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}

	cpResp := resp.GetColorPresetResponse()
	status := cpResp.GetStatus()
	if status != iterm2.ColorPresetResponse_OK {
		return nil, fmt.Errorf("list color presets: status %s", status.String())
	}

	return cpResp.GetListPresets().GetName(), nil
}

// GetColorPreset fetches a color preset by name.
func GetColorPreset(ctx context.Context, caller Caller, name string) (*ColorPreset, error) {
	req := newRequest()
	req.Submessage = &iterm2.ClientOriginatedMessage_ColorPresetRequest{
		ColorPresetRequest: &iterm2.ColorPresetRequest{
			Request: &iterm2.ColorPresetRequest_GetPreset_{
				GetPreset: &iterm2.ColorPresetRequest_GetPreset{
					Name: proto.String(name),
				},
			},
		},
	}

	resp, err := caller.Call(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}

	cpResp := resp.GetColorPresetResponse()
	status := cpResp.GetStatus()
	if status != iterm2.ColorPresetResponse_OK {
		return nil, fmt.Errorf("get color preset %q: status %s", name, status.String())
	}

	settings := cpResp.GetGetPreset().GetColorSettings()
	colors := make(map[string]*Color, len(settings))
	for _, s := range settings {
		colors[s.GetKey()] = &Color{
			Red:        float64(s.GetRed()) * 255,
			Green:      float64(s.GetGreen()) * 255,
			Blue:       float64(s.GetBlue()) * 255,
			Alpha:      float64(s.GetAlpha()) * 255,
			ColorSpace: s.GetColorSpace(),
		}
	}

	return &ColorPreset{Name: name, Colors: colors}, nil
}
