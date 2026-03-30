package codex

import (
	"encoding/json"
	"errors"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

// ThreadStartOptions configures a thread/start request.
type ThreadStartOptions struct {
	Model string
	Cwd   string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	// Prefer ApprovalPolicy* constants for simple policies.
	ApprovalPolicy any
	// SandboxPolicy is marshaled as JSON and sent as "sandbox".
	// Prefer SandboxMode* constants for simple policies.
	SandboxPolicy         any
	Config                map[string]any
	BaseInstructions      string
	DeveloperInstructions string
	// ExperimentalRawEvents is retained for source compatibility, but the current
	// app-server protocol no longer supports this option. Setting it returns an
	// error from toParams.
	ExperimentalRawEvents bool
}

func (o ThreadStartOptions) toParams() (protocol.ThreadStartParams, error) {
	params := protocol.ThreadStartParams{}
	if o.Model != "" {
		params.Model = stringPtr(o.Model)
	}
	if o.Cwd != "" {
		params.Cwd = stringPtr(o.Cwd)
	}
	if raw, err := normalizeJSONValue("approvalPolicy", o.ApprovalPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.ApprovalPolicy = raw
	}
	if raw, err := normalizeJSONValue("sandbox", o.SandboxPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.Sandbox = raw
	}
	if o.Config != nil {
		config := o.Config
		params.Config = &config
	}
	if o.BaseInstructions != "" {
		params.BaseInstructions = stringPtr(o.BaseInstructions)
	}
	if o.DeveloperInstructions != "" {
		params.DeveloperInstructions = stringPtr(o.DeveloperInstructions)
	}
	if o.ExperimentalRawEvents {
		return params, errors.New("experimental raw events are no longer supported by the current app-server protocol")
	}
	return params, nil
}

// ThreadResumeHistoryElem keeps the old unstable history field compilable for
// callers, but the current app-server protocol no longer accepts history-based
// thread resume.
type ThreadResumeHistoryElem = json.RawMessage

// ThreadResumeOptions configures a thread/resume request.
type ThreadResumeOptions struct {
	// ThreadID resumes a persisted thread by id.
	ThreadID string
	// History is retained for source compatibility, but the current app-server
	// protocol no longer supports history-based resume. Passing History returns an
	// error from toParams.
	History []ThreadResumeHistoryElem
	// Path is retained for source compatibility, but the current app-server
	// protocol no longer supports path-based resume. Passing Path returns an error
	// from toParams.
	Path          string
	Model         string
	ModelProvider string
	Cwd           string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	// Prefer ApprovalPolicy* constants for simple policies.
	ApprovalPolicy any
	// Sandbox is marshaled as JSON and sent as "sandbox".
	// Prefer SandboxMode* constants for simple policies.
	Sandbox               any
	Config                map[string]any
	BaseInstructions      string
	DeveloperInstructions string
}

func (o ThreadResumeOptions) toParams() (protocol.ThreadResumeParams, error) {
	params := protocol.ThreadResumeParams{}
	if o.ThreadID != "" {
		params.ThreadID = o.ThreadID
	}
	if len(o.History) > 0 {
		return params, errors.New("thread resume history is no longer supported by the current app-server protocol")
	}
	if o.Path != "" {
		return params, errors.New("thread resume path is no longer supported by the current app-server protocol")
	}
	if o.Model != "" {
		params.Model = stringPtr(o.Model)
	}
	if o.ModelProvider != "" {
		params.ModelProvider = stringPtr(o.ModelProvider)
	}
	if o.Cwd != "" {
		params.Cwd = stringPtr(o.Cwd)
	}
	if raw, err := normalizeJSONValue("approvalPolicy", o.ApprovalPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.ApprovalPolicy = raw
	}
	if raw, err := normalizeJSONValue("sandbox", o.Sandbox); err != nil {
		return params, err
	} else if raw != nil {
		params.Sandbox = raw
	}
	if o.Config != nil {
		config := o.Config
		params.Config = &config
	}
	if o.BaseInstructions != "" {
		params.BaseInstructions = stringPtr(o.BaseInstructions)
	}
	if o.DeveloperInstructions != "" {
		params.DeveloperInstructions = stringPtr(o.DeveloperInstructions)
	}
	return params, nil
}
