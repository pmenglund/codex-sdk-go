package rpc

import (
	"context"
	"net"
	"runtime"
	"strings"
	"testing"
)

func TestConnTransportReadWrite(t *testing.T) {
	conn1, conn2 := net.Pipe()
	defer conn1.Close()
	defer conn2.Close()

	transport := NewConnTransport(conn1)

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		buf := make([]byte, 64)
		n, _ := conn2.Read(buf)
		if strings.TrimSpace(string(buf[:n])) != "hello" {
			t.Errorf("unexpected conn2 read: %q", string(buf[:n]))
		}
		_, _ = conn2.Write([]byte("world\n"))
	}()

	if err := transport.WriteLine("hello"); err != nil {
		t.Fatalf("WriteLine error: %v", err)
	}
	if line, err := transport.ReadLine(); err != nil || line != "world" {
		t.Fatalf("ReadLine error: %v line=%q", err, line)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	<-writeDone
}

func TestSpawnStdioEmptyBinary(t *testing.T) {
	if _, err := SpawnStdio(context.Background(), "", nil, nil); err == nil {
		t.Fatalf("expected error for empty binary")
	}
}

func TestDefaultStderr(t *testing.T) {
	if DefaultStderr() == nil {
		t.Fatalf("expected default stderr")
	}
}

func TestStdioTransportEcho(t *testing.T) {
	ctx := context.Background()
	transport, err := SpawnStdio(ctx, "/bin/cat", nil, nil)
	if err != nil {
		t.Fatalf("SpawnStdio error: %v", err)
	}

	if err := transport.WriteLine("ping"); err != nil {
		t.Fatalf("WriteLine error: %v", err)
	}
	line, err := transport.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine error: %v", err)
	}
	if line != "ping" {
		t.Fatalf("unexpected line: %s", line)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestStdioTransportCloseReportsWaitError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell exit test is unix-only")
	}

	transport, err := SpawnStdio(context.Background(), "/bin/sh", []string{"-c", "exit 7"}, nil)
	if err != nil {
		t.Fatalf("SpawnStdio error: %v", err)
	}

	err = transport.Close()
	if err == nil {
		t.Fatalf("expected close error from process exit")
	}
	if !strings.Contains(err.Error(), "wait for process") {
		t.Fatalf("expected wait error, got %v", err)
	}
}
