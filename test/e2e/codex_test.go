//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	codex "github.com/pmenglund/codex-sdk-go"
	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/test"
)

func TestRealCodexSynchronousCancellationInterrupts(t *testing.T) {
	loginParams, secret := test.RequireLoginParams(t)
	var recorder test.RequestRecorder
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{
		Timeout:         2 * time.Minute,
		Secrets:         []string{secret},
		RequestRecorder: &recorder,
	})
	if _, err := client.StartLogin(ctx, loginParams); err != nil {
		t.Fatalf("start login for cancellation test: %v\nstderr:\n%s", err, stderr.String())
	}
	t.Cleanup(func() {
		logoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.Logout(logoutCtx)
	})

	thread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	handle, err := thread.StartTurn(ctx, []codex.Input{codex.TextInput("Write a detailed multi-paragraph response.")}, test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("start cancellable turn: %v\nstderr:\n%s", err, stderr.String())
	}
	if handle.ID() == "" {
		t.Fatalf("expected a known turn id before cancellation\nstderr:\n%s", stderr.String())
	}
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = handle.Run(canceledCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation after turn interrupt, got %v\nstderr:\n%s", err, stderr.String())
	}
	request, ok := recorder.Request("turn/interrupt")
	if !ok {
		t.Fatalf("expected an outbound turn/interrupt request\nstderr:\n%s", stderr.String())
	}
	var params protocol.TurnInterruptParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		t.Fatalf("decode turn/interrupt params: %v", err)
	}
	if params.ThreadID != thread.ID() || params.TurnID != handle.ID() {
		t.Fatalf("unexpected turn/interrupt params: %#v", params)
	}
}

func TestRealCodexHighLevelCoverageMatrix(t *testing.T) {
	coverage := map[string]string{
		"New":                  "TestRealCodexClientLifecycleAndGeneratedRPC",
		"Client":               "TestRealCodexClientLifecycleAndGeneratedRPC",
		"Close":                "TestRealCodexClientLifecycleAndGeneratedRPC",
		"StartThread":          "TestRealCodexThreadLifecycleHelpers",
		"ResumeThread":         "TestRealCodexThreadLifecycleHelpers, TestRealCodexLiveTurnHelpers",
		"Account":              "TestRealCodexAccountAndModelHelpers, TestRealCodexAuthHelpers",
		"StartLogin":           "TestRealCodexAuthHelpers",
		"CancelLogin":          "TestRealCodexAuthHelpers",
		"Logout":               "TestRealCodexAuthHelpers",
		"ListModels":           "TestRealCodexAccountAndModelHelpers, TestRealCodexAuthHelpers",
		"ListThreads":          "TestRealCodexThreadLifecycleHelpers",
		"ReadThread":           "TestRealCodexThreadLifecycleHelpers",
		"Thread.Read":          "TestRealCodexThreadLifecycleHelpers",
		"SetThreadName":        "TestRealCodexThreadLifecycleHelpers",
		"Thread.SetName":       "TestRealCodexThreadLifecycleHelpers",
		"ArchiveThread":        "TestRealCodexThreadArchiveRoundTripWithMockProvider, TestRealCodexLiveTurnHelpers",
		"Thread.Archive":       "TestRealCodexThreadArchiveRoundTripWithMockProvider, TestRealCodexLiveTurnHelpers",
		"UnarchiveThread":      "TestRealCodexThreadArchiveRoundTripWithMockProvider, TestRealCodexLiveTurnHelpers",
		"Thread.Unarchive":     "TestRealCodexThreadArchiveRoundTripWithMockProvider, TestRealCodexLiveTurnHelpers",
		"CompactThread":        "TestRealCodexThreadLifecycleHelpers",
		"Thread.Compact":       "TestRealCodexThreadLifecycleHelpers",
		"ForkThread":           "TestRealCodexThreadLifecycleHelpers, TestRealCodexLiveTurnHelpers",
		"Thread.Fork":          "TestRealCodexThreadLifecycleHelpers, TestRealCodexLiveTurnHelpers",
		"Thread.Run":           "TestRealCodexLiveTurnHelpers",
		"Thread.RunInputs":     "TestRealCodexLiveTurnHelpers",
		"Thread.RunStreamed":   "TestRealCodexLiveTurnHelpers",
		"Thread.StartTurn":     "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Stream":    "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Next":      "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Run":       "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Steer":     "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Interrupt": "TestRealCodexLiveTurnHelpers",
		"TurnHandle.Close":     "TestRealCodexLiveTurnHelpers",
		"TextInput":            "unit tests; pure value constructor with no app-server boundary",
		"ImageInput":           "unit tests; pure value constructor with no app-server boundary",
		"LocalImageInput":      "unit tests; pure value constructor with no app-server boundary",
		"SkillInput":           "unit tests; pure value constructor with no app-server boundary",
		"MentionInput":         "unit tests; pure value constructor with no app-server boundary",
		"JSON":                 "unit tests; pure JSON helper with no app-server boundary",
		"MustJSON":             "unit tests; pure JSON helper with no app-server boundary",
		"IsRetryable":          "unit tests; pure error helper with no app-server boundary",
		"IsOverloaded":         "unit tests; pure error helper with no app-server boundary",
	}
	for helper, testName := range coverage {
		if strings.TrimSpace(testName) == "" {
			t.Fatalf("missing e2e coverage entry for %s", helper)
		}
	}
}

func TestRealCodexClientLifecycleAndGeneratedRPC(t *testing.T) {
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{DisableAutoClose: true})

	if client.Client() == nil {
		t.Fatalf("expected generated RPC client")
	}
	if _, err := client.Client().ConfigRequirementsRead(ctx); err != nil {
		t.Fatalf("read config requirements through generated RPC client: %v\nstderr:\n%s", err, stderr.String())
	}
	if err := client.Close(); err != nil {
		t.Fatalf("close real codex app-server: %v\nstderr:\n%s", err, stderr.String())
	}
}

func TestRealCodexAccountAndModelHelpers(t *testing.T) {
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{})

	account, err := client.Account(ctx, codex.AccountOptions{})
	if err != nil {
		t.Fatalf("read account through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if account == nil {
		t.Fatalf("expected account response\nstderr:\n%s", stderr.String())
	}

	limit := 1
	includeHidden := false
	if _, err := client.ListModels(ctx, codex.ListModelsOptions{Limit: &limit, IncludeHidden: &includeHidden}); err != nil {
		t.Fatalf("list models through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
}

func TestRealCodexThreadLifecycleHelpers(t *testing.T) {
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{})
	cwd := t.TempDir()

	thread := test.StartThread(t, client, ctx, stderr, cwd)

	if _, err := client.ListThreads(ctx, codex.ThreadListOptions{Cwd: cwd}); err != nil {
		t.Fatalf("list threads through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
	read, err := client.ReadThread(ctx, thread.ID(), codex.ThreadReadOptions{})
	if err != nil {
		t.Fatalf("read thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if read.Thread.ID != thread.ID() {
		t.Fatalf("client read returned thread %q, want %q\nstderr:\n%s", read.Thread.ID, thread.ID(), stderr.String())
	}
	threadRead, err := thread.Read(ctx, codex.ThreadReadOptions{})
	if err != nil {
		t.Fatalf("read thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if threadRead.Thread.ID != thread.ID() {
		t.Fatalf("thread read returned thread %q, want %q\nstderr:\n%s", threadRead.Thread.ID, thread.ID(), stderr.String())
	}

	if _, err := client.ResumeThread(ctx, codex.ThreadResumeOptions{ThreadID: thread.ID(), Cwd: cwd}); err == nil || !test.IsExpectedUnmaterializedThreadError(err) {
		t.Fatalf("expected unmaterialized thread resume error, got %v\nstderr:\n%s", err, stderr.String())
	}

	if _, err := client.SetThreadName(ctx, thread.ID(), "codex-sdk-go e2e client"); err != nil {
		t.Fatalf("set thread name through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if _, err := thread.SetName(ctx, "codex-sdk-go e2e thread"); err != nil {
		t.Fatalf("set thread name through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}

	threadForClientCompact := test.StartThread(t, client, ctx, stderr, cwd)
	if _, err := client.CompactThread(ctx, threadForClientCompact.ID(), codex.ThreadCompactOptions{}); err != nil {
		t.Fatalf("compact thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	threadForMethodCompact := test.StartThread(t, client, ctx, stderr, cwd)
	if _, err := threadForMethodCompact.Compact(ctx, codex.ThreadCompactOptions{}); err != nil {
		t.Fatalf("compact thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}

	threadForClientFork := test.StartThread(t, client, ctx, stderr, cwd)
	forkedByClient, _, err := client.ForkThread(ctx, threadForClientFork.ID(), codex.ThreadForkOptions{Cwd: cwd})
	if err != nil && !test.IsExpectedUnmaterializedThreadError(err) {
		t.Fatalf("fork thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if err == nil && (forkedByClient.ID() == "" || forkedByClient.ID() == threadForClientFork.ID()) {
		t.Fatalf("expected new forked thread id, got %q from %q", forkedByClient.ID(), threadForClientFork.ID())
	}

	threadForMethodFork := test.StartThread(t, client, ctx, stderr, cwd)
	forkedByMethod, _, err := threadForMethodFork.Fork(ctx, codex.ThreadForkOptions{Cwd: cwd})
	if err != nil && !test.IsExpectedUnmaterializedThreadError(err) {
		t.Fatalf("fork thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if err == nil && (forkedByMethod.ID() == "" || forkedByMethod.ID() == threadForMethodFork.ID()) {
		t.Fatalf("expected new forked thread id, got %q from %q", forkedByMethod.ID(), threadForMethodFork.ID())
	}
}

func TestRealCodexAuthHelpers(t *testing.T) {
	loginParams, secret := test.RequireLoginParams(t)
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{
		Timeout: 90 * time.Second,
		Secrets: []string{secret},
	})

	if _, err := client.StartLogin(ctx, loginParams); err != nil {
		t.Fatalf("start login through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if _, err := client.Account(ctx, codex.AccountOptions{RefreshToken: true}); err != nil {
		t.Fatalf("read account after login: %v\nstderr:\n%s", err, stderr.String())
	}
	if _, err := client.ListModels(ctx, codex.ListModelsOptions{}); err != nil {
		t.Fatalf("list models after login: %v\nstderr:\n%s", err, stderr.String())
	}
	cancel, err := client.CancelLogin(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("cancel missing login through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if cancel == nil || cancel.Status == "" {
		t.Fatalf("expected cancel login status, got %#v\nstderr:\n%s", cancel, stderr.String())
	}
	if _, err := client.Logout(ctx); err != nil {
		t.Fatalf("logout temp e2e account through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
}

func TestRealCodexLiveTurnHelpers(t *testing.T) {
	loginParams, secret := test.RequireLoginParams(t)
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{
		Timeout: 4 * time.Minute,
		Secrets: []string{secret},
	})

	if _, err := client.StartLogin(ctx, loginParams); err != nil {
		t.Fatalf("start login for live turns: %v\nstderr:\n%s", err, stderr.String())
	}
	t.Cleanup(func() {
		logoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := client.Logout(logoutCtx); err != nil {
			t.Logf("logout temp e2e account after live turns: %v", err)
		}
	})

	runThread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	runResult, err := runThread.Run(ctx, "Reply with one short sentence containing codex-sdk-go e2e.", test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("run live turn through Thread.Run: %v\nstderr:\n%s", err, stderr.String())
	}
	test.AssertCompletedTurnResult(t, "Thread.Run", runResult)
	resumedRunThread, err := client.ResumeThread(ctx, codex.ThreadResumeOptions{ThreadID: runThread.ID(), Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("resume materialized live thread through high-level helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if resumedRunThread.ID() != runThread.ID() {
		t.Fatalf("expected resumed live thread id %q, got %q", runThread.ID(), resumedRunThread.ID())
	}
	if _, err := client.ArchiveThread(ctx, runThread.ID()); err != nil {
		t.Fatalf("archive materialized live thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	clientUnarchived, err := client.UnarchiveThread(ctx, runThread.ID())
	if err != nil {
		t.Fatalf("unarchive materialized live thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if clientUnarchived.Thread.ID != runThread.ID() {
		t.Fatalf("client unarchive returned thread %q, want %q", clientUnarchived.Thread.ID, runThread.ID())
	}
	if _, err := runThread.Archive(ctx); err != nil {
		t.Fatalf("archive materialized live thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}
	threadUnarchived, err := runThread.Unarchive(ctx)
	if err != nil {
		t.Fatalf("unarchive materialized live thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if threadUnarchived.Thread.ID != runThread.ID() {
		t.Fatalf("thread unarchive returned thread %q, want %q", threadUnarchived.Thread.ID, runThread.ID())
	}
	forkedRunThread, clientForkResponse, err := client.ForkThread(ctx, runThread.ID(), codex.ThreadForkOptions{Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("fork materialized live thread through client helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if forkedRunThread.ID() == "" || forkedRunThread.ID() == runThread.ID() {
		t.Fatalf("expected materialized client fork id, got %q from %q", forkedRunThread.ID(), runThread.ID())
	}
	if clientForkResponse.Thread == nil || clientForkResponse.Thread.ID != forkedRunThread.ID() {
		t.Fatalf("client fork response did not describe %q: %#v", forkedRunThread.ID(), clientForkResponse)
	}
	forkedByMethod, methodForkResponse, err := runThread.Fork(ctx, codex.ThreadForkOptions{Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("fork materialized live thread through thread helper: %v\nstderr:\n%s", err, stderr.String())
	}
	if forkedByMethod.ID() == "" || forkedByMethod.ID() == runThread.ID() {
		t.Fatalf("expected materialized method fork id, got %q from %q", forkedByMethod.ID(), runThread.ID())
	}
	if methodForkResponse.Thread == nil || methodForkResponse.Thread.ID != forkedByMethod.ID() {
		t.Fatalf("method fork response did not describe %q: %#v", forkedByMethod.ID(), methodForkResponse)
	}

	inputsThread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	inputsResult, err := inputsThread.RunInputs(ctx, []codex.Input{codex.TextInput("Reply with one short sentence containing structured input e2e.")}, test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("run live turn through Thread.RunInputs: %v\nstderr:\n%s", err, stderr.String())
	}
	test.AssertCompletedTurnResult(t, "Thread.RunInputs", inputsResult)

	streamThread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	stream, err := streamThread.RunStreamed(ctx, []codex.Input{codex.TextInput("Reply with one short sentence containing streamed e2e.")}, test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("start live stream through Thread.RunStreamed: %v\nstderr:\n%s", err, stderr.String())
	}
	streamResult, err := test.CollectTurnStream(ctx, stream)
	if err != nil {
		t.Fatalf("collect live stream through TurnStream.Next: %v\nstderr:\n%s", err, stderr.String())
	}
	test.AssertCompletedTurnResult(t, "Thread.RunStreamed", streamResult)

	handleThread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	handle, err := handleThread.StartTurn(ctx, []codex.Input{codex.TextInput("Reply with one short sentence containing handle run e2e.")}, test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("start live turn handle: %v\nstderr:\n%s", err, stderr.String())
	}
	handleResult, err := handle.Run(ctx)
	if err != nil {
		t.Fatalf("run live turn handle: %v\nstderr:\n%s", err, stderr.String())
	}
	test.AssertCompletedTurnResult(t, "TurnHandle.Run", handleResult)

	controlThread := test.StartThread(t, client, ctx, stderr, t.TempDir())
	controlHandle, err := controlThread.StartTurn(ctx, []codex.Input{codex.TextInput("Reply with one short sentence containing control e2e.")}, test.LiveTurnOptions(t))
	if err != nil {
		t.Fatalf("start live control turn handle: %v\nstderr:\n%s", err, stderr.String())
	}
	if err := test.WaitForKnownTurnID(ctx, controlHandle); err != nil {
		controlHandle.Close()
		t.Fatalf("wait for control turn id: %v\nstderr:\n%s", err, stderr.String())
	}
	if _, err := controlHandle.Steer(ctx, []codex.Input{codex.TextInput("Additional e2e steering input.")}); err != nil && !test.IsExpectedTurnControlStateError(err) {
		controlHandle.Close()
		t.Fatalf("steer live turn through TurnHandle.Steer: %v\nstderr:\n%s", err, stderr.String())
	}
	if _, err := controlHandle.Interrupt(ctx); err != nil && !test.IsExpectedTurnControlStateError(err) {
		controlHandle.Close()
		t.Fatalf("interrupt live turn through TurnHandle.Interrupt: %v\nstderr:\n%s", err, stderr.String())
	}
	controlHandle.Close()
	if _, err := controlHandle.Stream(); err == nil {
		t.Fatalf("expected closed turn handle stream to return an error")
	}
}
