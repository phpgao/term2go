package term2go

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestNewAlertTrigger tests NewAlertTrigger function for creating alert triggers.
func TestNewAlertTrigger(t *testing.T) {
	tr := NewAlertTrigger("error", "Something went wrong")
	assert.Equal(t, TriggerAlert, tr.Type)
	assert.Equal(t, "error", tr.Regex)
	assert.Equal(t, "Something went wrong", tr.Param)
	assert.True(t, tr.Enabled)
	assert.False(t, tr.Instant)
}

// TestNewBounceTrigger tests NewBounceTrigger function for creating bounce triggers.
func TestNewBounceTrigger(t *testing.T) {
	tr := NewBounceTrigger("beep", true)
	assert.Equal(t, TriggerBounce, tr.Type)
	assert.Equal(t, "1", tr.Param) // bounce once

	tr2 := NewBounceTrigger("beep", false)
	assert.Equal(t, "0", tr2.Param) // bounce until activated
}

// TestNewHighlightTrigger tests NewHighlightTrigger function for creating highlight triggers.
func TestNewHighlightTrigger(t *testing.T) {
	tr := NewHighlightTrigger("error", "#FF0000", "#FFFFFF")
	assert.Equal(t, TriggerHighlight, tr.Type)
	assert.Equal(t, "{#FF0000,#FFFFFF}", tr.Param)
	assert.Equal(t, []string{"#FF0000", "#FFFFFF"}, tr.Actions())
}

// TestNewPasswordTrigger tests NewPasswordTrigger function for creating password triggers.
func TestNewPasswordTrigger(t *testing.T) {
	tr := NewPasswordTrigger("password", "admin", "root")
	assert.Equal(t, TriggerPassword, tr.Type)
	assert.Contains(t, tr.Param, "admin")
	assert.Contains(t, tr.Param, "root")
	assert.Equal(t, []string{"admin", "root"}, tr.Actions())
}

// TestNewSetUserVariableTrigger tests NewSetUserVariableTrigger function for creating user variable triggers.
func TestNewSetUserVariableTrigger(t *testing.T) {
	tr := NewSetUserVariableTrigger(".*", "user.test", `"hello"`)
	assert.Equal(t, TriggerSetUserVariable, tr.Type)
	assert.Equal(t, "user.test\x01\"hello\"", tr.Param)
	assert.Equal(t, []string{"user.test", `"hello"`}, tr.Actions())
}

// TestNewCommandFinishedEventTrigger tests NewCommandFinishedEventTrigger function for creating command finished event triggers.
func TestNewCommandFinishedEventTrigger(t *testing.T) {
	tr := NewCommandFinishedEventTrigger(ExitCodeNonZero)
	assert.Equal(t, TriggerCommandFinishedEvent, tr.Type)
	assert.Equal(t, MatchTypeEventCommandFinished, tr.MatchType)
	assert.Equal(t, "!0", tr.EventParams["exitCodeFilter"])
	assert.True(t, tr.IsEvent())
	assert.False(t, tr.IsRegex())
}

// TestNewIdleEventTrigger tests NewIdleEventTrigger function for creating idle event triggers.
func TestNewIdleEventTrigger(t *testing.T) {
	tr := NewIdleEventTrigger(30)
	assert.Equal(t, MatchTypeEventIdle, tr.MatchType)
	assert.Equal(t, 30.0, tr.EventParams["timeout"])
}

// TestNewLongRunningCommandEventTrigger tests NewLongRunningCommandEventTrigger function for creating long running command event triggers.
func TestNewLongRunningCommandEventTrigger(t *testing.T) {
	tr := NewLongRunningCommandEventTrigger(60, "make")
	assert.Equal(t, MatchTypeEventLongRunningCommand, tr.MatchType)
	assert.Equal(t, 60.0, tr.EventParams["threshold"])
	assert.Equal(t, "make", tr.EventParams["commandRegex"])
}

// TestNewProgressBarChangedEventTrigger tests NewProgressBarChangedEventTrigger function for creating progress bar changed event triggers.
func TestNewProgressBarChangedEventTrigger(t *testing.T) {
	tr := NewProgressBarChangedEventTrigger(ProgressAppeared)
	assert.Equal(t, MatchTypeEventProgressBarChanged, tr.MatchType)
	assert.Equal(t, "appeared", tr.EventParams["progressBarFilter"])
}

// TestEncodeAlertTrigger tests Trigger.Encode method for alert triggers.
func TestEncodeAlertTrigger(t *testing.T) {
	tr := NewAlertTrigger("err", "boom")
	m := tr.Encode()
	assert.Equal(t, "err", m["regex"])
	assert.Equal(t, "AlertTrigger", m["action"])
	assert.Equal(t, "boom", m["parameter"])
	assert.Equal(t, 0, m["matchType"])
	assert.Equal(t, false, m["partial"])
	assert.Equal(t, false, m["disabled"])
}

// TestEncodeEventTrigger tests Trigger.Encode method for event triggers.
func TestEncodeEventTrigger(t *testing.T) {
	tr := NewIdleEventTrigger(30)
	m := tr.Encode()
	assert.Equal(t, "", m["regex"])
	assert.Equal(t, 105, m["matchType"])
	assert.Equal(t, false, m["partial"])
	assert.Equal(t, false, m["disabled"])
	ep := m["eventParams"].(map[string]interface{})
	assert.Equal(t, 30.0, ep["timeout"])
}

// TestEncodeDisabledTrigger tests Trigger.Encode method when trigger is disabled.
func TestEncodeDisabledTrigger(t *testing.T) {
	tr := NewBellTrigger("beep")
	tr.Enabled = false
	m := tr.Encode()
	assert.Equal(t, true, m["disabled"])
}

// TestDecodeAlertTrigger tests DecodeTrigger function for alert triggers.
func TestDecodeAlertTrigger(t *testing.T) {
	encoded := map[string]interface{}{
		"regex":     "error",
		"action":    "AlertTrigger",
		"parameter": "boom",
		"partial":   false,
		"disabled":  false,
		"matchType": float64(0),
	}
	tr, err := DecodeTrigger(encoded)
	require.NoError(t, err)
	assert.Equal(t, TriggerAlert, tr.Type)
	assert.Equal(t, "error", tr.Regex)
	assert.Equal(t, "boom", tr.Param)
	assert.True(t, tr.Enabled)
}

// TestDecodeEventTrigger tests DecodeTrigger function for event triggers.
func TestDecodeEventTrigger(t *testing.T) {
	encoded := map[string]interface{}{
		"regex":     "",
		"action":    "IdleEventTrigger",
		"parameter": "",
		"partial":   false,
		"disabled":  false,
		"matchType": float64(105),
		"eventParams": map[string]interface{}{
			"timeout": float64(30),
		},
	}
	tr, err := DecodeTrigger(encoded)
	require.NoError(t, err)
	assert.Equal(t, TriggerIdleEvent, tr.Type)
	assert.Equal(t, float64(30), tr.Timeout)
}

// TestDecodeDisabledTrigger tests DecodeTrigger function for disabled triggers.
func TestDecodeDisabledTrigger(t *testing.T) {
	encoded := map[string]interface{}{
		"regex":     ".*",
		"action":    "BellTrigger",
		"parameter": "",
		"partial":   false,
		"disabled":  true,
		"matchType": float64(0),
	}
	tr, err := DecodeTrigger(encoded)
	require.NoError(t, err)
	assert.False(t, tr.Enabled)
}

// TestDecodeTrigger_Instant tests DecodeTrigger function for instant triggers.
func TestDecodeTrigger_Instant(t *testing.T) {
	encoded := map[string]interface{}{
		"regex":     ".*",
		"action":    "SendTextTrigger",
		"parameter": "hello",
		"partial":   true,
		"disabled":  false,
		"matchType": float64(0),
	}
	tr, err := DecodeTrigger(encoded)
	require.NoError(t, err)
	assert.True(t, tr.Instant)
}

// TestEncodedTriggers tests encodeTriggers function for encoding multiple triggers.
func TestEncodedTriggers(t *testing.T) {
	triggers := []*Trigger{
		NewAlertTrigger("err", "boom"),
		NewBellTrigger("beep"),
		NewIdleEventTrigger(30),
	}
	encoded := encodeTriggers(triggers)
	assert.Len(t, encoded, 3)

	raw, err := json.Marshal(encoded)
	require.NoError(t, err)

	var decoded []map[string]interface{}
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded, 3)

	// Decode back
	for i, d := range decoded {
		tr, err := DecodeTrigger(d)
		require.NoError(t, err)
		assert.Equal(t, triggers[i].Type, tr.Type)
	}
}

// TestGetTriggers tests GetTriggers function for retrieving triggers from a session.
func TestGetTriggers(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{
		resp: successResp(&iterm2.GetProfilePropertyResponse{
			Properties: []*iterm2.ProfileProperty{
				{Key: proto.String("Triggers"), JsonValue: proto.String(`[{"regex":"error","action":"AlertTrigger","parameter":"boom","partial":false,"disabled":false,"matchType":0}]`)},
			},
		}),
	}
	triggers, err := GetTriggers(ctx, mc, "s1")
	require.NoError(t, err)
	require.Len(t, triggers, 1)
	assert.Equal(t, TriggerAlert, triggers[0].Type)
	assert.Equal(t, "error", triggers[0].Regex)
	assert.Equal(t, "boom", triggers[0].Param)
}

// TestSetTriggers tests SetTriggers function for setting triggers on a session.
func TestSetTriggers(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	mc := &mockCaller{resp: &iterm2.ServerOriginatedMessage{}}
	err := SetTriggers(ctx, mc, "s1", []*Trigger{
		NewBellTrigger("beep"),
	})
	require.NoError(t, err)
	req := mc.req.GetSetProfilePropertyRequest()
	require.NotNil(t, req)
	assignments := req.GetAssignments()
	require.Len(t, assignments, 1)
	assert.Equal(t, "Triggers", assignments[0].GetKey())

	// Verify JSON round-trip
	var decoded []map[string]interface{}
	err = json.Unmarshal([]byte(assignments[0].GetJsonValue()), &decoded)
	require.NoError(t, err)
	require.Len(t, decoded, 1)
	assert.Equal(t, "BellTrigger", decoded[0]["action"])
}

// TestAllTriggerTypes tests that all trigger type constants are unique.
func TestAllTriggerTypes(t *testing.T) {
	// Verify all trigger type constants are unique
	types := map[TriggerType]bool{}
	add := func(tt TriggerType) {
		assert.False(t, types[tt], "duplicate trigger type: %s", tt)
		types[tt] = true
	}
	add(TriggerAlert)
	add(TriggerAnnotate)
	add(TriggerBell)
	add(TriggerBounce)
	add(TriggerBufferInput)
	add(TriggerRPC)
	add(TriggerCapture)
	add(TriggerSetNamedMark)
	add(TriggerSGR)
	add(TriggerFold)
	add(TriggerInject)
	add(TriggerHighlightLine)
	add(TriggerHighlight)
	add(TriggerUserNotification)
	add(TriggerSetUserVariable)
	add(TriggerShellPrompt)
	add(TriggerSetTitle)
	add(TriggerSendText)
	add(TriggerRunCommand)
	add(TriggerCoprocess)
	add(TriggerMuteCoprocess)
	add(TriggerMark)
	add(TriggerPassword)
	add(TriggerHyperlink)
	add(TriggerSetDirectory)
	add(TriggerSetHostname)
	add(TriggerStop)
	add(TriggerPromptDetectedEvent)
	add(TriggerCommandFinishedEvent)
	add(TriggerDirectoryChangedEvent)
	add(TriggerHostChangedEvent)
	add(TriggerUserChangedEvent)
	add(TriggerIdleEvent)
	add(TriggerActivityAfterIdleEvent)
	add(TriggerSessionEndedEvent)
	add(TriggerBellReceivedEvent)
	add(TriggerLongRunningCommandEvent)
	add(TriggerCustomEscapeSequenceEvent)
	add(TriggerNotificationPostedEvent)
	add(TriggerProgressBarChangedEvent)
	assert.Len(t, types, 40, "should have 40 unique trigger types")
}

// TestMatchTypeValues tests MatchType constants have correct values.
func TestMatchTypeValues(t *testing.T) {
	assert.Equal(t, TriggerMatchType(0), MatchTypeREGEX)
	assert.Equal(t, TriggerMatchType(100), MatchTypeEventPromptDetected)
	assert.Equal(t, TriggerMatchType(112), MatchTypeEventProgressBarChanged)
}

// TestIsEvent tests Trigger.IsEvent method for detecting event-based triggers.
func TestIsEvent(t *testing.T) {
	assert.True(t, NewIdleEventTrigger(30).IsEvent())
	assert.False(t, NewAlertTrigger("err", "msg").IsEvent())
}

// TestIsRegex tests Trigger.IsRegex method for detecting regex-based triggers.
func TestIsRegex(t *testing.T) {
	assert.True(t, NewSendTextTrigger(".*", "hello").IsRegex())
	assert.False(t, NewSessionEndedEventTrigger().IsRegex())
}

// TestString tests Trigger.String method for generating string representation.
func TestString(t *testing.T) {
	tr := NewAlertTrigger("error", "boom")
	assert.Contains(t, tr.String(), "AlertTrigger")
	assert.Contains(t, tr.String(), "error")
}

// TestNewActivityAfterIdleEventTrigger tests NewActivityAfterIdleEventTrigger factory function.
func TestNewActivityAfterIdleEventTrigger(t *testing.T) {
	tr := NewActivityAfterIdleEventTrigger(45)
	assert.Equal(t, MatchTypeEventActivityAfterIdle, tr.MatchType)
	assert.Equal(t, 45.0, tr.EventParams["timeout"])
	assert.True(t, tr.IsEvent())
}

// TestNewBellReceivedEventTrigger tests NewBellReceivedEventTrigger factory function.
func TestNewBellReceivedEventTrigger(t *testing.T) {
	tr := NewBellReceivedEventTrigger()
	assert.Equal(t, MatchTypeEventBellReceived, tr.MatchType)
	assert.True(t, tr.IsEvent())
}

// TestNewCustomEscapeSequenceEventTrigger tests NewCustomEscapeSequenceEventTrigger factory function.
func TestNewCustomEscapeSequenceEventTrigger(t *testing.T) {
	tr := NewCustomEscapeSequenceEventTrigger("my-id")
	assert.Equal(t, MatchTypeEventCustomEscapeSequence, tr.MatchType)
	assert.Equal(t, "my-id", tr.EventParams["sequenceId"])
}

// TestNewNotificationPostedEventTrigger tests NewNotificationPostedEventTrigger factory function.
func TestNewNotificationPostedEventTrigger(t *testing.T) {
	tr := NewNotificationPostedEventTrigger("pattern")
	assert.Equal(t, MatchTypeEventNotificationPosted, tr.MatchType)
	assert.Equal(t, "pattern", tr.EventParams["messageRegex"])
}

// TestMarkStopScrolling tests MarkStopScrolling helper function.
func TestMarkStopScrolling(t *testing.T) {
	assert.Equal(t, "1", MarkStopScrolling())
}

// TestMarkNoStopScrolling tests MarkNoStopScrolling function for generating mark parameter.
func TestMarkNoStopScrolling(t *testing.T) {
	assert.Equal(t, "0", MarkNoStopScrolling())
}

// TestExitCodeFilter tests ExitCodeFilter function for generating exit code filter strings.
func TestExitCodeFilter(t *testing.T) {
	assert.Equal(t, "42", ExitCodeFilter(42))
	assert.Equal(t, "0", ExitCodeFilter(0))
	assert.Equal(t, "-1", ExitCodeFilter(-1))
}

// TestExitCodeConstants tests exit code constant values.
func TestExitCodeConstants(t *testing.T) {
	assert.Equal(t, "*", ExitCodeAny)
	assert.Equal(t, "0", ExitCodeSuccess)
	assert.Equal(t, "!0", ExitCodeNonZero)
}

// TestProgressConstants tests progress constant values.
func TestProgressConstants(t *testing.T) {
	assert.Equal(t, "*", ProgressAny)
	assert.Equal(t, "appeared", ProgressAppeared)
	assert.Equal(t, "disappeared", ProgressDisappeared)
}

// TestBounceConstants tests bounce constant values.
func TestBounceConstants(t *testing.T) {
	assert.Equal(t, 0, BounceUntilActivated)
	assert.Equal(t, 1, BounceOnce)
}

// TestBufferInputConstants tests buffer input constant values.
func TestBufferInputConstants(t *testing.T) {
	assert.Equal(t, 0, BufferInputStart)
	assert.Equal(t, 1, BufferInputStop)
}

// ============================================================================
// Remaining trigger factory tests (for coverage)
// ============================================================================

// TestNewAnnotateTrigger tests NewAnnotateTrigger factory function.
func TestNewAnnotateTrigger(t *testing.T) {
	tr := NewAnnotateTrigger("regex", "annotation")
	assert.Equal(t, TriggerAnnotate, tr.Type)
	assert.Equal(t, "annotation", tr.Param)
}

// TestNewBufferInputTrigger tests NewBufferInputTrigger factory function.
func TestNewBufferInputTrigger(t *testing.T) {
	tr := NewBufferInputTrigger("regex", true)
	assert.Equal(t, TriggerBufferInput, tr.Type)
	assert.Equal(t, "1", tr.Param)
}

// TestNewRPCTrigger tests NewRPCTrigger factory function.
func TestNewRPCTrigger(t *testing.T) {
	tr := NewRPCTrigger("regex", "my_func()")
	assert.Equal(t, TriggerRPC, tr.Type)
	assert.Equal(t, "my_func()", tr.Param)
}

// TestNewCaptureTrigger tests NewCaptureTrigger factory function.
func TestNewCaptureTrigger(t *testing.T) {
	tr := NewCaptureTrigger("regex", "cmd")
	assert.Equal(t, TriggerCapture, tr.Type)
}

// TestNewSetNamedMarkTrigger tests NewSetNamedMarkTrigger factory function.
func TestNewSetNamedMarkTrigger(t *testing.T) {
	tr := NewSetNamedMarkTrigger("regex", "mark1")
	assert.Equal(t, TriggerSetNamedMark, tr.Type)
}

// TestNewSGRTrigger tests NewSGRTrigger factory function.
func TestNewSGRTrigger(t *testing.T) {
	tr := NewSGRTrigger("regex", "sgr")
	assert.Equal(t, TriggerSGR, tr.Type)
}

// TestNewFoldTrigger tests NewFoldTrigger factory function.
func TestNewFoldTrigger(t *testing.T) {
	tr := NewFoldTrigger("regex", "fold1")
	assert.Equal(t, TriggerFold, tr.Type)
}

// TestNewInjectTrigger tests NewInjectTrigger factory function.
func TestNewInjectTrigger(t *testing.T) {
	tr := NewInjectTrigger("regex", "text")
	assert.Equal(t, TriggerInject, tr.Type)
}

// TestNewHighlightLineTrigger tests NewHighlightLineTrigger factory function.
func TestNewHighlightLineTrigger(t *testing.T) {
	tr := NewHighlightLineTrigger("regex", "#FF0000", "#0000FF")
	assert.Equal(t, TriggerHighlightLine, tr.Type)
	assert.Equal(t, "{#FF0000,#0000FF}", tr.Param)
}

// TestNewUserNotificationTrigger tests NewUserNotificationTrigger factory function.
func TestNewUserNotificationTrigger(t *testing.T) {
	tr := NewUserNotificationTrigger("regex", "msg")
	assert.Equal(t, TriggerUserNotification, tr.Type)
}

// TestNewShellPromptTrigger tests NewShellPromptTrigger factory function.
func TestNewShellPromptTrigger(t *testing.T) {
	tr := NewShellPromptTrigger("regex")
	assert.Equal(t, TriggerShellPrompt, tr.Type)
}

// TestNewSetTitleTrigger tests NewSetTitleTrigger factory function.
func TestNewSetTitleTrigger(t *testing.T) {
	tr := NewSetTitleTrigger("regex", "title")
	assert.Equal(t, TriggerSetTitle, tr.Type)
}

// TestNewRunCommandTrigger tests NewRunCommandTrigger factory function.
func TestNewRunCommandTrigger(t *testing.T) {
	tr := NewRunCommandTrigger("regex", "ls")
	assert.Equal(t, TriggerRunCommand, tr.Type)
}

// TestNewCoprocessTrigger tests NewCoprocessTrigger factory function.
func TestNewCoprocessTrigger(t *testing.T) {
	tr := NewCoprocessTrigger("regex", "cmd")
	assert.Equal(t, TriggerCoprocess, tr.Type)
}

// TestNewMuteCoprocessTrigger tests NewMuteCoprocessTrigger factory function.
func TestNewMuteCoprocessTrigger(t *testing.T) {
	tr := NewMuteCoprocessTrigger("regex", "cmd")
	assert.Equal(t, TriggerMuteCoprocess, tr.Type)
}

// TestNewMarkTrigger tests NewMarkTrigger factory function.
func TestNewMarkTrigger(t *testing.T) {
	tr := NewMarkTrigger("regex", true)
	assert.Equal(t, TriggerMark, tr.Type)
	assert.Equal(t, "1", tr.Param)

	tr2 := NewMarkTrigger("regex", false)
	assert.Equal(t, "0", tr2.Param)
}

// TestNewHyperlinkTrigger tests NewHyperlinkTrigger factory function.
func TestNewHyperlinkTrigger(t *testing.T) {
	tr := NewHyperlinkTrigger("regex", "https://example.com")
	assert.Equal(t, TriggerHyperlink, tr.Type)
}

// TestNewSetDirectoryTrigger tests NewSetDirectoryTrigger factory function.
func TestNewSetDirectoryTrigger(t *testing.T) {
	tr := NewSetDirectoryTrigger("regex", "/tmp")
	assert.Equal(t, TriggerSetDirectory, tr.Type)
}

// TestNewSetHostnameTrigger tests NewSetHostnameTrigger factory function.
func TestNewSetHostnameTrigger(t *testing.T) {
	tr := NewSetHostnameTrigger("regex", "host1")
	assert.Equal(t, TriggerSetHostname, tr.Type)
}

// TestNewStopTrigger tests NewStopTrigger factory function.
func TestNewStopTrigger(t *testing.T) {
	tr := NewStopTrigger("regex")
	assert.Equal(t, TriggerStop, tr.Type)
}

// TestNewPromptDetectedEventTrigger tests NewPromptDetectedEventTrigger factory function.
func TestNewPromptDetectedEventTrigger(t *testing.T) {
	tr := NewPromptDetectedEventTrigger()
	assert.Equal(t, MatchTypeEventPromptDetected, tr.MatchType)
	assert.True(t, tr.IsEvent())
}

// TestNewDirectoryChangedEventTrigger tests NewDirectoryChangedEventTrigger factory function.
func TestNewDirectoryChangedEventTrigger(t *testing.T) {
	tr := NewDirectoryChangedEventTrigger("")
	assert.Equal(t, MatchTypeEventDirectoryChanged, tr.MatchType)
}

// TestNewHostChangedEventTrigger tests NewHostChangedEventTrigger factory function.
func TestNewHostChangedEventTrigger(t *testing.T) {
	tr := NewHostChangedEventTrigger("")
	assert.Equal(t, MatchTypeEventHostChanged, tr.MatchType)
}

// TestNewUserChangedEventTrigger tests NewUserChangedEventTrigger factory function.
func TestNewUserChangedEventTrigger(t *testing.T) {
	tr := NewUserChangedEventTrigger("")
	assert.Equal(t, MatchTypeEventUserChanged, tr.MatchType)
}
