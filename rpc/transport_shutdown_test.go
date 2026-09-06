package rpc

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStdioTransportForcedShutdown(t *testing.T) {
	// This child deliberately remains alive after stdin EOF. Use real process I/O,
	// not synctest: readiness and EOF acknowledgments establish the ordering.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tr, err := SpawnStdio(ctx, binary, []string{"-test.run=^TestStdioShutdownHelperProcess$", "--", "codex-shutdown-helper"}, os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			_ = tr.cmd.Process.Kill()
			_ = tr.cmd.Wait()
		}
	}()
	if line, err := tr.ReadLine(); err != nil || line != "ready" {
		t.Fatalf("child readiness = %q, %v", line, err)
	}
	done := make(chan error, 1)
	go func() { done <- tr.Close() }()
	// From here Close owns Wait. Kill and join it if any assertion fails.
	defer func() {
		if !waited {
			_ = tr.cmd.Process.Kill()
			<-done
			waited = true
		}
	}()
	if line, err := tr.ReadLine(); err != nil || line != "stdin closed" {
		t.Fatalf("EOF acknowledgment = %q, %v", line, err)
	}
	err = <-done
	waited = true
	if ctx.Err() != nil {
		t.Fatalf("watchdog expired: %v", ctx.Err())
	}
	if err == nil || !strings.Contains(err.Error(), "wait after kill") {
		t.Fatalf("expected forced-shutdown error, got %v", err)
	}
	if tr.cmd.ProcessState == nil || tr.cmd.ProcessState.Success() {
		t.Fatalf("child was not reaped after forced termination: %v", tr.cmd.ProcessState)
	}
	// A second Wait must report that the first Wait already reaped the child.
	if err := tr.cmd.Wait(); err == nil {
		t.Fatal("second Wait unexpectedly succeeded")
	}
}

func TestStdioShutdownHelperProcess(t *testing.T) {
	args := os.Args
	if len(args) < 2 || args[len(args)-2] != "--" || args[len(args)-1] != "codex-shutdown-helper" {
		return
	}
	if _, err := fmt.Fprintln(os.Stdout, "ready"); err != nil {
		os.Exit(2)
	}
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		os.Exit(3)
	}
	if _, err := fmt.Fprintln(os.Stdout, "stdin closed"); err != nil {
		os.Exit(4)
	}
	// A timer keeps the helper alive without a runtime deadlock panic.
	time.Sleep(time.Hour)
	os.Exit(5)
}
