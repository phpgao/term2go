package term2go

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Trigger — unified struct for all iTerm2 trigger types

// Trigger provides a unified representation of all iTerm2 trigger types.
// Use New*Trigger factory functions to create specific types, and the
// generic DecodeTrigger function to parse JSON-encoded triggers.
type Trigger struct {
	// Common fields
	Type      TriggerType
	Regex     string
	Param     string // serialised parameter string (type-dependent format)
	Instant   bool   // fire immediately, don't wait for newline
	Enabled   bool
	MatchType TriggerMatchType

	// Event-trigger parameters (MatchType >= 100)
	EventParams map[string]interface{}

	// Pre-parsed parameters for convenience (set by factory/decode)
	parsedActions []string // for types that use `param` as multiple values
	ExitCode      string   // CommandFinishedEvent: "*", "0", "!0"
	Threshold     float64  // IdleEvent/ActivityAfterIdle/LongRunningCommand
	Timeout       float64  // IdleEvent/ActivityAfterIdle
	Sequence      string   // CustomEscapeSequenceEvent
	Progress      string   // ProgressBarChangedEvent: "*", "appeared", "disappeared"
}

// TriggerType identifies the kind of trigger.
type TriggerType string

const (
	TriggerAlert            TriggerType = "AlertTrigger"
	TriggerAnnotate         TriggerType = "AnnotateTrigger"
	TriggerBell             TriggerType = "BellTrigger"
	TriggerBounce           TriggerType = "BounceTrigger"
	TriggerBufferInput      TriggerType = "iTermBufferInputTrigger"
	TriggerRPC              TriggerType = "iTermRPCTrigger"
	TriggerCapture          TriggerType = "CaptureTrigger"
	TriggerSetNamedMark     TriggerType = "iTermSetNamedMarkTrigger"
	TriggerSGR              TriggerType = "iTermSGRTrigger"
	TriggerFold             TriggerType = "iTermFoldTrigger"
	TriggerInject           TriggerType = "iTermInjectTrigger"
	TriggerHighlightLine    TriggerType = "iTermHighlightLineTrigger"
	TriggerHighlight        TriggerType = "HighlightTrigger"
	TriggerUserNotification TriggerType = "iTermUserNotificationTrigger"
	TriggerSetUserVariable  TriggerType = "iTermSetUserVariableTrigger"
	TriggerShellPrompt      TriggerType = "iTermShellPromptTrigger"
	TriggerSetTitle         TriggerType = "iTermSetTitleTrigger"
	TriggerSendText         TriggerType = "SendTextTrigger"
	TriggerRunCommand       TriggerType = "ScriptTrigger"
	TriggerCoprocess        TriggerType = "CoprocessTrigger"
	TriggerMuteCoprocess    TriggerType = "MuteCoprocessTrigger"
	TriggerMark             TriggerType = "MarkTrigger"
	TriggerPassword         TriggerType = "PasswordTrigger"
	TriggerHyperlink        TriggerType = "iTermHyperlinkTrigger"
	TriggerSetDirectory     TriggerType = "SetDirectoryTrigger"
	TriggerSetHostname      TriggerType = "SetHostnameTrigger"
	TriggerStop             TriggerType = "StopTrigger"
	// Event triggers (MatchType >= 100)
	TriggerPromptDetectedEvent       TriggerType = "PromptDetectedEventTrigger"
	TriggerCommandFinishedEvent      TriggerType = "CommandFinishedEventTrigger"
	TriggerDirectoryChangedEvent     TriggerType = "DirectoryChangedEventTrigger"
	TriggerHostChangedEvent          TriggerType = "HostChangedEventTrigger"
	TriggerUserChangedEvent          TriggerType = "UserChangedEventTrigger"
	TriggerIdleEvent                 TriggerType = "IdleEventTrigger"
	TriggerActivityAfterIdleEvent    TriggerType = "ActivityAfterIdleEventTrigger"
	TriggerSessionEndedEvent         TriggerType = "SessionEndedEventTrigger"
	TriggerBellReceivedEvent         TriggerType = "BellReceivedEventTrigger"
	TriggerLongRunningCommandEvent   TriggerType = "LongRunningCommandEventTrigger"
	TriggerCustomEscapeSequenceEvent TriggerType = "CustomEscapeSequenceEventTrigger"
	TriggerNotificationPostedEvent   TriggerType = "NotificationPostedEventTrigger"
	TriggerProgressBarChangedEvent   TriggerType = "ProgressBarChangedEventTrigger"
)

// TriggerMatchType matches Python's MatchType enum.
type TriggerMatchType int

const (
	MatchTypeREGEX                     TriggerMatchType = 0
	MatchTypeURLRegex                  TriggerMatchType = 1
	MatchTypePageContentRegex          TriggerMatchType = 2
	MatchTypeEventPromptDetected       TriggerMatchType = 100
	MatchTypeEventCommandFinished      TriggerMatchType = 101
	MatchTypeEventDirectoryChanged     TriggerMatchType = 102
	MatchTypeEventHostChanged          TriggerMatchType = 103
	MatchTypeEventUserChanged          TriggerMatchType = 104
	MatchTypeEventIdle                 TriggerMatchType = 105
	MatchTypeEventActivityAfterIdle    TriggerMatchType = 106
	MatchTypeEventSessionEnded         TriggerMatchType = 107
	MatchTypeEventBellReceived         TriggerMatchType = 108
	MatchTypeEventLongRunningCommand   TriggerMatchType = 109
	MatchTypeEventCustomEscapeSequence TriggerMatchType = 110
	MatchTypeEventNotificationPosted   TriggerMatchType = 111
	MatchTypeEventProgressBarChanged   TriggerMatchType = 112
)

// Factory Functions — Regular Triggers (MatchType < 100)

func NewAlertTrigger(regex, message string) *Trigger {
	return &Trigger{Type: TriggerAlert, Regex: regex, Param: message, Enabled: true}
}

func NewAnnotateTrigger(regex, annotation string) *Trigger {
	return &Trigger{Type: TriggerAnnotate, Regex: regex, Param: annotation, Enabled: true}
}

func NewBellTrigger(regex string) *Trigger {
	return &Trigger{Type: TriggerBell, Regex: regex, Enabled: true}
}

func NewBounceTrigger(regex string, bounceOnce bool) *Trigger {
	p := "0"
	if bounceOnce {
		p = "1"
	}
	return &Trigger{Type: TriggerBounce, Regex: regex, Param: p, Enabled: true}
}

func NewBufferInputTrigger(regex string, start bool) *Trigger {
	p := "0"
	if start {
		p = "1"
	}
	return &Trigger{Type: TriggerBufferInput, Regex: regex, Param: p, Enabled: true}
}

func NewRPCTrigger(regex, invocation string) *Trigger {
	return &Trigger{Type: TriggerRPC, Regex: regex, Param: invocation, Enabled: true}
}

func NewCaptureTrigger(regex, command string) *Trigger {
	return &Trigger{Type: TriggerCapture, Regex: regex, Param: command, Enabled: true}
}

func NewSetNamedMarkTrigger(regex, markname string) *Trigger {
	return &Trigger{Type: TriggerSetNamedMark, Regex: regex, Param: markname, Enabled: true}
}

func NewSGRTrigger(regex, sgr string) *Trigger {
	return &Trigger{Type: TriggerSGR, Regex: regex, Param: sgr, Enabled: true}
}

func NewFoldTrigger(regex, markname string) *Trigger {
	return &Trigger{Type: TriggerFold, Regex: regex, Param: markname, Enabled: true}
}

func NewInjectTrigger(regex, injection string) *Trigger {
	return &Trigger{Type: TriggerInject, Regex: regex, Param: injection, Enabled: true}
}

func NewHighlightLineTrigger(regex, textColor, bgColor string) *Trigger {
	return &Trigger{Type: TriggerHighlightLine, Regex: regex, Param: fmt.Sprintf("{%s,%s}", textColor, bgColor), Enabled: true}
}

func NewHighlightTrigger(regex, textColor, bgColor string) *Trigger {
	return &Trigger{Type: TriggerHighlight, Regex: regex, Param: fmt.Sprintf("{%s,%s}", textColor, bgColor), Enabled: true}
}

func NewUserNotificationTrigger(regex, message string) *Trigger {
	return &Trigger{Type: TriggerUserNotification, Regex: regex, Param: message, Enabled: true}
}

func NewSetUserVariableTrigger(regex, name, jsonValue string) *Trigger {
	return &Trigger{Type: TriggerSetUserVariable, Regex: regex, Param: name + "\x01" + jsonValue, Enabled: true}
}

func NewShellPromptTrigger(regex string) *Trigger {
	return &Trigger{Type: TriggerShellPrompt, Regex: regex, Enabled: true}
}

func NewSetTitleTrigger(regex, title string) *Trigger {
	return &Trigger{Type: TriggerSetTitle, Regex: regex, Param: title, Enabled: true}
}

func NewSendTextTrigger(regex, text string) *Trigger {
	return &Trigger{Type: TriggerSendText, Regex: regex, Param: text, Enabled: true}
}

func NewRunCommandTrigger(regex, command string) *Trigger {
	return &Trigger{Type: TriggerRunCommand, Regex: regex, Param: command, Enabled: true}
}

func NewCoprocessTrigger(regex, command string) *Trigger {
	return &Trigger{Type: TriggerCoprocess, Regex: regex, Param: command, Enabled: true}
}

func NewMuteCoprocessTrigger(regex, command string) *Trigger {
	return &Trigger{Type: TriggerMuteCoprocess, Regex: regex, Param: command, Enabled: true}
}

func NewMarkTrigger(regex string, stopScrolling bool) *Trigger {
	p := "0"
	if stopScrolling {
		p = "1"
	}
	return &Trigger{Type: TriggerMark, Regex: regex, Param: p, Enabled: true}
}

func NewPasswordTrigger(regex, accountName, userName string) *Trigger {
	p := accountName
	if userName != "" {
		p += "\u2002\u2014\u2002" + userName
	}
	return &Trigger{Type: TriggerPassword, Regex: regex, Param: p, Enabled: true}
}

func NewHyperlinkTrigger(regex, url string) *Trigger {
	return &Trigger{Type: TriggerHyperlink, Regex: regex, Param: url, Enabled: true}
}

func NewSetDirectoryTrigger(regex, directory string) *Trigger {
	return &Trigger{Type: TriggerSetDirectory, Regex: regex, Param: directory, Enabled: true}
}

func NewSetHostnameTrigger(regex, hostname string) *Trigger {
	return &Trigger{Type: TriggerSetHostname, Regex: regex, Param: hostname, Enabled: true}
}

func NewStopTrigger(regex string) *Trigger {
	return &Trigger{Type: TriggerStop, Regex: regex, Enabled: true}
}

// Factory Functions — Event Triggers (MatchType >= 100)

func eventTrigger(t TriggerType, mt TriggerMatchType, eventParams map[string]interface{}) *Trigger {
	return &Trigger{Type: t, MatchType: mt, EventParams: eventParams, Enabled: true}
}

func NewPromptDetectedEventTrigger() *Trigger {
	return eventTrigger(TriggerPromptDetectedEvent, MatchTypeEventPromptDetected, nil)
}

func NewCommandFinishedEventTrigger(exitCodeFilter string) *Trigger {
	p := map[string]interface{}{"exitCodeFilter": exitCodeFilter}
	return eventTrigger(TriggerCommandFinishedEvent, MatchTypeEventCommandFinished, p)
}

func NewDirectoryChangedEventTrigger(dirRegex string) *Trigger {
	p := map[string]interface{}{}
	if dirRegex != "" {
		p["directoryRegex"] = dirRegex
	}
	return eventTrigger(TriggerDirectoryChangedEvent, MatchTypeEventDirectoryChanged, p)
}

func NewHostChangedEventTrigger(hostRegex string) *Trigger {
	p := map[string]interface{}{}
	if hostRegex != "" {
		p["hostRegex"] = hostRegex
	}
	return eventTrigger(TriggerHostChangedEvent, MatchTypeEventHostChanged, p)
}

func NewUserChangedEventTrigger(userRegex string) *Trigger {
	p := map[string]interface{}{}
	if userRegex != "" {
		p["userRegex"] = userRegex
	}
	return eventTrigger(TriggerUserChangedEvent, MatchTypeEventUserChanged, p)
}

func NewIdleEventTrigger(timeout float64) *Trigger {
	return eventTrigger(TriggerIdleEvent, MatchTypeEventIdle, map[string]interface{}{"timeout": timeout})
}

func NewActivityAfterIdleEventTrigger(timeout float64) *Trigger {
	return eventTrigger(TriggerActivityAfterIdleEvent, MatchTypeEventActivityAfterIdle, map[string]interface{}{"timeout": timeout})
}

func NewSessionEndedEventTrigger() *Trigger {
	return eventTrigger(TriggerSessionEndedEvent, MatchTypeEventSessionEnded, nil)
}

func NewBellReceivedEventTrigger() *Trigger {
	return eventTrigger(TriggerBellReceivedEvent, MatchTypeEventBellReceived, nil)
}

func NewLongRunningCommandEventTrigger(threshold float64, commandRegex string) *Trigger {
	p := map[string]interface{}{"threshold": threshold}
	if commandRegex != "" {
		p["commandRegex"] = commandRegex
	}
	return eventTrigger(TriggerLongRunningCommandEvent, MatchTypeEventLongRunningCommand, p)
}

func NewCustomEscapeSequenceEventTrigger(sequenceID string) *Trigger {
	p := map[string]interface{}{}
	if sequenceID != "" {
		p["sequenceId"] = sequenceID
	}
	return eventTrigger(TriggerCustomEscapeSequenceEvent, MatchTypeEventCustomEscapeSequence, p)
}

func NewNotificationPostedEventTrigger(messageRegex string) *Trigger {
	p := map[string]interface{}{}
	if messageRegex != "" {
		p["messageRegex"] = messageRegex
	}
	return eventTrigger(TriggerNotificationPostedEvent, MatchTypeEventNotificationPosted, p)
}

func NewProgressBarChangedEventTrigger(filter string) *Trigger {
	return eventTrigger(TriggerProgressBarChangedEvent, MatchTypeEventProgressBarChanged,
		map[string]interface{}{"progressBarFilter": filter})
}

// Serialisation / Deserialisation

func (t *Trigger) encode() map[string]interface{} {
	m := map[string]interface{}{
		"regex":    t.Regex,
		"action":   string(t.Type),
		"partial":  t.Instant,
		"disabled": !t.Enabled,
	}
	if t.MatchType >= 100 {
		m["matchType"] = int(t.MatchType)
		m["parameter"] = t.Param
		if len(t.EventParams) > 0 {
			m["eventParams"] = t.EventParams
		}
	} else {
		m["matchType"] = int(t.MatchType)
		m["parameter"] = t.Param
	}
	return m
}

// Encode serialises the trigger to a JSON-compatible map.
func (t *Trigger) Encode() map[string]interface{} { return t.encode() }

// eventTriggerMatchTypes maps MatchType values to TriggerType names.

// DecodeTrigger parses a JSON-encoded trigger dict from iTerm2.
func DecodeTrigger(encoded map[string]interface{}) (*Trigger, error) {
	action, _ := encoded["action"].(string)
	regex, _ := encoded["regex"].(string)
	param, _ := encoded["parameter"].(string)

	instant := false
	if v, ok := encoded["partial"].(bool); ok {
		instant = v
	}

	enabled := true
	if v, ok := encoded["disabled"].(bool); ok {
		enabled = !v
	}

	mt := TriggerMatchType(0)
	switch v := encoded["matchType"].(type) {
	case float64:
		mt = TriggerMatchType(int(v))
	case int:
		mt = TriggerMatchType(v)
	}

	t := &Trigger{
		Type:      TriggerType(action),
		Regex:     regex,
		Param:     param,
		Instant:   instant,
		Enabled:   enabled,
		MatchType: mt,
	}

	// Event triggers
	if mt >= 100 {
		if ep, ok := encoded["eventParams"].(map[string]interface{}); ok {
			t.EventParams = ep
			// Pre-parse common event params
			if v, ok := ep["exitCodeFilter"].(string); ok {
				t.ExitCode = v
			}
			if v, ok := ep["timeout"].(float64); ok {
				t.Timeout = v
			}
			if v, ok := ep["threshold"].(float64); ok {
				t.Threshold = v
			}
			if v, ok := ep["sequenceId"].(string); ok {
				t.Sequence = v
			}
			if v, ok := ep["messageRegex"].(string); ok {
				t.Regex = v
			}
			if v, ok := ep["directoryRegex"].(string); ok {
				t.Regex = v
			}
			if v, ok := ep["hostRegex"].(string); ok {
				t.Regex = v
			}
			if v, ok := ep["userRegex"].(string); ok {
				t.Regex = v
			}
			if v, ok := ep["commandRegex"].(string); ok {
				t.Regex = v
			}
			if v, ok := ep["progressBarFilter"].(string); ok {
				t.Progress = v
			}
		}
	}

	return t, nil
}

// encodeTriggers serialises a slice of triggers to JSON-compatible list.
func encodeTriggers(triggers []*Trigger) []interface{} {
	result := make([]interface{}, len(triggers))
	for i, t := range triggers {
		result[i] = t.encode()
	}
	return result
}

// Profile-level convenience functions

// GetTriggers reads triggers from the session's profile.
func GetTriggers(ctx context.Context, caller Caller, sessionID string) ([]*Trigger, error) {
	resp, err := GetProfileProperty(ctx, caller, sessionID, []string{"Triggers"})
	if err != nil {
		return nil, fmt.Errorf("get triggers: %w", err)
	}
	props := resp.GetProperties()
	if len(props) == 0 {
		return nil, nil
	}
	raw := props[0].GetJsonValue()
	if raw == "" || raw == "null" {
		return nil, nil
	}
	var encoded []map[string]interface{}
	if err = json.Unmarshal([]byte(raw), &encoded); err != nil {
		return nil, fmt.Errorf("get triggers: %w", err)
	}
	triggers := make([]*Trigger, 0, len(encoded))
	var t *Trigger
	for _, e := range encoded {
		t, err = DecodeTrigger(e)
		if err != nil {
			return nil, fmt.Errorf("get triggers: %w", err)
		}
		triggers = append(triggers, t)
	}
	return triggers, nil
}

// SetTriggers writes triggers to the session's profile.
func SetTriggers(ctx context.Context, caller Caller, sessionID string, triggers []*Trigger) error {
	raw, err := json.Marshal(encodeTriggers(triggers))
	if err != nil {
		return fmt.Errorf("set triggers: %w", err)
	}
	return SetProfileProperty(ctx, caller, sessionID, "Triggers", string(raw))
}

// Actions returns the decoded parameter as actions (for triggers with multiple values).
func (t *Trigger) Actions() []string {
	if t.parsedActions != nil {
		return t.parsedActions
	}
	switch t.Type {
	case TriggerHighlight, TriggerHighlightLine:
		t.parsedActions = parseHighlightParam(t.Param)
	case TriggerPassword:
		t.parsedActions = parsePasswordParam(t.Param)
	case TriggerSetUserVariable:
		t.parsedActions = parseSetUserVariableParam(t.Param)
	default:
		t.parsedActions = []string{t.Param}
	}
	return t.parsedActions
}

func parseHighlightParam(p string) []string {
	p = strings.TrimPrefix(p, "{")
	p = strings.TrimSuffix(p, "}")
	parts := strings.SplitN(p, ",", 2)
	if len(parts) < 2 {
		return []string{"", ""}
	}
	return parts
}

func parsePasswordParam(p string) []string {
	sep := "\u2002\u2014\u2002"
	if strings.Contains(p, sep) {
		return strings.SplitN(p, sep, 2)
	}
	return []string{p, ""}
}

func parseSetUserVariableParam(p string) []string {
	return strings.SplitN(p, "\x01", 2)
}

// IsEvent returns true if this is an event-based trigger (MatchType >= 100).
func (t *Trigger) IsEvent() bool { return t.MatchType >= 100 }

// IsRegex returns true if this is a regex-based trigger (MatchType < 100).
func (t *Trigger) IsRegex() bool { return t.MatchType < 100 }

// String returns a human-readable representation.
func (t *Trigger) String() string {
	return fmt.Sprintf("%s(regex=%q enabled=%v)", t.Type, t.Regex, t.Enabled)
}

// BounceTrigger.Action mapping (convenience)

const (
	BounceUntilActivated = 0
	BounceOnce           = 1
)

// BufferInputTrigger.Action mapping (convenience)

const (
	BufferInputStart = 0
	BufferInputStop  = 1
)

// MarkStopScrolling returns the param value for a MarkTrigger with stop scrolling.
func MarkStopScrolling() string { return "1" }

// MarkNoStopScrolling returns the param value for a MarkTrigger without stop scrolling.
func MarkNoStopScrolling() string { return "0" }

// CommandFinishedEvent exitCodeFilter helpers

const (
	ExitCodeAny     = "*"
	ExitCodeSuccess = "0"
	ExitCodeNonZero = "!0"
)

// ExitCodeFilter returns an exit-code filter string from an int.
func ExitCodeFilter(code int) string { return strconv.Itoa(code) }

// ProgressBarChangedEvent filter helpers

const (
	ProgressAny         = "*"
	ProgressAppeared    = "appeared"
	ProgressDisappeared = "disappeared"
)
