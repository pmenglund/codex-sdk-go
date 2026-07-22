// Package codex provides an idiomatic Go SDK for the Codex app-server.
//
// The SDK spawns the `codex app-server` process (or uses a custom transport)
// and exposes a high-level facade for accounts, models, threads, turns, and
// streaming turn control. For lower-level access, you can reach the JSON-RPC
// client via (*Codex).Client().
//
// Typical usage:
//
//	ctx := context.Background()
//	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
//	prompt := "Diagnose the test failure and propose a fix"
//	client, err := codex.New(ctx, codex.Options{Logger: logger})
//	if err != nil {
//		panic(err)
//	}
//	defer client.Close()
//
// The constructor context is used for initialization only. Once New returns
// successfully, the spawned app-server lifetime is managed by Close.
//
//	thread, err := client.StartThread(ctx, codex.ThreadStartOptions{})
//	if err != nil {
//		panic(err)
//	}
//
//	result, err := thread.Run(ctx, prompt, nil)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(result.FinalResponse)
//
// For a running turn that needs steering or interruption, start a turn handle:
//
//	handle, err := thread.StartTurn(ctx, []codex.Input{codex.TextInput("Inspect the repo")}, nil)
//	if err != nil {
//		panic(err)
//	}
//	defer handle.Close()
//	_, err = handle.Steer(ctx, []codex.Input{codex.TextInput("Focus on tests")})
//	if err != nil {
//		panic(err)
//	}
//	result, err = handle.Run(ctx)
//	if err != nil {
//		panic(err)
//	}
//
// Account, model, and thread lifecycle helpers cover common app-server calls:
//
//	account, err := client.Account(ctx, codex.AccountOptions{})
//	models, err := client.ListModels(ctx, codex.ListModelsOptions{})
//	threads, err := client.ListThreads(ctx, codex.ThreadListOptions{})
//	_ = account
//	_ = models
//	_ = threads
//
// JSON-typed options (approval policies, sandbox policies, output schemas, etc.)
// accept any JSON-marshalable value. If you already have raw JSON, pass
// json.RawMessage or codex.MustJSON(...) to avoid double encoding.
//
// For common values, prefer typed constants:
//   - codex.ApprovalPolicyNever / codex.ApprovalPolicyOnRequest / ...
//   - codex.SandboxModeReadOnly / codex.SandboxModeWorkspaceWrite / ...
//   - codex.ReasoningEffortLow / codex.ReasoningEffortMedium / ...
//
// Overload errors can be classified with codex.IsOverloaded. The helper works
// with wrapped errors, but callers must decide retry safety from the operation's
// idempotency; classification alone cannot prove that a retry is safe.
package codex
