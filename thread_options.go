package codex

import "github.com/pmenglund/codex-sdk-go/protocol"

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
		params.ExperimentalRawEvents = true
	}
	return params, nil
}

// ThreadResumeOptions configures a thread/resume request.
//
// There are three ways to resume a thread:
//  1. ThreadID: load from disk by thread id.
//  2. Path: load from disk by rollout path.
//  3. History: resume from in-memory history.
//
// Precedence is History > Path > ThreadID. Prefer ThreadID when possible.
type ThreadResumeOptions struct {
	// ThreadID resumes a persisted thread by id.
	ThreadID string
	// History is an unstable API used for in-memory resume and takes precedence over ThreadID.
	History []protocol.ThreadResumeParamsHistoryElem
	// Path is an unstable API used for rollout resume and takes precedence over ThreadID.
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
		params.History = make([]protocol.ThreadResumeParamsHistoryElem, 0, len(o.History))
		for _, entry := range o.History {
			if entry == nil {
				params.History = append(params.History, nil)
				continue
			}
			raw, err := normalizeJSONValue("history", entry)
			if err != nil {
				return params, err
			}
			params.History = append(params.History, raw)
		}
	}
	if o.Path != "" {
		params.Path = stringPtr(o.Path)
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
