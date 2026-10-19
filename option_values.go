package codex

import "github.com/pmenglund/codex-sdk-go/protocol"

// ApprovalPolicy is a typed alias for common approval policy values.
type ApprovalPolicy = string

const (
	// ApprovalPolicyNever selects the never-ask approval policy.
	ApprovalPolicyNever ApprovalPolicy = "never"
	// ApprovalPolicyOnFailure selects the on-failure approval policy.
	ApprovalPolicyOnFailure ApprovalPolicy = "on-failure"
	// ApprovalPolicyOnRequest lets the app-server request approval when needed.
	ApprovalPolicyOnRequest ApprovalPolicy = "on-request"
	// ApprovalPolicyUntrusted requests approval for commands outside the trusted set.
	ApprovalPolicyUntrusted ApprovalPolicy = "untrusted"
)

// SandboxMode is a typed alias for simple sandbox mode values.
type SandboxMode = protocol.SandboxMode

const (
	// SandboxModeReadOnly selects the read-only sandbox.
	SandboxModeReadOnly SandboxMode = protocol.SandboxModeReadOnly
	// SandboxModeWorkspaceWrite permits writes within the workspace sandbox.
	SandboxModeWorkspaceWrite SandboxMode = protocol.SandboxModeWorkspaceWrite
	// SandboxModeDangerFullAccess disables sandbox restrictions.
	SandboxModeDangerFullAccess SandboxMode = protocol.SandboxModeDangerFullAccess
)

// ReasoningEffort is a typed alias for standard effort values.
type ReasoningEffort = protocol.ReasoningEffort

const (
	// ReasoningEffortNone requests no model reasoning effort.
	ReasoningEffortNone ReasoningEffort = "none"
	// ReasoningEffortMinimal requests minimal model reasoning effort.
	ReasoningEffortMinimal ReasoningEffort = "minimal"
	// ReasoningEffortLow requests low model reasoning effort.
	ReasoningEffortLow ReasoningEffort = "low"
	// ReasoningEffortMedium requests medium model reasoning effort.
	ReasoningEffortMedium ReasoningEffort = "medium"
	// ReasoningEffortHigh requests high model reasoning effort.
	ReasoningEffortHigh ReasoningEffort = "high"
	// ReasoningEffortXHigh requests extra-high model reasoning effort.
	ReasoningEffortXHigh ReasoningEffort = "xhigh"
)
