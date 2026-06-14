package term2go

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// ============================================================================
// Alert
// ============================================================================

// ShowAlert displays a modal alert with buttons. Returns the button index (0-based).
func ShowAlert(ctx context.Context, caller Caller, title, message string, buttons []string) (int, error) {
	bs, _ := json.Marshal(buttons)
	invocation := fmt.Sprintf("iterm2.alert(title: %s, subtitle: %s, buttons: %s, window_id: null)",
		jsonEncode(title), jsonEncode(message), string(bs))
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return 0, err
	}
	return parseAlertResult(resp)
}

// ShowTextInputAlert displays a modal alert with a text field. Returns the entered text.
func ShowTextInputAlert(ctx context.Context, caller Caller, title, message, defaultValue string) (string, error) {
	invocation := fmt.Sprintf("iterm2.get_string(title: %s, subtitle: %s, placeholder: %s, defaultValue: %s, window_id: null)",
		jsonEncode(title), jsonEncode(message), jsonEncode(""), jsonEncode(defaultValue))
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return "", err
	}
	return parseTextInputResult(resp)
}

// ============================================================================
// PolyModalAlert
// ============================================================================

// PolyModalResult holds the returned values of a PolyModalAlert.
type PolyModalResult struct {
	Button     string   // label of the clicked button
	TextField  string   // text entered into the field
	ComboBox   string   // selected combobox item
	Checkboxes []string // checked checkbox labels
}

// PolyModalAlert is a modal alert with checkboxes, combobox, and text field.
type PolyModalAlert struct {
	Title            string
	Subtitle         string
	WindowID         string
	Width            int
	Buttons          []string
	CheckboxItems    []string
	CheckboxDefaults []int
	ComboBoxItems    []string
	ComboBoxDefault  string
	TextFieldDefault string
	TextFieldLabel   string
}

// NewPolyModalAlert creates a new PolyModalAlert.
func NewPolyModalAlert(title, subtitle string) *PolyModalAlert {
	return &PolyModalAlert{
		Title:    title,
		Subtitle: subtitle,
		Width:    300,
	}
}

// AddButton adds a button.
func (a *PolyModalAlert) AddButton(label string) {
	a.Buttons = append(a.Buttons, label)
}

// AddCheckbox adds a checkbox with default state (1=checked, 0=unchecked).
func (a *PolyModalAlert) AddCheckbox(label string, checked bool) {
	a.CheckboxItems = append(a.CheckboxItems, label)
	if checked {
		a.CheckboxDefaults = append(a.CheckboxDefaults, 1)
	} else {
		a.CheckboxDefaults = append(a.CheckboxDefaults, 0)
	}
}

// AddComboBox replaces cometable items and sets the default selection.
func (a *PolyModalAlert) AddComboBox(items []string, defaultItem string) {
	a.ComboBoxItems = append(a.ComboBoxItems, items...)
	a.ComboBoxDefault = defaultItem
}

// AddTextField adds a text field with placeholder and default value.
func (a *PolyModalAlert) AddTextField(placeholder, defaultValue string) {
	a.TextFieldLabel = placeholder
	a.TextFieldDefault = defaultValue
}

// Run displays the poly modal alert and returns the result.
func (a *PolyModalAlert) Run(ctx context.Context, caller Caller) (*PolyModalResult, error) {
	invocation := fmt.Sprintf(
		"iterm2.get_poly_modal_alert(title: %s, subtitle: %s, buttons: %s, checkboxes: %s, checkboxDefaults: %s, comboboxItems: %s, comboboxDefault: %s, textFieldParams: %s, width: %s, window_id: %s)",
		jsonEncode(a.Title),
		jsonEncode(a.Subtitle),
		jsonMarshal(a.Buttons),
		jsonMarshal(a.CheckboxItems),
		jsonMarshal(a.CheckboxDefaults),
		jsonMarshal(a.ComboBoxItems),
		jsonEncode(a.ComboBoxDefault),
		jsonMarshal([]string{a.TextFieldLabel, a.TextFieldDefault}),
		jsonMarshal(a.Width),
		jsonMarshal(a.WindowID),
	)
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return nil, err
	}
	var raw string
	raw, err = parseInvokeResult(resp)
	if err != nil {
		return nil, err
	}
	var result struct {
		Button     string   `json:"button"`
		TextField  string   `json:"tf_text"`
		ComboBox   string   `json:"combo"`
		Checkboxes []string `json:"checks"`
	}
	if err = json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse poly modal result: %w", err)
	}
	return &PolyModalResult{
		Button:     result.Button,
		TextField:  result.TextField,
		ComboBox:   result.ComboBox,
		Checkboxes: result.Checkboxes,
	}, nil
}

// ============================================================================
// File Panel Options
// ============================================================================

// OpenPanelOptions are flags for ShowOpenPanel.
type OpenPanelOptions int

const (
	OpenPanelCanCreateDirectories            OpenPanelOptions = 1 << 0
	OpenPanelTreatsFilePackagesAsDirectories OpenPanelOptions = 1 << 1
	OpenPanelShowsHiddenFiles                OpenPanelOptions = 1 << 2
	OpenPanelResolvesAliases                 OpenPanelOptions = 1 << 32
	OpenPanelCanChooseDirectories            OpenPanelOptions = 1 << 33
	OpenPanelAllowsMultipleSelection         OpenPanelOptions = 1 << 34
	OpenPanelCanChooseFiles                  OpenPanelOptions = 1 << 35
)

// OpenPanelResult holds the files selected in the open panel.
type OpenPanelResult struct {
	Files []string
}

// SavePanelOptions are flags for ShowSavePanel.
type SavePanelOptions int

const (
	SavePanelCanCreateDirectories            SavePanelOptions = 1 << 0
	SavePanelTreatsFilePackagesAsDirectories SavePanelOptions = 1 << 1
	SavePanelShowsHiddenFiles                SavePanelOptions = 1 << 2
	SavePanelAllowsOtherFileTypes            SavePanelOptions = 1 << 3
	SavePanelCanSelectHiddenExtension        SavePanelOptions = 1 << 4
	SavePanelExtensionHidden                 SavePanelOptions = 1 << 5
)

// SavePanelResult holds the file path selected in the save panel.
type SavePanelResult struct {
	File string
}

// ShowOpenPanel displays an open file panel and returns selected files.
func ShowOpenPanel(ctx context.Context, caller Caller, title, initialPath string) (string, error) {
	invocation := fmt.Sprintf("iterm2.open_panel(path: %s, options: -1, extensions: null, prompt: %s, message: %s)",
		jsonEncode(initialPath), jsonEncode("Open"), jsonEncode(title))
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return "", err
	}
	return parseOpenPanelResult(resp)
}

// ShowOpenPanelWithOptions displays an open file panel with full options.
func ShowOpenPanelWithOptions(ctx context.Context, caller Caller, title, message, initialPath, prompt string, options OpenPanelOptions, extensions []string) (*OpenPanelResult, error) {
	path := jsonEncode(initialPath)
	if initialPath == "" {
		path = "null"
	}
	extJSON := "null"
	if extensions != nil {
		extJSON = string(jsonMarshal(extensions))
	}
	invocation := fmt.Sprintf(
		"iterm2.open_panel(path: %s, options: %d, extensions: %s, prompt: %s, message: %s)",
		path, int(options), extJSON, jsonEncode(prompt), jsonEncode(message),
	)
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return nil, err
	}
	var raw string
	raw, err = parseInvokeResult(resp)
	if err != nil {
		return nil, err
	}
	var files []string
	if err = json.Unmarshal([]byte(raw), &files); err != nil {
		return nil, fmt.Errorf("parse open panel result: %w", err)
	}
	return &OpenPanelResult{Files: files}, nil
}

// ShowSavePanel displays a save file panel and returns the selected path.
func ShowSavePanel(ctx context.Context, caller Caller, title, initialPath string) (string, error) {
	invocation := fmt.Sprintf("iterm2.save_panel(path: %s, options: 0, extensions: null, prompt: %s, title: %s, message: %s, name_field_label: null, default_filename: null)",
		jsonEncode(initialPath), jsonEncode("Save"), jsonEncode(title), jsonEncode(""))
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return "", err
	}
	return parseSavePanelResult(resp)
}

// ShowSavePanelWithOptions displays a save file panel with full options.
func ShowSavePanelWithOptions(ctx context.Context, caller Caller, title, message, initialPath, prompt, defaultFilename, nameFieldLabel string, options SavePanelOptions, extensions []string) (*SavePanelResult, error) {
	path := jsonEncode(initialPath)
	if initialPath == "" {
		path = "null"
	}
	extJSON := "null"
	if extensions != nil {
		extJSON = string(jsonMarshal(extensions))
	}
	df := jsonEncode(defaultFilename)
	if defaultFilename == "" {
		df = "null"
	}
	nfl := jsonEncode(nameFieldLabel)
	if nameFieldLabel == "" {
		nfl = "null"
	}
	invocation := fmt.Sprintf(
		"iterm2.save_panel(path: %s, options: %d, extensions: %s, prompt: %s, title: %s, message: %s, name_field_label: %s, default_filename: %s)",
		path, int(options), extJSON, jsonEncode(prompt), jsonEncode(title), jsonEncode(message), nfl, df,
	)
	resp, err := InvokeFunction(ctx, caller, &iterm2.InvokeFunctionRequest{Invocation: proto.String(invocation)})
	if err != nil {
		return nil, err
	}
	var raw string
	raw, err = parseInvokeResult(resp)
	if err != nil {
		return nil, err
	}
	if raw == "null" {
		return nil, nil
	}
	var file string
	if err = json.Unmarshal([]byte(raw), &file); err != nil {
		return nil, fmt.Errorf("parse save panel result: %w", err)
	}
	return &SavePanelResult{File: file}, nil
}

// ============================================================================
// Helpers
// ============================================================================

func jsonEncode(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		log.Printf("term2go: jsonEncode error: %v", err)
		return `""`
	}
	return string(b)
}

func jsonMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("term2go: jsonMarshal error: %v", err)
		return "[]"
	}
	return string(b)
}

func parseInvokeResult(resp *iterm2.InvokeFunctionResponse) (string, error) {
	if err := resp.GetError(); err != nil {
		return "", fmt.Errorf("invoke function failed (%v): %s", err.GetStatus(), err.GetErrorReason())
	}
	s := resp.GetSuccess()
	if s == nil {
		return "", fmt.Errorf("invoke function returned no result")
	}
	return s.GetJsonResult(), nil
}

func parseAlertResult(resp *iterm2.InvokeFunctionResponse) (int, error) {
	raw, err := parseInvokeResult(resp)
	if err != nil {
		return 0, err
	}
	var n int
	if err = json.Unmarshal([]byte(raw), &n); err != nil {
		return 0, fmt.Errorf("parse alert result: %w", err)
	}
	// iTerm2 returns button index + 1000 (1000 is an internal offset used by iTerm2)
	return n - 1000, nil
}

func parseTextInputResult(resp *iterm2.InvokeFunctionResponse) (string, error) {
	raw, err := parseInvokeResult(resp)
	if err != nil {
		return "", err
	}
	if raw == "null" {
		return "", nil
	}
	var s string
	if err = json.Unmarshal([]byte(raw), &s); err != nil {
		return "", fmt.Errorf("parse text input result: %w", err)
	}
	return s, nil
}

func parseOpenPanelResult(resp *iterm2.InvokeFunctionResponse) (string, error) {
	raw, err := parseInvokeResult(resp)
	if err != nil {
		return "", err
	}
	var files []string
	if err = json.Unmarshal([]byte(raw), &files); err != nil {
		return "", fmt.Errorf("parse open panel result: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no file selected")
	}
	return files[0], nil
}

func parseSavePanelResult(resp *iterm2.InvokeFunctionResponse) (string, error) {
	raw, err := parseInvokeResult(resp)
	if err != nil {
		return "", err
	}
	if raw == "null" {
		return "", nil
	}
	var s string
	if err = json.Unmarshal([]byte(raw), &s); err != nil {
		return "", fmt.Errorf("parse save panel result: %w", err)
	}
	return s, nil
}
