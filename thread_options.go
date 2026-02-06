package codex

import "github.com/pmenglund/codex-sdk-go/protocol"

// ThreadStartOptions configures a thread/start request.
type ThreadStartOptions struct {
	Model string
	Cwd   string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	ApprovalPolicy any
	// SandboxPolicy is marshaled as JSON and sent as "sandbox".
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
type ThreadResumeOptions struct {
	ThreadID      string
	History       []protocol.ThreadResumeParamsHistoryElem
	Path          string
	Model         string
	ModelProvider string
	Cwd           string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	ApprovalPolicy any
	// Sandbox is marshaled as JSON and sent as "sandbox".
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
