package term2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckboxKnob tests CheckboxKnob function for various cases.
func TestCheckboxKnob(t *testing.T) {
	t.Run("default false", func(t *testing.T) {
		key, value := CheckboxKnob("test-key", false)
		assert.Equal(t, "test-key", key)
		assert.Equal(t, "false", value)
	})
	t.Run("default true", func(t *testing.T) {
		key, value := CheckboxKnob("enabled", true)
		assert.Equal(t, "enabled", key)
		assert.Equal(t, "true", value)
	})
}

// TestStringKnob tests StringKnob function for various cases.
func TestStringKnob(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		key, value := StringKnob("name", "")
		assert.Equal(t, "name", key)
		assert.Equal(t, `""`, value)
	})
	t.Run("simple string", func(t *testing.T) {
		key, value := StringKnob("prefix", "hello")
		assert.Equal(t, "prefix", key)
		assert.Equal(t, `"hello"`, value)
	})
	t.Run("string with quotes", func(t *testing.T) {
		key, value := StringKnob("msg", `say "hello"`)
		assert.Equal(t, "msg", key)
		assert.Equal(t, `"say \"hello\""`, value)
	})
}

// TestFloatKnob tests FloatKnob function for various cases.
func TestFloatKnob(t *testing.T) {
	t.Run("integer float", func(t *testing.T) {
		key, value := FloatKnob("interval", 5.0)
		assert.Equal(t, "interval", key)
		assert.Equal(t, "5", value)
	})
	t.Run("decimal float", func(t *testing.T) {
		key, value := FloatKnob("ratio", 0.5)
		assert.Equal(t, "ratio", key)
		assert.Equal(t, "0.5", value)
	})
	t.Run("negative float", func(t *testing.T) {
		key, value := FloatKnob("offset", -10.5)
		assert.Equal(t, "offset", key)
		assert.Equal(t, "-10.5", value)
	})
}

// TestColorKnob tests ColorKnob function for various cases.
func TestColorKnob(t *testing.T) {
	t.Run("simple hex color", func(t *testing.T) {
		key, value := ColorKnob("bg", `#FF0000`)
		assert.Equal(t, "bg", key)
		assert.Equal(t, `#FF0000`, value)
	})
	t.Run("empty string", func(t *testing.T) {
		key, value := ColorKnob("color", "")
		assert.Equal(t, "color", key)
		assert.Equal(t, "", value)
	})
}

// TestStatusBarComponent_Fields tests StatusBarComponent field settings.
func TestStatusBarComponent_Fields(t *testing.T) {
	t.Run("basic fields", func(t *testing.T) {
		component := StatusBarComponent{
			Identifier:          "my-status-bar",
			ShortDescription:    "My Status Bar",
			DetailedDescription: "A custom status bar component",
			Knobs:               map[string]string{"enabled": "true"},
			Exemplar:            "● Online",
			UpdateCadence:       5.0,
			Format:              StatusBarFormatHTML,
		}
		assert.Equal(t, "my-status-bar", component.Identifier)
		assert.Equal(t, "My Status Bar", component.ShortDescription)
		assert.Equal(t, "A custom status bar component", component.DetailedDescription)
		assert.Equal(t, "true", component.Knobs["enabled"])
		assert.Equal(t, "● Online", component.Exemplar)
		assert.Equal(t, 5.0, component.UpdateCadence)
		assert.Equal(t, StatusBarFormatHTML, component.Format)
	})
	t.Run("with icons", func(t *testing.T) {
		icons := []StatusBarIcon{
			{Scale: 1, Data: []byte("icon1-png")},
			{Scale: 2, Data: []byte("icon2-retina")},
		}
		component := StatusBarComponent{
			Identifier: "with-icons",
			Icons:      icons,
		}
		assert.Len(t, component.Icons, 2)
		assert.Equal(t, 1.0, component.Icons[0].Scale)
		assert.Equal(t, 2.0, component.Icons[1].Scale)
	})
	t.Run("with empty knobs", func(t *testing.T) {
		component := StatusBarComponent{
			Identifier: "no-knobs",
			Knobs:      nil,
		}
		assert.Nil(t, component.Knobs)
	})
	t.Run("with zero update cadence", func(t *testing.T) {
		component := StatusBarComponent{
			Identifier:    "no-timer",
			UpdateCadence: 0,
		}
		assert.Equal(t, 0.0, component.UpdateCadence)
	})
	t.Run("plain text format", func(t *testing.T) {
		component := StatusBarComponent{
			Identifier: "plain",
			Format:     StatusBarFormatPlainText,
		}
		assert.Equal(t, StatusBarFormatPlainText, component.Format)
	})
}

// TestStatusBarFormat_Constants tests StatusBarFormat constant values.
func TestStatusBarFormat_Constants(t *testing.T) {
	assert.Equal(t, StatusBarFormat(0), StatusBarFormatPlainText)
	assert.Equal(t, StatusBarFormat(1), StatusBarFormatHTML)
}

// TestStatusBarIcon_Fields tests StatusBarIcon field settings.
func TestStatusBarIcon_Fields(t *testing.T) {
	icon := StatusBarIcon{
		Scale: 2.0,
		Data:  []byte("PNG data here"),
	}
	assert.Equal(t, 2.0, icon.Scale)
	assert.Equal(t, "PNG data here", string(icon.Data))
}
