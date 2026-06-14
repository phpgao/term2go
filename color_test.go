package term2go

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestNewColor tests NewColor function for creating a color with RGB values.
func TestNewColor(t *testing.T) {
	c := NewColor(100, 150, 200)
	assert.Equal(t, float64(100), c.Red)
	assert.Equal(t, float64(150), c.Green)
	assert.Equal(t, float64(200), c.Blue)
	assert.Equal(t, float64(255), c.Alpha)
	assert.Equal(t, "sRGB", c.ColorSpace)
}

// TestNewColorWithAlpha tests NewColorWithAlpha function for creating a color with alpha.
func TestNewColorWithAlpha(t *testing.T) {
	c := NewColorWithAlpha(10, 20, 30, 128)
	assert.Equal(t, float64(10), c.Red)
	assert.Equal(t, float64(20), c.Green)
	assert.Equal(t, float64(30), c.Blue)
	assert.Equal(t, float64(128), c.Alpha)
	assert.Equal(t, "sRGB", c.ColorSpace)
}

// TestNewColorWithColorSpace tests NewColorWithColorSpace function for creating a color with custom color space.
func TestNewColorWithColorSpace(t *testing.T) {
	c := NewColorWithColorSpace(50, 60, 70, "Display P3")
	assert.Equal(t, float64(50), c.Red)
	assert.Equal(t, float64(60), c.Green)
	assert.Equal(t, float64(70), c.Blue)
	assert.Equal(t, float64(255), c.Alpha)
	assert.Equal(t, "Display P3", c.ColorSpace)
}

// TestListColorPresets_Success tests ListColorPresets function success return.
func TestListColorPresets_Success(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := &iterm2.ColorPresetResponse{
				Status: iterm2.ColorPresetResponse_OK.Enum(),
				Response: &iterm2.ColorPresetResponse_ListPresets_{
					ListPresets: &iterm2.ColorPresetResponse_ListPresets{
						Name: []string{"Default Dark", "Espresso", "Catppuccin Mocha"},
					},
				},
			}
			return colorSuccessResp(resp), nil
		},
	}

	ctx := context.Background()
	names, err := ListColorPresets(ctx, caller)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Default Dark", "Espresso", "Catppuccin Mocha"}, names)
}

// TestListColorPresets_CallerError tests ListColorPresets function when caller returns an error.
func TestListColorPresets_CallerError(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ListColorPresets(ctx, caller)

	assert.Error(t, err)
}

// TestListColorPresets_NonOKStatus tests ListColorPresets function when iTerm2 returns a non-OK status.
func TestListColorPresets_NonOKStatus(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := &iterm2.ColorPresetResponse{
				Status: iterm2.ColorPresetResponse_PRESET_NOT_FOUND.Enum(),
			}
			return colorSuccessResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ListColorPresets(ctx, caller)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list color presets")
}

// TestGetColorPreset_Success tests GetColorPreset function success return.
func TestGetColorPreset_Success(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := &iterm2.ColorPresetResponse{
				Status: iterm2.ColorPresetResponse_OK.Enum(),
				Response: &iterm2.ColorPresetResponse_GetPreset_{
					GetPreset: &iterm2.ColorPresetResponse_GetPreset{
						ColorSettings: []*iterm2.ColorPresetResponse_GetPreset_ColorSetting{
							{Key: strPtr("Background Color"), Red: float32Ptr(0), Green: float32Ptr(0), Blue: float32Ptr(0), Alpha: float32Ptr(1.0), ColorSpace: strPtr("sRGB")},
							{Key: strPtr("Foreground Color"), Red: float32Ptr(1.0), Green: float32Ptr(1.0), Blue: float32Ptr(1.0), Alpha: float32Ptr(1.0), ColorSpace: strPtr("sRGB")},
						},
					},
				},
			}
			return colorSuccessResp(resp), nil
		},
	}

	ctx := context.Background()
	preset, err := GetColorPreset(ctx, caller, "Default Dark")

	assert.NoError(t, err)
	assert.Equal(t, "Default Dark", preset.Name)
	assert.Equal(t, 2, len(preset.Colors))

	bgColor := preset.Colors["Background Color"]
	assert.NotNil(t, bgColor)
	assert.Equal(t, float64(0), bgColor.Red)
	assert.Equal(t, float64(0), bgColor.Green)
	assert.Equal(t, float64(0), bgColor.Blue)
	assert.Equal(t, float64(255), bgColor.Alpha)

	fgColor := preset.Colors["Foreground Color"]
	assert.NotNil(t, fgColor)
	assert.Equal(t, float64(255), fgColor.Red)
	assert.Equal(t, float64(255), fgColor.Green)
	assert.Equal(t, float64(255), fgColor.Blue)
}

// TestGetColorPreset_CallerError tests GetColorPreset function when caller returns an error.
func TestGetColorPreset_CallerError(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := GetColorPreset(ctx, caller, "Default Dark")

	assert.Error(t, err)
}

// TestGetColorPreset_NonOKStatus tests GetColorPreset function when iTerm2 returns a non-OK status.
func TestGetColorPreset_NonOKStatus(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := &iterm2.ColorPresetResponse{
				Status: iterm2.ColorPresetResponse_PRESET_NOT_FOUND.Enum(),
			}
			return colorSuccessResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := GetColorPreset(ctx, caller, "NonExistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get color preset")
}

// TestGetColorPreset_WithDisplayP3 tests GetColorPreset function with Display P3 color space.
func TestGetColorPreset_WithDisplayP3(t *testing.T) {
	caller := &colorMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := &iterm2.ColorPresetResponse{
				Status: iterm2.ColorPresetResponse_OK.Enum(),
				Response: &iterm2.ColorPresetResponse_GetPreset_{
					GetPreset: &iterm2.ColorPresetResponse_GetPreset{
						ColorSettings: []*iterm2.ColorPresetResponse_GetPreset_ColorSetting{
							{Key: strPtr("Cursor Color"), Red: float32Ptr(0.5), Green: float32Ptr(0.5), Blue: float32Ptr(0.5), Alpha: float32Ptr(0.5), ColorSpace: strPtr("Display P3")},
						},
					},
				},
			}
			return colorSuccessResp(resp), nil
		},
	}

	ctx := context.Background()
	preset, err := GetColorPreset(ctx, caller, "Custom Preset")

	assert.NoError(t, err)
	assert.Equal(t, "Custom Preset", preset.Name)

	cursorColor := preset.Colors["Cursor Color"]
	assert.NotNil(t, cursorColor)
	assert.Equal(t, "Display P3", cursorColor.ColorSpace)
	assert.Equal(t, float64(127.5), cursorColor.Alpha) // 0.5 * 255 = 127.5
}

// colorSuccessResp wraps a ColorPresetResponse in a ServerOriginatedMessage
func colorSuccessResp(sub proto.Message) *iterm2.ServerOriginatedMessage {
	var s iterm2.ServerOriginatedMessage
	switch v := sub.(type) {
	case *iterm2.ColorPresetResponse:
		s.Submessage = &iterm2.ServerOriginatedMessage_ColorPresetResponse{ColorPresetResponse: v}
	}
	return &s
}

// colorMockCaller is a test helper that implements Caller interface
type colorMockCaller struct {
	callFunc func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error)
}

func (m *colorMockCaller) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
	if m.callFunc != nil {
		return m.callFunc(ctx, req)
	}
	return nil, nil
}

func (m *colorMockCaller) Send(req *iterm2.ClientOriginatedMessage) error {
	return nil
}

func float32Ptr(f float32) *float32 { return &f }
