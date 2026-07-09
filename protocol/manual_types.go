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

// ThreadStartParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadStartParams struct {
	Model                 *string         `json:"model,omitempty"`
	ModelProvider         *string         `json:"modelProvider,omitempty"`
	ServiceTier           *string         `json:"serviceTier,omitempty"`
	Cwd                   *string         `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage `json:"approvalPolicy,omitempty"`
	Sandbox               json.RawMessage `json:"sandbox,omitempty"`
	Config                *map[string]any `json:"config,omitempty"`
	ServiceName           *string         `json:"serviceName,omitempty"`
	BaseInstructions      *string         `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string         `json:"developerInstructions,omitempty"`
	Ephemeral             *bool           `json:"ephemeral,omitempty"`
}

// ThreadResumeParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadResumeParams struct {
	ThreadID              string          `json:"threadId"`
	Model                 *string         `json:"model,omitempty"`
	ModelProvider         *string         `json:"modelProvider,omitempty"`
	ServiceTier           *string         `json:"serviceTier,omitempty"`
	Cwd                   *string         `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage `json:"approvalPolicy,omitempty"`
	Sandbox               json.RawMessage `json:"sandbox,omitempty"`
	Config                *map[string]any `json:"config,omitempty"`
	BaseInstructions      *string         `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string         `json:"developerInstructions,omitempty"`
}

// ThreadForkParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadForkParams struct {
	ThreadID              string          `json:"threadId"`
	Ephemeral             *bool           `json:"ephemeral,omitempty"`
	Model                 *string         `json:"model,omitempty"`
	ModelProvider         *string         `json:"modelProvider,omitempty"`
	ServiceTier           *string         `json:"serviceTier,omitempty"`
	LastTurnID            *string         `json:"lastTurnId,omitempty"`
	Cwd                   *string         `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage `json:"approvalPolicy,omitempty"`
	Sandbox               json.RawMessage `json:"sandbox,omitempty"`
	Config                *map[string]any `json:"config,omitempty"`
	BaseInstructions      *string         `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string         `json:"developerInstructions,omitempty"`
}

// TurnStartParamsInputElem is maintained manually because turn input entries are
// represented by the high-level codex.Input type before marshaling.
type TurnStartParamsInputElem interface{}

// TurnStartParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type TurnStartParams struct {
	ThreadID            string                     `json:"threadId"`
	Input               []TurnStartParamsInputElem `json:"input"`
	ClientUserMessageID *string                    `json:"clientUserMessageId,omitempty"`
	Cwd                 *string                    `json:"cwd,omitempty"`
	ApprovalPolicy      json.RawMessage            `json:"approvalPolicy,omitempty"`
	SandboxPolicy       json.RawMessage            `json:"sandboxPolicy,omitempty"`
	Model               *string                    `json:"model,omitempty"`
	ServiceTier         *string                    `json:"serviceTier,omitempty"`
	Effort              json.RawMessage            `json:"effort,omitempty"`
	Summary             json.RawMessage            `json:"summary,omitempty"`
	OutputSchema        json.RawMessage            `json:"outputSchema,omitempty"`
}

// TurnSteerParamsInputElem is retained for source compatibility.
type TurnSteerParamsInputElem = UserInput

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

// ThreadGoal describes persisted long-running goal metadata for a thread.
type ThreadGoal struct {
	ThreadID        string           `json:"threadId"`
	Objective       string           `json:"objective"`
	Status          ThreadGoalStatus `json:"status"`
	TokenBudget     *int64           `json:"tokenBudget"`
	TokensUsed      int64            `json:"tokensUsed"`
	TimeUsedSeconds int64            `json:"timeUsedSeconds"`
	CreatedAt       int64            `json:"createdAt"`
	UpdatedAt       int64            `json:"updatedAt"`
}

// ThreadGoalUpdatedNotification is the payload for thread/goal/updated.
type ThreadGoalUpdatedNotification struct {
	ThreadID string     `json:"threadId"`
	TurnID   *string    `json:"turnId"`
	Goal     ThreadGoal `json:"goal"`
}

// ThreadGoalGetResponse is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadGoalGetResponse struct {
	Goal *ThreadGoal `json:"goal"`
}

// ThreadGoalSetResponse is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadGoalSetResponse struct {
	Goal ThreadGoal `json:"goal"`
}

// ApplyPatchApprovalParams is maintained manually to keep FileChanges source
// compatible with older untyped callers.
type ApplyPatchApprovalParams struct {
	CallID         string         `json:"callId"`
	ConversationID string         `json:"conversationId"`
	FileChanges    map[string]any `json:"fileChanges"`
	GrantRoot      *string        `json:"grantRoot,omitempty"`
	Reason         *string        `json:"reason,omitempty"`
}

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
