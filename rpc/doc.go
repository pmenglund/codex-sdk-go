// Package rpc provides a minimal JSON-RPC client tailored to the Codex app-server.
// Notification subscriptions have a hard caller-selected capacity. A slow
// consumer is closed with NotificationOverflowError without blocking responses
// or other subscribers. The package intentionally implements only the subset needed for Codex.
// For general-purpose JSON-RPC support, consider using a dedicated library.
package rpc
