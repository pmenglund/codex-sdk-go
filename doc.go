// Package codex provides an idiomatic Go SDK for the Codex app-server.
//
// The SDK spawns the `codex app-server` process (or uses a custom transport)
// and exposes a high-level facade for threads and turns. For lower-level access,
// you can reach the JSON-RPC client via (*Codex).Client().
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
// JSON-typed options (approval policies, sandbox policies, output schemas, etc.)
// accept any JSON-marshalable value. If you already have raw JSON, pass
// json.RawMessage or codex.MustJSON(...) to avoid double encoding.
package codex
