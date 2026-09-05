// Package protocol contains the app-server wire model generated from the Codex
// release identified by GeneratedCodexVersion and GeneratedCodexCommit.
//
// Prefer canonical exported names such as MCPServerOAuthLoginParams. Deprecated
// aliases preserve older acronym spellings for a migration window. Types named
// Sanitized...JSON are generator implementation details retained for source
// compatibility; application code should use the corresponding canonical alias.
// Legacy acronym spellings such as Mcp... and Oauth... are also deprecated in
// favor of MCP... and OAuth....
//
// Discriminated unions such as ThreadItem preserve their complete JSON payload
// and expose Kind, IsKnown, and RawJSON. Unknown future variants therefore round
// trip without forcing callers to depend on generator implementation structs.
// Mixed approval decisions use validated raw-preserving wrappers with checked
// and Must constructors. Fixed choices such as thread sorting and file-change
// decisions use named string types and constants.
package protocol
