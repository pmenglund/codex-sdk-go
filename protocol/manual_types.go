package protocol

import "encoding/json"

// Thread represents a minimal thread descriptor used in thread responses.
type Thread struct {
	ID string `json:"id,omitempty"`
}

// ThreadResponse is the shared shape for thread/start and thread/resume responses.
type ThreadResponse struct {
	ThreadID string  `json:"threadId,omitempty"`
	Thread   *Thread `json:"thread,omitempty"`
}

// ThreadStartResponse is the response payload for thread/start.
type ThreadStartResponse = ThreadResponse

// ThreadResumeResponse is the response payload for thread/resume.
type ThreadResumeResponse = ThreadResponse

// TurnNotification describes turn/started and turn/completed notifications.
type TurnNotification struct {
	ThreadID string                `json:"threadId,omitempty"`
	Turn     *TurnNotificationTurn `json:"turn,omitempty"`
}

// TurnStartedNotification is the payload for turn/started.
type TurnStartedNotification = TurnNotification

// TurnCompletedNotification is the payload for turn/completed.
type TurnCompletedNotification = TurnNotification

// TurnNotificationTurn describes a turn summary in notifications.
type TurnNotificationTurn struct {
	ID     string                 `json:"id,omitempty"`
	Status string                 `json:"status,omitempty"`
	Error  *TurnNotificationError `json:"error,omitempty"`
}

// TurnNotificationError describes a turn error payload.
type TurnNotificationError struct {
	Message string `json:"message,omitempty"`
}

// ItemCompletedNotification is the payload for item/completed.
type ItemCompletedNotification struct {
	ThreadID string          `json:"threadId,omitempty"`
	Item     json.RawMessage `json:"item,omitempty"`
}

// ErrorNotification is the payload for error notifications.
type ErrorNotification struct {
	ThreadID  string                 `json:"threadId,omitempty"`
	WillRetry *bool                  `json:"willRetry,omitempty"`
	Error     *TurnNotificationError `json:"error,omitempty"`
}

// ApplyPatchApprovalParams uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ApplyPatchApprovalParams = SanitizedApplyPatchApprovalParams

// ApplyPatchApprovalResponse uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ApplyPatchApprovalResponse = SanitizedApplyPatchApprovalResponse

// ExecCommandApprovalParams uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ExecCommandApprovalParams = SanitizedExecCommandApprovalParams

// ExecCommandApprovalResponse uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ExecCommandApprovalResponse = SanitizedExecCommandApprovalResponse

// FileChangeRequestApprovalParams uses the sanitized schema variant because the
// raw schema currently exceeds the generator's capabilities.
type FileChangeRequestApprovalParams = SanitizedFileChangeRequestApprovalParams

// FileChangeRequestApprovalResponse uses the sanitized schema variant because
// the raw schema currently exceeds the generator's capabilities.
type FileChangeRequestApprovalResponse = SanitizedFileChangeRequestApprovalResponse

// ToolRequestUserInputParams uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ToolRequestUserInputParams = SanitizedToolRequestUserInputParams

// ToolRequestUserInputResponse uses the sanitized schema variant because the raw
// schema currently exceeds the generator's capabilities.
type ToolRequestUserInputResponse = SanitizedToolRequestUserInputResponse

// CommandExecutionRequestApprovalParams is maintained manually because the raw
// schema uses nested unions that the generator does not currently emit.
type CommandExecutionRequestApprovalParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
	ItemID   string `json:"itemId"`

	ApprovalID *string `json:"approvalId,omitempty"`
	Reason     *string `json:"reason,omitempty"`

	NetworkApprovalContext          interface{}                        `json:"networkApprovalContext,omitempty"`
	Command                         *string                            `json:"command,omitempty"`
	Cwd                             *string                            `json:"cwd,omitempty"`
	CommandActions                  []interface{}                      `json:"commandActions,omitempty"`
	AdditionalPermissions           interface{}                        `json:"additionalPermissions,omitempty"`
	ProposedExecpolicyAmendment     []string                           `json:"proposedExecpolicyAmendment,omitempty"`
	ProposedNetworkPolicyAmendments []NetworkPolicyAmendment           `json:"proposedNetworkPolicyAmendments,omitempty"`
	AvailableDecisions              []CommandExecutionApprovalDecision `json:"availableDecisions,omitempty"`
}

// CommandExecutionRequestApprovalResponse is maintained manually because the raw
// schema uses nested unions that the generator does not currently emit.
type CommandExecutionRequestApprovalResponse struct {
	Decision CommandExecutionApprovalDecision `json:"decision"`
}

// PermissionsRequestApprovalParams is maintained manually because the raw
// schema uses nested unions that the generator does not currently emit.
type PermissionsRequestApprovalParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
	ItemID   string `json:"itemId"`

	Reason      *string     `json:"reason,omitempty"`
	Permissions interface{} `json:"permissions"`
}

// PermissionsRequestApprovalResponse is maintained manually because the raw
// schema uses nested unions that the generator does not currently emit.
type PermissionsRequestApprovalResponse struct {
	Permissions interface{} `json:"permissions"`
	Scope       interface{} `json:"scope,omitempty"`
}
