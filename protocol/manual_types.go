package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ReviewDecisionKind identifies a legacy approval-decision variant.
type ReviewDecisionKind string

const (
	ReviewDecisionKindApproved                    ReviewDecisionKind = "approved"
	ReviewDecisionKindApprovedExecpolicyAmendment ReviewDecisionKind = "approved_execpolicy_amendment"
	ReviewDecisionKindApprovedForSession          ReviewDecisionKind = "approved_for_session"
	ReviewDecisionKindNetworkPolicyAmendment      ReviewDecisionKind = "network_policy_amendment"
	ReviewDecisionKindDenied                      ReviewDecisionKind = "denied"
	ReviewDecisionKindTimedOut                    ReviewDecisionKind = "timed_out"
	ReviewDecisionKindAbort                       ReviewDecisionKind = "abort"
)

// ReviewDecision preserves simple, structured, and future legacy approval
// decisions. Checked constructors accept only known variants; JSON decoding
// retains well-formed future string or single-key object variants.
type ReviewDecision struct {
	raw   json.RawMessage
	kind  ReviewDecisionKind
	known bool
}

// NewReviewDecision validates and wraps a legacy approval decision.
func NewReviewDecision(value any) (ReviewDecision, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return ReviewDecision{}, err
	}
	var decision ReviewDecision
	if err := decision.decode(data, true); err != nil {
		return ReviewDecision{}, err
	}
	return decision, nil
}

// MustReviewDecision is like NewReviewDecision but panics when value is not a
// known decision. It is intended for constants and static policy configuration.
func MustReviewDecision(value any) ReviewDecision {
	decision, err := NewReviewDecision(value)
	if err != nil {
		panic(err)
	}
	return decision
}

func (d *ReviewDecision) UnmarshalJSON(data []byte) error {
	return d.decode(data, false)
}

func (d *ReviewDecision) decode(data []byte, requireKnown bool) error {
	var simple string
	if json.Unmarshal(data, &simple) == nil {
		kind := ReviewDecisionKind(simple)
		known := isSimpleReviewDecisionKind(kind)
		if isStructuredReviewDecisionKind(kind) {
			return fmt.Errorf("review decision %q must use an object representation", simple)
		}
		if requireKnown && !known {
			return fmt.Errorf("unknown review decision %q", simple)
		}
		d.raw, d.kind, d.known = append(json.RawMessage(nil), data...), kind, known
		return nil
	}
	var structured map[string]json.RawMessage
	if err := json.Unmarshal(data, &structured); err != nil || len(structured) != 1 {
		return errors.New("review decision must be a string or single-key object")
	}
	for rawKind, body := range structured {
		kind := ReviewDecisionKind(rawKind)
		known := isStructuredReviewDecisionKind(kind)
		if isSimpleReviewDecisionKind(kind) {
			return fmt.Errorf("review decision %q must use a string representation", kind)
		}
		if requireKnown && !known {
			return fmt.Errorf("unknown review decision %q", kind)
		}
		if known {
			required := "proposed_execpolicy_amendment"
			if kind == ReviewDecisionKindNetworkPolicyAmendment {
				required = "network_policy_amendment"
			}
			var err error
			if kind == ReviewDecisionKindDenied {
				required = "rejection"
				err = requireDecisionStringField(body, required)
			} else {
				err = requireDecisionObjectField(body, required)
			}
			if err != nil {
				return fmt.Errorf("review decision %q: %w", kind, err)
			}
		}
		d.raw, d.kind, d.known = append(json.RawMessage(nil), data...), kind, known
	}
	return nil
}

func isKnownReviewDecisionKind(kind ReviewDecisionKind) bool {
	return isSimpleReviewDecisionKind(kind) || isStructuredReviewDecisionKind(kind)
}

func isSimpleReviewDecisionKind(kind ReviewDecisionKind) bool {
	switch kind {
	case ReviewDecisionKindApproved, ReviewDecisionKindApprovedForSession,
		ReviewDecisionKindTimedOut, ReviewDecisionKindAbort:
		return true
	default:
		return false
	}
}

func isStructuredReviewDecisionKind(kind ReviewDecisionKind) bool {
	switch kind {
	case ReviewDecisionKindApprovedExecpolicyAmendment, ReviewDecisionKindNetworkPolicyAmendment,
		ReviewDecisionKindDenied:
		return true
	default:
		return false
	}
}

// Kind returns the decoded wire variant.
func (d ReviewDecision) Kind() ReviewDecisionKind { return d.kind }

// IsKnown reports whether Kind is defined by the pinned protocol.
func (d ReviewDecision) IsKnown() bool { return d.known }

func (d ReviewDecision) MarshalJSON() ([]byte, error) {
	if len(d.raw) == 0 {
		return []byte("null"), nil
	}
	return append([]byte(nil), d.raw...), nil
}

// RawJSON returns a copy of the complete legacy approval decision.
func (d ReviewDecision) RawJSON() json.RawMessage {
	return append(json.RawMessage(nil), d.raw...)
}

// CommandExecutionApprovalDecisionKind identifies a command approval variant.
type CommandExecutionApprovalDecisionKind string

const (
	CommandExecutionApprovalDecisionKindAccept                        CommandExecutionApprovalDecisionKind = "accept"
	CommandExecutionApprovalDecisionKindAcceptForSession              CommandExecutionApprovalDecisionKind = "acceptForSession"
	CommandExecutionApprovalDecisionKindDecline                       CommandExecutionApprovalDecisionKind = "decline"
	CommandExecutionApprovalDecisionKindCancel                        CommandExecutionApprovalDecisionKind = "cancel"
	CommandExecutionApprovalDecisionKindAcceptWithExecpolicyAmendment CommandExecutionApprovalDecisionKind = "acceptWithExecpolicyAmendment"
	CommandExecutionApprovalDecisionKindApplyNetworkPolicyAmendment   CommandExecutionApprovalDecisionKind = "applyNetworkPolicyAmendment"
)

// CommandExecutionApprovalDecision preserves simple, structured, and future
// command-approval decisions. Checked constructors accept only known variants;
// JSON decoding retains well-formed future string or single-key object variants.
type CommandExecutionApprovalDecision struct {
	raw   json.RawMessage
	kind  CommandExecutionApprovalDecisionKind
	known bool
}

// NewCommandExecutionApprovalDecision validates and wraps a command approval decision.
func NewCommandExecutionApprovalDecision(value any) (CommandExecutionApprovalDecision, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return CommandExecutionApprovalDecision{}, err
	}
	var decision CommandExecutionApprovalDecision
	if err := decision.decode(data, true); err != nil {
		return CommandExecutionApprovalDecision{}, err
	}
	return decision, nil
}

// MustCommandExecutionApprovalDecision is like NewCommandExecutionApprovalDecision
// but panics when value is not a known decision. It is intended for constants
// and static policy configuration.
func MustCommandExecutionApprovalDecision(value any) CommandExecutionApprovalDecision {
	decision, err := NewCommandExecutionApprovalDecision(value)
	if err != nil {
		panic(err)
	}
	return decision
}

func (d *CommandExecutionApprovalDecision) UnmarshalJSON(data []byte) error {
	return d.decode(data, false)
}

func (d *CommandExecutionApprovalDecision) decode(data []byte, requireKnown bool) error {
	var simple string
	if json.Unmarshal(data, &simple) == nil {
		kind := CommandExecutionApprovalDecisionKind(simple)
		known := isSimpleCommandExecutionApprovalDecisionKind(kind)
		if isStructuredCommandExecutionApprovalDecisionKind(kind) {
			return fmt.Errorf("command approval decision %q must use an object representation", simple)
		}
		if requireKnown && !known {
			return fmt.Errorf("unknown command approval decision %q", simple)
		}
		d.raw, d.kind, d.known = append(json.RawMessage(nil), data...), kind, known
		return nil
	}
	var structured map[string]json.RawMessage
	if err := json.Unmarshal(data, &structured); err != nil || len(structured) != 1 {
		return errors.New("command approval decision must be a string or single-key object")
	}
	for rawKind, body := range structured {
		kind := CommandExecutionApprovalDecisionKind(rawKind)
		known := isStructuredCommandExecutionApprovalDecisionKind(kind)
		if isSimpleCommandExecutionApprovalDecisionKind(kind) {
			return fmt.Errorf("command approval decision %q must use a string representation", kind)
		}
		if requireKnown && !known {
			return fmt.Errorf("unknown command approval decision %q", kind)
		}
		if known {
			required := "execpolicyAmendment"
			if kind == CommandExecutionApprovalDecisionKindApplyNetworkPolicyAmendment {
				required = "networkPolicyAmendment"
			}
			if err := requireDecisionObjectField(body, required); err != nil {
				return fmt.Errorf("command approval decision %q: %w", kind, err)
			}
		}
		d.raw, d.kind, d.known = append(json.RawMessage(nil), data...), kind, known
	}
	return nil
}

func isKnownCommandExecutionApprovalDecisionKind(kind CommandExecutionApprovalDecisionKind) bool {
	return isSimpleCommandExecutionApprovalDecisionKind(kind) || isStructuredCommandExecutionApprovalDecisionKind(kind)
}

func isSimpleCommandExecutionApprovalDecisionKind(kind CommandExecutionApprovalDecisionKind) bool {
	switch kind {
	case CommandExecutionApprovalDecisionKindAccept, CommandExecutionApprovalDecisionKindAcceptForSession,
		CommandExecutionApprovalDecisionKindDecline, CommandExecutionApprovalDecisionKindCancel:
		return true
	default:
		return false
	}
}

func isStructuredCommandExecutionApprovalDecisionKind(kind CommandExecutionApprovalDecisionKind) bool {
	switch kind {
	case CommandExecutionApprovalDecisionKindAcceptWithExecpolicyAmendment,
		CommandExecutionApprovalDecisionKindApplyNetworkPolicyAmendment:
		return true
	default:
		return false
	}
}

func requireDecisionObjectField(body json.RawMessage, field string) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return errors.New("variant payload must be an object")
	}
	value, ok := object[field]
	if !ok || string(value) == "null" {
		return fmt.Errorf("required field %q is missing", field)
	}
	return nil
}

func requireDecisionStringField(body json.RawMessage, field string) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return errors.New("variant payload must be an object")
	}
	value, ok := object[field]
	if !ok {
		return fmt.Errorf("required field %q is missing", field)
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil || text == "" {
		return fmt.Errorf("required field %q must be a non-empty string", field)
	}
	return nil
}

// Kind returns the decoded wire variant.
func (d CommandExecutionApprovalDecision) Kind() CommandExecutionApprovalDecisionKind {
	return d.kind
}

// IsKnown reports whether Kind is defined by the pinned protocol.
func (d CommandExecutionApprovalDecision) IsKnown() bool { return d.known }

func (d CommandExecutionApprovalDecision) MarshalJSON() ([]byte, error) {
	if len(d.raw) == 0 {
		return []byte("null"), nil
	}
	return append([]byte(nil), d.raw...), nil
}

// RawJSON returns a copy of the complete command approval decision.
func (d CommandExecutionApprovalDecision) RawJSON() json.RawMessage {
	return append(json.RawMessage(nil), d.raw...)
}

// FileChangeApprovalDecision is a fixed file-change approval choice.
type FileChangeApprovalDecision string

const (
	FileChangeApprovalDecisionAccept           FileChangeApprovalDecision = "accept"
	FileChangeApprovalDecisionAcceptForSession FileChangeApprovalDecision = "acceptForSession"
	FileChangeApprovalDecisionDecline          FileChangeApprovalDecision = "decline"
	FileChangeApprovalDecisionCancel           FileChangeApprovalDecision = "cancel"
)

// MCPServerElicitationRequestParams preserves the complete mixed elicitation
// request payload. Its variants do not share a stable discriminator.
type MCPServerElicitationRequestParams = json.RawMessage

// ThreadExtra preserves implementation-specific thread metadata.
type ThreadExtra map[string]json.RawMessage

// GitInfo describes the source-control state captured for a thread.
type GitInfo struct {
	SHA       *string `json:"sha,omitempty"`
	Branch    *string `json:"branch,omitempty"`
	OriginURL *string `json:"originUrl,omitempty"`
}

// TurnStatus describes the current state of a turn.
type TurnStatus string

const (
	TurnStatusCompleted   TurnStatus = "completed"
	TurnStatusInterrupted TurnStatus = "interrupted"
	TurnStatusFailed      TurnStatus = "failed"
	TurnStatusInProgress  TurnStatus = "inProgress"
)

// TurnItemsView describes how much of a turn's item history is present.
type TurnItemsView string

const (
	TurnItemsViewNotLoaded TurnItemsView = "notLoaded"
	TurnItemsViewSummary   TurnItemsView = "summary"
	TurnItemsViewFull      TurnItemsView = "full"
)

// Turn is the complete turn summary returned inside thread payloads.
type Turn struct {
	ID          string        `json:"id"`
	Items       []ThreadItem  `json:"items"`
	ItemsView   TurnItemsView `json:"itemsView,omitempty"`
	Status      TurnStatus    `json:"status"`
	Error       *TurnError    `json:"error,omitempty"`
	StartedAt   *int64        `json:"startedAt,omitempty"`
	CompletedAt *int64        `json:"completedAt,omitempty"`
	DurationMs  *int64        `json:"durationMs,omitempty"`
}

// Thread represents the complete thread summary returned by the app-server.
type Thread struct {
	ID             string            `json:"id"`
	Extra          *ThreadExtra      `json:"extra,omitempty"`
	SessionID      string            `json:"sessionId"`
	ForkedFromID   *string           `json:"forkedFromId,omitempty"`
	ParentThreadID *string           `json:"parentThreadId,omitempty"`
	Preview        string            `json:"preview"`
	Ephemeral      bool              `json:"ephemeral"`
	HistoryMode    ThreadHistoryMode `json:"historyMode,omitempty"`
	ModelProvider  string            `json:"modelProvider"`
	CreatedAt      int64             `json:"createdAt"`
	UpdatedAt      int64             `json:"updatedAt"`
	RecencyAt      *int64            `json:"recencyAt,omitempty"`
	Status         ThreadStatus      `json:"status"`
	Path           *string           `json:"path,omitempty"`
	Cwd            string            `json:"cwd"`
	CLIVersion     string            `json:"cliVersion"`
	Source         json.RawMessage   `json:"source"`
	ThreadSource   *ThreadSource     `json:"threadSource,omitempty"`
	AgentNickname  *string           `json:"agentNickname,omitempty"`
	AgentRole      *string           `json:"agentRole,omitempty"`
	GitInfo        *GitInfo          `json:"gitInfo,omitempty"`
	Name           *string           `json:"name,omitempty"`
	Turns          []Turn            `json:"turns"`
}

// ThreadStartResponse is the response payload for thread/start.
type ThreadStartResponse struct {
	ThreadID                string           `json:"threadId,omitempty"`
	Thread                  *Thread          `json:"thread"`
	Model                   string           `json:"model"`
	ModelProvider           string           `json:"modelProvider"`
	ServiceTier             *string          `json:"serviceTier,omitempty"`
	Cwd                     string           `json:"cwd"`
	RuntimeWorkspaceRoots   []string         `json:"runtimeWorkspaceRoots,omitempty"`
	InstructionSources      []string         `json:"instructionSources,omitempty"`
	ApprovalPolicy          json.RawMessage  `json:"approvalPolicy"`
	ApprovalsReviewer       json.RawMessage  `json:"approvalsReviewer"`
	Sandbox                 SandboxPolicy    `json:"sandbox"`
	ActivePermissionProfile json.RawMessage  `json:"activePermissionProfile,omitempty"`
	ReasoningEffort         *ReasoningEffort `json:"reasoningEffort,omitempty"`
	MultiAgentMode          json.RawMessage  `json:"multiAgentMode,omitempty"`
}

// ThreadResponse is the former shared lifecycle response shape.
//
// Deprecated: use the method-specific ThreadStartResponse,
// ThreadResumeResponse, or ThreadForkResponse.
type ThreadResponse = ThreadStartResponse

// ThreadResumeResponse is the response payload for thread/resume.
type ThreadResumeResponse struct {
	ThreadID                string           `json:"threadId,omitempty"`
	Thread                  *Thread          `json:"thread"`
	Model                   string           `json:"model"`
	ModelProvider           string           `json:"modelProvider"`
	ServiceTier             *string          `json:"serviceTier,omitempty"`
	Cwd                     string           `json:"cwd"`
	RuntimeWorkspaceRoots   []string         `json:"runtimeWorkspaceRoots,omitempty"`
	InstructionSources      []string         `json:"instructionSources,omitempty"`
	ApprovalPolicy          json.RawMessage  `json:"approvalPolicy"`
	ApprovalsReviewer       json.RawMessage  `json:"approvalsReviewer"`
	Sandbox                 SandboxPolicy    `json:"sandbox"`
	ActivePermissionProfile json.RawMessage  `json:"activePermissionProfile,omitempty"`
	ReasoningEffort         *ReasoningEffort `json:"reasoningEffort,omitempty"`
	MultiAgentMode          json.RawMessage  `json:"multiAgentMode,omitempty"`
	InitialTurnsPage        *TurnsPage       `json:"initialTurnsPage,omitempty"`
}

// ThreadForkResponse is the response payload for thread/fork.
type ThreadForkResponse struct {
	ThreadID                string           `json:"threadId,omitempty"`
	Thread                  *Thread          `json:"thread"`
	Model                   string           `json:"model"`
	ModelProvider           string           `json:"modelProvider"`
	ServiceTier             *string          `json:"serviceTier,omitempty"`
	Cwd                     string           `json:"cwd"`
	RuntimeWorkspaceRoots   []string         `json:"runtimeWorkspaceRoots,omitempty"`
	InstructionSources      []string         `json:"instructionSources,omitempty"`
	ApprovalPolicy          json.RawMessage  `json:"approvalPolicy"`
	ApprovalsReviewer       json.RawMessage  `json:"approvalsReviewer"`
	Sandbox                 SandboxPolicy    `json:"sandbox"`
	ActivePermissionProfile json.RawMessage  `json:"activePermissionProfile,omitempty"`
	ReasoningEffort         *ReasoningEffort `json:"reasoningEffort,omitempty"`
	MultiAgentMode          json.RawMessage  `json:"multiAgentMode,omitempty"`
}

// TurnsPage is a typed page of turns returned while resuming a thread.
type TurnsPage struct {
	Data            []Turn  `json:"data"`
	NextCursor      *string `json:"nextCursor,omitempty"`
	BackwardsCursor *string `json:"backwardsCursor,omitempty"`
}

// ThreadListResponse is a page of persisted threads.
type ThreadListResponse struct {
	Data            []Thread `json:"data"`
	NextCursor      *string  `json:"nextCursor,omitempty"`
	BackwardsCursor *string  `json:"backwardsCursor,omitempty"`
}

// ThreadReadResponse is the response payload for thread/read.
type ThreadReadResponse struct {
	Thread Thread `json:"thread"`
}

// ThreadUnarchiveResponse is the response payload for thread/unarchive.
type ThreadUnarchiveResponse struct {
	Thread Thread `json:"thread"`
}

// TurnStartResponse is the response payload for turn/start.
type TurnStartResponse struct {
	Turn Turn `json:"turn"`
}

// ThreadListParams configures thread/list using typed sort values.
type ThreadListParams struct {
	Archived       *bool                           `json:"archived,omitempty"`
	Cursor         *string                         `json:"cursor,omitempty"`
	Cwd            any                             `json:"cwd,omitempty"`
	Limit          *int                            `json:"limit,omitempty"`
	ModelProviders *ThreadListParamsModelProviders `json:"modelProviders,omitempty"`
	SearchTerm     *string                         `json:"searchTerm,omitempty"`
	SortDirection  SortDirection                   `json:"sortDirection,omitempty"`
	SortKey        ThreadSortKey                   `json:"sortKey,omitempty"`
	SourceKinds    *ThreadListParamsSourceKinds    `json:"sourceKinds,omitempty"`
	UseStateDbOnly *bool                           `json:"useStateDbOnly,omitempty"`
}

// ThreadStartParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadStartParams struct {
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	ServiceTier           *string            `json:"serviceTier,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	Sandbox               json.RawMessage    `json:"sandbox,omitempty"`
	Config                *map[string]any    `json:"config,omitempty"`
	ServiceName           *string            `json:"serviceName,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	Ephemeral             *bool              `json:"ephemeral,omitempty"`
	Personality           *Personality       `json:"personality,omitempty"`
	SessionStartSource    *ThreadStartSource `json:"sessionStartSource,omitempty"`
	ThreadSource          *ThreadSource      `json:"threadSource,omitempty"`
}

// ThreadResumeParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadResumeParams struct {
	ThreadID              string             `json:"threadId"`
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	ServiceTier           *string            `json:"serviceTier,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	Sandbox               json.RawMessage    `json:"sandbox,omitempty"`
	Config                *map[string]any    `json:"config,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	Personality           *Personality       `json:"personality,omitempty"`
}

// ThreadForkParams is maintained manually because the raw schema currently
// exceeds the generator's capabilities.
type ThreadForkParams struct {
	ThreadID              string             `json:"threadId"`
	Ephemeral             *bool              `json:"ephemeral,omitempty"`
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	ServiceTier           *string            `json:"serviceTier,omitempty"`
	LastTurnID            *string            `json:"lastTurnId,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	ApprovalPolicy        json.RawMessage    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	Sandbox               json.RawMessage    `json:"sandbox,omitempty"`
	Config                *map[string]any    `json:"config,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	ThreadSource          *ThreadSource      `json:"threadSource,omitempty"`
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
	ApprovalsReviewer   *ApprovalsReviewer         `json:"approvalsReviewer,omitempty"`
	SandboxPolicy       json.RawMessage            `json:"sandboxPolicy,omitempty"`
	Model               *string                    `json:"model,omitempty"`
	ServiceTier         *string                    `json:"serviceTier,omitempty"`
	Effort              json.RawMessage            `json:"effort,omitempty"`
	Summary             json.RawMessage            `json:"summary,omitempty"`
	OutputSchema        json.RawMessage            `json:"outputSchema,omitempty"`
	Personality         *Personality               `json:"personality,omitempty"`
}

// TurnSteerParamsInputElem is retained for source compatibility.
type TurnSteerParamsInputElem = UserInput

// TurnNotification describes turn/started and turn/completed notifications.
type TurnNotification struct {
	ThreadID string `json:"threadId,omitempty"`
	Turn     *Turn  `json:"turn,omitempty"`
}

// TurnNotificationTurn is the former name of Turn.
//
// Deprecated: use Turn.
type TurnNotificationTurn = Turn

// TurnNotificationError is the former name of TurnError.
//
// Deprecated: use TurnError.
type TurnNotificationError = TurnError

// TurnStartedNotification is the payload for turn/started.
type TurnStartedNotification = TurnNotification

// TurnCompletedNotification is the payload for turn/completed.
type TurnCompletedNotification = TurnNotification

// ItemCompletedNotification is the payload for item/completed.
type ItemCompletedNotification struct {
	ThreadID      string     `json:"threadId"`
	TurnID        string     `json:"turnId"`
	Item          ThreadItem `json:"item"`
	CompletedAtMs int64      `json:"completedAtMs"`
}

// ErrorNotification is the payload for error notifications.
type ErrorNotification struct {
	ThreadID  string    `json:"threadId"`
	TurnID    string    `json:"turnId"`
	WillRetry bool      `json:"willRetry"`
	Error     TurnError `json:"error"`
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
	ThreadID      string  `json:"threadId"`
	TurnID        string  `json:"turnId"`
	ItemID        string  `json:"itemId"`
	StartedAtMs   int64   `json:"startedAtMs"`
	EnvironmentID *string `json:"environmentId"`

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
	ThreadID      string          `json:"threadId"`
	TurnID        string          `json:"turnId"`
	ItemID        string          `json:"itemId"`
	EnvironmentID *string         `json:"environmentId"`
	StartedAtMs   int64           `json:"startedAtMs"`
	Cwd           AbsolutePathBuf `json:"cwd"`

	Reason      *string     `json:"reason"`
	Permissions interface{} `json:"permissions"`
}

// PermissionsRequestApprovalResponse is maintained manually because the raw
// schema uses nested unions that the generator does not currently emit.
type PermissionsRequestApprovalResponse struct {
	Permissions      interface{} `json:"permissions"`
	Scope            interface{} `json:"scope,omitempty"`
	StrictAutoReview *bool       `json:"strictAutoReview,omitempty"`
}
