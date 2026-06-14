package term2go

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	iterm2 "github.com/phpgao/term2go/proto"
)

// TestJsonEncode tests jsonEncode function.
func TestJsonEncode(t *testing.T) {
	assert.Equal(t, `"hello"`, jsonEncode("hello"))
	assert.Equal(t, `"hello world"`, jsonEncode("hello world"))
	assert.Equal(t, `"line1\nline2"`, jsonEncode("line1\nline2"))
}

// TestJsonMarshal tests jsonMarshal function.
func TestJsonMarshal(t *testing.T) {
	assert.Equal(t, "42", jsonMarshal(42))
	assert.Equal(t, `"hello"`, jsonMarshal("hello"))
	assert.Equal(t, "[1,2,3]", jsonMarshal([]int{1, 2, 3}))
	assert.Equal(t, "null", jsonMarshal(nil))
}

// TestParseInvokeResult tests parseInvokeResult function.
func TestParseInvokeResult(t *testing.T) {
	t.Run("error disposition", func(t *testing.T) {
		resp := &iterm2.InvokeFunctionResponse{}
		resp.Disposition = &iterm2.InvokeFunctionResponse_Error_{
			Error: &iterm2.InvokeFunctionResponse_Error{
				ErrorReason: strPtr("test error"),
			},
		}
		_, err := parseInvokeResult(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
	})

	t.Run("no success", func(t *testing.T) {
		resp := &iterm2.InvokeFunctionResponse{}
		_, err := parseInvokeResult(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no result")
	})

	t.Run("success with json result", func(t *testing.T) {
		resp := &iterm2.InvokeFunctionResponse{}
		resp.Disposition = &iterm2.InvokeFunctionResponse_Success_{
			Success: &iterm2.InvokeFunctionResponse_Success{
				JsonResult: strPtr(`"hello"`),
			},
		}
		result, err := parseInvokeResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, `"hello"`, result)
	})
}

// TestParseAlertResult tests parseAlertResult function.
func TestParseAlertResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := invokeFuncSuccess("1000")
		n, err := parseAlertResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, n) // button index: 1000-1000=0
	})

	t.Run("button 1", func(t *testing.T) {
		resp := invokeFuncSuccess("1001")
		n, err := parseAlertResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	t.Run("invalid json", func(t *testing.T) {
		resp := invokeFuncSuccess("not-a-number")
		_, err := parseAlertResult(resp)
		assert.Error(t, err)
	})
}

// TestParseTextInputResult tests parseTextInputResult function.
func TestParseTextInputResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := invokeFuncSuccess(`"typed text"`)
		s, err := parseTextInputResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, "typed text", s)
	})

	t.Run("null", func(t *testing.T) {
		resp := invokeFuncSuccess("null")
		s, err := parseTextInputResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, "", s)
	})

	t.Run("invalid json", func(t *testing.T) {
		resp := invokeFuncSuccess("not-json")
		_, err := parseTextInputResult(resp)
		assert.Error(t, err)
	})
}

// TestParseOpenPanelResult tests parseOpenPanelResult function.
func TestParseOpenPanelResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := invokeFuncSuccess(`["/tmp/file.txt"]`)
		s, err := parseOpenPanelResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/file.txt", s)
	})

	t.Run("empty array", func(t *testing.T) {
		resp := invokeFuncSuccess(`[]`)
		_, err := parseOpenPanelResult(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no file selected")
	})

	t.Run("invalid json", func(t *testing.T) {
		resp := invokeFuncSuccess("not-json")
		_, err := parseOpenPanelResult(resp)
		assert.Error(t, err)
	})
}

// TestParseSavePanelResult tests parseSavePanelResult function.
func TestParseSavePanelResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := invokeFuncSuccess(`"/tmp/save.txt"`)
		s, err := parseSavePanelResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, "/tmp/save.txt", s)
	})

	t.Run("null", func(t *testing.T) {
		resp := invokeFuncSuccess("null")
		s, err := parseSavePanelResult(resp)
		assert.NoError(t, err)
		assert.Equal(t, "", s)
	})
}

// TestNewPolyModalAlert tests NewPolyModalAlert function and control methods.
func TestNewPolyModalAlert(t *testing.T) {
	a := NewPolyModalAlert("My Title", "My Subtitle")
	assert.Equal(t, "My Title", a.Title)
	assert.Equal(t, "My Subtitle", a.Subtitle)
	assert.Equal(t, 300, a.Width)

	a.AddButton("OK")
	assert.Equal(t, []string{"OK"}, a.Buttons)

	a.AddCheckbox("Enable", true)
	assert.Equal(t, []string{"Enable"}, a.CheckboxItems)
	assert.Equal(t, []int{1}, a.CheckboxDefaults)

	a.AddComboBox([]string{"A", "B"}, "A")
	assert.Equal(t, []string{"A", "B"}, a.ComboBoxItems)
	assert.Equal(t, "A", a.ComboBoxDefault)

	a.AddTextField("placeholder", "default")
	assert.Equal(t, "placeholder", a.TextFieldLabel)
	assert.Equal(t, "default", a.TextFieldDefault)
}

// TestShowAlert_Success tests ShowAlert function success return.
func TestShowAlert_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("1000")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	buttonIndex, err := ShowAlert(ctx, caller, "Title", "Message", []string{"OK", "Cancel"})

	assert.NoError(t, err)
	assert.Equal(t, 0, buttonIndex)
}

// TestShowAlert_ButtonIndex1 tests ShowAlert function returns button index 1 when second button is clicked.
func TestShowAlert_ButtonIndex1(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("1001")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	buttonIndex, err := ShowAlert(ctx, caller, "Title", "Message", []string{"OK", "Cancel"})

	assert.NoError(t, err)
	assert.Equal(t, 1, buttonIndex)
}

// TestShowAlert_CallerError tests ShowAlert function when caller returns an error.
func TestShowAlert_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowAlert(ctx, caller, "Title", "Message", []string{"OK"})

	assert.Error(t, err)
}

// TestShowAlert_ParseError tests ShowAlert function when the response contains invalid button index.
func TestShowAlert_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("not-a-number")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowAlert(ctx, caller, "Title", "Message", []string{"OK"})

	assert.Error(t, err)
}

// TestShowTextInputAlert_Success tests ShowTextInputAlert function success return.
func TestShowTextInputAlert_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`"typed value"`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	text, err := ShowTextInputAlert(ctx, caller, "Title", "Message", "default")

	assert.NoError(t, err)
	assert.Equal(t, "typed value", text)
}

// TestShowTextInputAlert_Null tests ShowTextInputAlert function returns empty string when user cancels.
func TestShowTextInputAlert_Null(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("null")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	text, err := ShowTextInputAlert(ctx, caller, "Title", "Message", "default")

	assert.NoError(t, err)
	assert.Equal(t, "", text)
}

// TestShowTextInputAlert_CallerError tests ShowTextInputAlert function when caller returns an error.
func TestShowTextInputAlert_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowTextInputAlert(ctx, caller, "Title", "Message", "default")

	assert.Error(t, err)
}

// TestShowTextInputAlert_ParseError tests ShowTextInputAlert function when the response contains invalid JSON.
func TestShowTextInputAlert_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowTextInputAlert(ctx, caller, "Title", "Message", "default")

	assert.Error(t, err)
}

// TestShowOpenPanel_Success tests ShowOpenPanel function success return.
func TestShowOpenPanel_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`["/tmp/file.txt"]`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	filePath, err := ShowOpenPanel(ctx, caller, "Open File", "/tmp")

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/file.txt", filePath)
}

// TestShowOpenPanel_CallerError tests ShowOpenPanel function when caller returns an error.
func TestShowOpenPanel_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowOpenPanel(ctx, caller, "Open File", "/tmp")

	assert.Error(t, err)
}

// TestShowOpenPanel_ParseError tests ShowOpenPanel function when the response contains invalid JSON.
func TestShowOpenPanel_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowOpenPanel(ctx, caller, "Open File", "/tmp")

	assert.Error(t, err)
}

// TestShowOpenPanelWithOptions_Success tests ShowOpenPanelWithOptions function success return.
func TestShowOpenPanelWithOptions_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`["/tmp/a.txt", "/tmp/b.txt"]`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	result, err := ShowOpenPanelWithOptions(ctx, caller, "Open", "Message", "/tmp", "Open", OpenPanelCanCreateDirectories, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/tmp/a.txt", "/tmp/b.txt"}, result.Files)
}

// TestShowOpenPanelWithOptions_EmptyPath tests ShowOpenPanelWithOptions function with empty directory path.
func TestShowOpenPanelWithOptions_EmptyPath(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`["/home/file.txt"]`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	result, err := ShowOpenPanelWithOptions(ctx, caller, "Open", "Message", "", "Open", 0, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/home/file.txt"}, result.Files)
}

// TestShowOpenPanelWithOptions_WithExtensions tests ShowOpenPanelWithOptions function with file extensions filter.
func TestShowOpenPanelWithOptions_WithExtensions(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`["/tmp/test.go"]`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	extensions := []string{"go", "py"}
	result, err := ShowOpenPanelWithOptions(ctx, caller, "Open", "Message", "/tmp", "Open", OpenPanelCanCreateDirectories, extensions)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/tmp/test.go"}, result.Files)
}

// TestShowOpenPanelWithOptions_CallerError tests ShowOpenPanelWithOptions function when caller returns an error.
func TestShowOpenPanelWithOptions_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowOpenPanelWithOptions(ctx, caller, "Open", "Message", "/tmp", "Open", 0, nil)

	assert.Error(t, err)
}

// TestShowOpenPanelWithOptions_ParseError tests ShowOpenPanelWithOptions function when the response contains invalid JSON.
func TestShowOpenPanelWithOptions_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowOpenPanelWithOptions(ctx, caller, "Open", "Message", "/tmp", "Open", 0, nil)

	assert.Error(t, err)
}

// TestShowSavePanel_Success tests ShowSavePanel function success return.
func TestShowSavePanel_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`"/tmp/save.txt"`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	filePath, err := ShowSavePanel(ctx, caller, "Save File", "/tmp")

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/save.txt", filePath)
}

// TestShowSavePanel_Null tests ShowSavePanel function returns empty string when user cancels.
func TestShowSavePanel_Null(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("null")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	filePath, err := ShowSavePanel(ctx, caller, "Save File", "/tmp")

	assert.NoError(t, err)
	assert.Equal(t, "", filePath)
}

// TestShowSavePanel_CallerError tests ShowSavePanel function when caller returns an error.
func TestShowSavePanel_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowSavePanel(ctx, caller, "Save File", "/tmp")

	assert.Error(t, err)
}

// TestShowSavePanel_ParseError tests ShowSavePanel function when the response contains invalid JSON.
func TestShowSavePanel_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowSavePanel(ctx, caller, "Save File", "/tmp")

	assert.Error(t, err)
}

// TestShowSavePanelWithOptions_Success tests ShowSavePanelWithOptions function success return.
func TestShowSavePanelWithOptions_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`"/tmp/output.txt"`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	result, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "/tmp", "Save", "output", "File name:", SavePanelCanCreateDirectories, nil)

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/output.txt", result.File)
}

// TestShowSavePanelWithOptions_Null tests ShowSavePanelWithOptions function returns empty string when user cancels.
func TestShowSavePanelWithOptions_Null(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("null")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	result, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "/tmp", "Save", "", "", 0, nil)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestShowSavePanelWithOptions_EmptyPath tests ShowSavePanelWithOptions function with empty directory path.
func TestShowSavePanelWithOptions_EmptyPath(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`"/home/file.txt"`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	result, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "", "Save", "", "", 0, nil)

	assert.NoError(t, err)
	assert.Equal(t, "/home/file.txt", result.File)
}

// TestShowSavePanelWithOptions_WithExtensions tests ShowSavePanelWithOptions function with file extensions filter.
func TestShowSavePanelWithOptions_WithExtensions(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess(`"/tmp/data.csv"`)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	extensions := []string{"csv", "json"}
	result, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "/tmp", "Save", "data", "Filename:", 0, extensions)

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/data.csv", result.File)
}

// TestShowSavePanelWithOptions_CallerError tests ShowSavePanelWithOptions function when caller returns an error.
func TestShowSavePanelWithOptions_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	_, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "/tmp", "Save", "", "", 0, nil)

	assert.Error(t, err)
}

// TestShowSavePanelWithOptions_ParseError tests ShowSavePanelWithOptions function when the response contains invalid JSON.
func TestShowSavePanelWithOptions_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	_, err := ShowSavePanelWithOptions(ctx, caller, "Save", "Message", "/tmp", "Save", "", "", 0, nil)

	assert.Error(t, err)
}

// TestPolyModalAlert_Run_Success tests PolyModalAlert.Run function success return.
func TestPolyModalAlert_Run_Success(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			jsonResult := `{"button":"OK","tf_text":"user input","combo":"Option B","checks":["Check1","Check2"]}`
			resp := invokeFuncSuccess(jsonResult)
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	alert := NewPolyModalAlert("Title", "Subtitle")
	alert.AddButton("OK")
	alert.AddButton("Cancel")
	alert.AddCheckbox("Check1", true)
	alert.AddCheckbox("Check2", false)
	alert.AddComboBox([]string{"Option A", "Option B"}, "Option A")
	alert.AddTextField("Placeholder", "Default")

	result, err := alert.Run(ctx, caller)

	assert.NoError(t, err)
	assert.Equal(t, "OK", result.Button)
	assert.Equal(t, "user input", result.TextField)
	assert.Equal(t, "Option B", result.ComboBox)
	assert.Equal(t, []string{"Check1", "Check2"}, result.Checkboxes)
}

// TestPolyModalAlert_Run_CallerError tests PolyModalAlert.Run function when caller returns an error.
func TestPolyModalAlert_Run_CallerError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	alert := NewPolyModalAlert("Title", "Subtitle")
	alert.AddButton("OK")

	_, err := alert.Run(ctx, caller)

	assert.Error(t, err)
}

// TestPolyModalAlert_Run_ParseError tests PolyModalAlert.Run function when the response contains invalid JSON.
func TestPolyModalAlert_Run_ParseError(t *testing.T) {
	caller := &alertMockCaller{
		callFunc: func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
			resp := invokeFuncSuccess("invalid-json")
			return successResp(resp), nil
		},
	}

	ctx := context.Background()
	alert := NewPolyModalAlert("Title", "Subtitle")
	alert.AddButton("OK")

	_, err := alert.Run(ctx, caller)

	assert.Error(t, err)
}

// helper: build an InvokeFunctionResponse with a Success JsonResult
func invokeFuncSuccess(jsonResult string) *iterm2.InvokeFunctionResponse {
	return &iterm2.InvokeFunctionResponse{
		Disposition: &iterm2.InvokeFunctionResponse_Success_{
			Success: &iterm2.InvokeFunctionResponse_Success{
				JsonResult: strPtr(jsonResult),
			},
		},
	}
}

func strPtr(s string) *string { return &s }

// alertMockCaller is a test helper that implements Caller interface
// with configurable call behavior for alert tests
type alertMockCaller struct {
	callFunc func(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error)
}

func (m *alertMockCaller) Call(ctx context.Context, req *iterm2.ClientOriginatedMessage) (*iterm2.ServerOriginatedMessage, error) {
	if m.callFunc != nil {
		return m.callFunc(ctx, req)
	}
	return nil, nil
}

func (m *alertMockCaller) Send(req *iterm2.ClientOriginatedMessage) error {
	return nil
}
