package codex

import "github.com/pmenglund/codex-sdk-go/protocol"

// ApprovalPolicy is a typed alias for common approval policy values.
type ApprovalPolicy = protocol.AskForApproval

const (
	ApprovalPolicyNever     ApprovalPolicy = protocol.AskForApprovalNever
	ApprovalPolicyOnFailure ApprovalPolicy = protocol.AskForApprovalOnFailure
	ApprovalPolicyOnRequest ApprovalPolicy = protocol.AskForApprovalOnRequest
	ApprovalPolicyUntrusted ApprovalPolicy = protocol.AskForApprovalUntrusted
)

// SandboxMode is a typed alias for simple sandbox mode values.
type SandboxMode = protocol.SandboxMode

const (
	SandboxModeReadOnly         SandboxMode = protocol.SandboxModeReadOnly
	SandboxModeWorkspaceWrite   SandboxMode = protocol.SandboxModeWorkspaceWrite
	SandboxModeDangerFullAccess SandboxMode = protocol.SandboxModeDangerFullAccess
)

// ReasoningEffort is a typed alias for standard effort values.
type ReasoningEffort = protocol.ReasoningEffort

const (
	ReasoningEffortNone    ReasoningEffort = protocol.ReasoningEffortNone
	ReasoningEffortMinimal ReasoningEffort = protocol.ReasoningEffortMinimal
	ReasoningEffortLow     ReasoningEffort = protocol.ReasoningEffortLow
	ReasoningEffortMedium  ReasoningEffort = protocol.ReasoningEffortMedium
	ReasoningEffortHigh    ReasoningEffort = protocol.ReasoningEffortHigh
	ReasoningEffortXHigh   ReasoningEffort = protocol.ReasoningEffortXhigh
)
