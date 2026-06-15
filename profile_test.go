package term2go

import (
	"encoding/json"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iterm2 "github.com/phpgao/term2go/proto"
)

// NewLocalWriteOnlyProfile

func TestNewLocalWriteOnlyProfile(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	assert.NotNil(t, p)
	assert.Empty(t, p.Values())
}

// Assignments

func TestLocalWriteOnlyProfile_Assignments(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetFont("Menlo").SetUseBrightBold(true)
	as := p.Assignments()
	require.Len(t, as, 2)

	keys := map[string]string{}
	for _, a := range as {
		keys[a.GetKey()] = a.GetJsonValue()
	}
	var fontName string
	json.Unmarshal([]byte(keys[ProfileKeyFont]), &fontName)
	assert.Equal(t, "Menlo", fontName)
	assert.Equal(t, "true", keys[ProfileKeyUseBrightBold])
}

func TestLocalWriteOnlyProfile_Assignments_Empty(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	assert.Empty(t, p.Assignments())
}

// Color Setters

func TestLocalWriteOnlyProfile_SetForegroundColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	c := NewColor(128, 64, 32)
	p.SetForegroundColor(c)
	v := p.Values()[ProfileKeyForegroundColor]
	assert.Contains(t, v, "Red Component")
	assert.Contains(t, v, "0.501")
}

func TestLocalWriteOnlyProfile_SetBackgroundColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	c := NewColor(0, 0, 0)
	p.SetBackgroundColor(c)
	assert.Contains(t, p.Values()[ProfileKeyBackgroundColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetCursorColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	c := NewColor(255, 255, 255)
	p.SetCursorColor(c)
	assert.Contains(t, p.Values()[ProfileKeyCursorColor], `"Red Component":1`)
}

func TestLocalWriteOnlyProfile_SetBoldColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBoldColor(NewColor(200, 100, 50))
	assert.Contains(t, p.Values()[ProfileKeyBoldColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetSelectionColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetSelectionColor(NewColor(50, 100, 200))
	assert.Contains(t, p.Values()[ProfileKeySelectionColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetSelectedTextColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetSelectedTextColor(NewColor(255, 255, 0))
	assert.Contains(t, p.Values()[ProfileKeySelectedTextColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetLinkColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetLinkColor(NewColor(0, 0, 255))
	assert.Contains(t, p.Values()[ProfileKeyLinkColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetCursorTextColor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetCursorTextColor(NewColor(0, 0, 0))
	assert.Contains(t, p.Values()[ProfileKeyCursorTextColor], "Red Component")
}

func TestLocalWriteOnlyProfile_SetColor_Nil(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetForegroundColor(nil)
	assert.Empty(t, p.Values())
}

// Boolean Setters

func TestLocalWriteOnlyProfile_SetUseBrightBold(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetUseBrightBold(true)
	assert.Equal(t, "true", p.Values()[ProfileKeyUseBrightBold])
	p.SetUseBrightBold(false)
	assert.Equal(t, "false", p.Values()[ProfileKeyUseBrightBold])
}

func TestLocalWriteOnlyProfile_SetBoldIsBright(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBoldIsBright(true)
	assert.Equal(t, "true", p.Values()[ProfileKeyBoldIsBright])
}

func TestLocalWriteOnlyProfile_SetBlinkingCursor(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBlinkingCursor(true)
	assert.Equal(t, "true", p.Values()[ProfileKeyBlinkingCursor])
}

func TestLocalWriteOnlyProfile_SetUseLigatures(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetUseLigatures(false)
	assert.Equal(t, "false", p.Values()[ProfileKeyUseLigatures])
}

func TestLocalWriteOnlyProfile_SetSilenceBell(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetSilenceBell(true)
	assert.Equal(t, "true", p.Values()[ProfileKeySilenceBell])
}

func TestLocalWriteOnlyProfile_SetUnlimitedScrollback(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetUnlimitedScrollback(false)
	assert.Equal(t, "false", p.Values()[ProfileKeyUnlimitedScrollback])
}

// Number Setters

func TestLocalWriteOnlyProfile_SetScrollbackLines(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetScrollbackLines(5000)
	var v int
	json.Unmarshal([]byte(p.Values()[ProfileKeyScrollbackLines]), &v)
	assert.Equal(t, 5000, v)
}

func TestLocalWriteOnlyProfile_SetTransparency(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetTransparency(0.3)
	var v float64
	json.Unmarshal([]byte(p.Values()[ProfileKeyTransparency]), &v)
	assert.InDelta(t, 0.3, v, 0.001)
}

func TestLocalWriteOnlyProfile_SetBlend(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBlend(0.5)
	var v float64
	json.Unmarshal([]byte(p.Values()[ProfileKeyBlend]), &v)
	assert.InDelta(t, 0.5, v, 0.001)
}

func TestLocalWriteOnlyProfile_SetBlur(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBlur(0.8)
	var v float64
	json.Unmarshal([]byte(p.Values()[ProfileKeyBlur]), &v)
	assert.InDelta(t, 0.8, v, 0.001)
}

// String Setters

func TestLocalWriteOnlyProfile_SetFont(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetFont("Fira Code")
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyFont]), &v)
	assert.Equal(t, "Fira Code", v)
}

func TestLocalWriteOnlyProfile_SetBadgeText(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBadgeText("PROD")
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyBadgeText]), &v)
	assert.Equal(t, "PROD", v)
}

func TestLocalWriteOnlyProfile_SetCustomCommand(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetCustomCommand("ssh prod")
	assert.Equal(t, "true", p.Values()[ProfileKeyCustomCommand])
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyCommand]), &v)
	assert.Equal(t, "ssh prod", v)
}

func TestLocalWriteOnlyProfile_SetCommand(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetCommand("htop")
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyCommand]), &v)
	assert.Equal(t, "htop", v)
}

func TestLocalWriteOnlyProfile_SetCursorType(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetCursorType("underline")
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyCursorType]), &v)
	assert.Equal(t, "underline", v)
}

func TestLocalWriteOnlyProfile_SetBackgroundImageLocation(t *testing.T) {
	p := NewLocalWriteOnlyProfile()
	p.SetBackgroundImageLocation("/path/to/img.png")
	var v string
	json.Unmarshal([]byte(p.Values()[ProfileKeyBackgroundImageLocation]), &v)
	assert.Equal(t, "/path/to/img.png", v)
}

// Chaining

func TestLocalWriteOnlyProfile_Chaining(t *testing.T) {
	p := NewLocalWriteOnlyProfile().
		SetForegroundColor(NewColor(255, 0, 0)).
		SetBackgroundColor(NewColor(0, 0, 0)).
		SetFont("Monaco").
		SetUseBrightBold(true).
		SetTransparency(0.3).
		SetScrollbackLines(10000)
	assert.Len(t, p.Values(), 6)
}

// Session.SetProfileProperties

func TestSession_SetProfileProperties(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	s := &Session{caller: mc, ID: "s1"}

	p := NewLocalWriteOnlyProfile().SetFont("Menlo").SetUseBrightBold(false)
	err := s.SetProfileProperties(ctx, p)
	require.NoError(t, err)

	req := mc.req.GetSetProfilePropertyRequest()
	require.NotNil(t, req)
	assert.Equal(t, "s1", req.GetSession())
	assert.Len(t, req.GetAssignments(), 2)
}

func TestSession_SetProfileProperties_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("rpc fail")}
	s := &Session{caller: mc, ID: "s1"}
	p := NewLocalWriteOnlyProfile().SetFont("Menlo")
	err := s.SetProfileProperties(ctx, p)
	require.Error(t, err)
}

// Session.GetProfile

func TestSession_GetProfile(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetProfilePropertyResponse{
			Properties: []*iterm2.ProfileProperty{
				{Key: proto.String(ProfileKeyFont), JsonValue: proto.String(`"Menlo"`)},
				{Key: proto.String(ProfileKeyUseBrightBold), JsonValue: proto.String("true")},
			},
		}),
	}
	s := &Session{caller: mc, ID: "s1"}

	props, err := s.GetProfile(ctx)
	require.NoError(t, err)
	assert.Equal(t, `"Menlo"`, props[ProfileKeyFont])
	assert.Equal(t, "true", props[ProfileKeyUseBrightBold])
}

func TestSession_GetProfile_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.GetProfile(ctx)
	require.Error(t, err)
}

// Session.GetProfileProperty

func TestSession_GetProfileProperty(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetProfilePropertyResponse{
			Properties: []*iterm2.ProfileProperty{
				{Key: proto.String(ProfileKeyFont), JsonValue: proto.String(`"Menlo"`)},
			},
		}),
	}
	s := &Session{caller: mc, ID: "s1"}

	v, err := s.GetProfileProperty(ctx, ProfileKeyFont)
	require.NoError(t, err)
	assert.Equal(t, `"Menlo"`, v)
}

func TestSession_GetProfileProperty_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.GetProfileProperty(ctx, ProfileKeyFont)
	require.Error(t, err)
}

// Session.GetCoprocess

func TestSession_GetCoprocess(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.InvokeFunctionResponse{
			Disposition: &iterm2.InvokeFunctionResponse_Success_{
				Success: &iterm2.InvokeFunctionResponse_Success{
					JsonResult: proto.String(`"/usr/bin/copro"`),
				},
			},
		}),
	}
	s := &Session{caller: mc, ID: "s1"}
	result, err := s.GetCoprocess(ctx)
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/copro", result)
}

func TestSession_GetCoprocess_Error(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{err: errors.New("fail")}
	s := &Session{caller: mc, ID: "s1"}
	_, err := s.GetCoprocess(ctx)
	require.Error(t, err)
}

// Color.ToProfileJSON

func TestColor_ToProfileJSON(t *testing.T) {
	c := NewColor(255, 128, 0)
	j := c.ToProfileJSON()
	var m map[string]interface{}
	json.Unmarshal([]byte(j), &m)
	assert.InDelta(t, 1.0, m["Red Component"], 0.001)
	assert.InDelta(t, 128.0/255, m["Green Component"], 0.001)
	assert.InDelta(t, 0.0, m["Blue Component"], 0.001)
	assert.Equal(t, "sRGB", m["Color Space"])
}

// Profile key constants exist

func TestProfileKeyConstants(t *testing.T) {
	assert.Equal(t, "Foreground Color", ProfileKeyForegroundColor)
	assert.Equal(t, "Background Color", ProfileKeyBackgroundColor)
	assert.Equal(t, "Font", ProfileKeyFont)
	assert.Equal(t, "Transparency", ProfileKeyTransparency)
	assert.Equal(t, "Custom Command", ProfileKeyCustomCommand)
}
