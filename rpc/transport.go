package rpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"
)

const stdioCloseTimeout = 2 * time.Second

// Transport reads and writes JSON-RPC lines. Implementations should not block
// WriteLine indefinitely, and Close must unblock in-flight reads and writes so
// bounded SDK cleanup cannot leave transport goroutines running.
type Transport interface {
	ReadLine() (string, error)
	WriteLine(line string) error
	Close() error
}

// ContextTransport is an optional Transport capability for writes that can be
// interrupted when their request context ends. Client also bounds callers that
// use a legacy Transport, but the underlying legacy write can remain blocked
// until Transport.Close unblocks it.
type ContextTransport interface {
	Transport
	WriteLineContext(ctx context.Context, line string) error
}

// StdioTransport wraps a spawned process using stdin/stdout JSONL.
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

// SpawnStdio starts a command and uses its stdin/stdout for JSON-RPC.
func SpawnStdio(ctx context.Context, binary string, args []string, stderr io.Writer) (*StdioTransport, error) {
	if binary == "" {
		return nil, errors.New("codex binary path is empty")
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &StdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

// ReadLine reads a single line from stdout.
func (t *StdioTransport) ReadLine() (string, error) {
	line, err := t.stdout.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return strings.TrimRight(line, "\n"), nil
		}
		return "", err
	}
	return strings.TrimRight(line, "\n"), nil
}

// WriteLine writes a single line to stdin.
func (t *StdioTransport) WriteLine(line string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	_, err := io.WriteString(t.stdin, line)
	return err
}

// Close shuts down the process.
func (t *StdioTransport) Close() error {
	var errs []error
	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stdin: %w", err))
		}
	}
	if t.cmd == nil {
		return errors.Join(errs...)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- t.cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		if err != nil {
			errs = append(errs, fmt.Errorf("wait for process: %w", err))
		}
	case <-time.After(stdioCloseTimeout):
		if t.cmd.Process != nil {
			if err := t.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				errs = append(errs, fmt.Errorf("kill process: %w", err))
			}
		}
		if err := <-waitCh; err != nil {
			errs = append(errs, fmt.Errorf("wait after kill: %w", err))
		}
	}

	return errors.Join(errs...)
}

// ConnTransport wraps an io.ReadWriteCloser.
type ConnTransport struct {
	conn   io.ReadWriteCloser
	reader *bufio.Reader
	mu     sync.Mutex
}

// NewConnTransport wraps the connection in a Transport. It panics if conn is nil.
// Use NewConnTransportChecked when the dependency is not statically known.
func NewConnTransport(conn io.ReadWriteCloser) *ConnTransport {
	transport, err := NewConnTransportChecked(conn)
	if err != nil {
		panic(err)
	}
	return transport
}

// NewConnTransportChecked wraps the connection in a Transport.
func NewConnTransportChecked(conn io.ReadWriteCloser) (*ConnTransport, error) {
	if isNilInterface(conn) {
		return nil, errors.New("rpc connection is nil")
	}
	return &ConnTransport{conn: conn, reader: bufio.NewReader(conn)}, nil
}

// ReadLine reads a line from the connection.
func (t *ConnTransport) ReadLine() (string, error) {
	line, err := t.reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return strings.TrimRight(line, "\n"), nil
		}
		return "", err
	}
	return strings.TrimRight(line, "\n"), nil
}

// WriteLine writes a line to the connection.
func (t *ConnTransport) WriteLine(line string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	_, err := io.WriteString(t.conn, line)
	return err
}

// Close closes the connection.
func (t *ConnTransport) Close() error {
	return t.conn.Close()
}

// DefaultStderr returns a safe default for spawned processes.
func DefaultStderr() io.Writer {
	return os.Stderr
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
