package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicAPIDeclarationsHaveGoDoc(t *testing.T) {
	for _, dir := range []string{".", "rpc", "protocol"} {
		path := filepath.Join("..", "..", dir)
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			t.Run(filepath.Join(dir, entry.Name()), func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join(path, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if err := validateExportedDeclarationComments(data); err != nil {
					t.Error(err)
				}
			})
		}
	}
}

func TestValidateInterfaceMethodGoDoc(t *testing.T) {
	for _, tc := range []struct {
		name    string
		comment string
		wantErr bool
	}{
		{name: "missing", wantErr: true},
		{name: "wrong prefix", comment: "// Other calls the server.", wantErr: true},
		{name: "documented", comment: "// Call calls the server.\n//\n// Deprecated: use a narrower interface."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte("package rpc\n// Public is a client interface.\ntype Public interface {\nEmbedded\n" + tc.comment + "\nCall() error\nprivate()\n}\n")
			err := validateExportedDeclarationComments(source)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "Public.Call") {
					t.Fatalf("expected undocumented Public.Call, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestValidateGroupedValueGoDoc(t *testing.T) {
	for _, tc := range []struct {
		name    string
		second  string
		wantErr bool
	}{
		{name: "missing member comment", wantErr: true},
		{name: "documented members", second: "// Second is the second choice."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte("package rpc\n// Choices describe available choices.\nconst (\n// First is the first choice.\nFirst = 1\n" + tc.second + "\nSecond = 2\n)\n")
			err := validateExportedDeclarationComments(source)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "Second") {
					t.Fatalf("expected undocumented Second, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRPCMethodGoDoc(t *testing.T) {
	methods := []rpcMethod{
		{Method: "ping", ResponseType: "PingResponse"},
		{Method: "thread/start", ParamsType: "ThreadStartParams", ResponseType: "ThreadStartResponse"},
		{Method: "account/usage/read", ParamsType: "*GetAccountTokenUsageParams", ResponseType: "GetAccountTokenUsageResponse"},
		{Method: "mcpServer/elicitation/request", ParamsType: "MCPServerElicitationRequestParams", ResponseType: "MCPServerElicitationRequestResponse"},
	}
	for _, tc := range []struct {
		name      string
		source    []byte
		want      []string
		wantTwice []string
	}{
		{
			name:   "client",
			source: renderClientRequests(methods, testCodexCommit),
			want: []string{
				"// AccountUsageRead calls account/usage/read without optional filters.",
				"// AccountUsageRead reads without optional filters. Use AccountUsageReadWithParams to supply parameters.",
				"// McpServerElicitationRequest is the former spelling of MCPServerElicitationRequest.",
				"// Deprecated: use MCPServerElicitationRequest.",
			},
			wantTwice: []string{
				"// Ping calls ping on the app-server.",
				"// ThreadStart calls thread/start on the app-server.",
				"// AccountUsageReadWithParams calls account/usage/read on the app-server.",
				"// MCPServerElicitationRequest calls mcpServer/elicitation/request on the app-server.",
			},
		},
		{
			name:   "server",
			source: renderServerRequests(methods, testCodexCommit),
			want: []string{
				"// Ping handles the ping request from the app-server.",
				"// ThreadStart handles the thread/start request from the app-server.",
				"// MCPServerElicitationRequest handles the mcpServer/elicitation/request request from the app-server.",
				"// McpServerElicitationRequest handles the mcpServer/elicitation/request request from the app-server.",
				"// Ping returns ErrServerRequestUnsupported for ping.",
				"// McpServerElicitationRequest returns ErrServerRequestUnsupported for mcpServer/elicitation/request.",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateExportedDeclarationComments(tc.source); err != nil {
				t.Error(err)
			}
			for _, want := range tc.want {
				if !strings.Contains(string(tc.source), want) {
					t.Errorf("missing comment %q", want)
				}
			}
			for _, want := range tc.wantTwice {
				if got := strings.Count(string(tc.source), want); got != 2 {
					t.Errorf("expected interface and concrete method comments %q; got %d copies", want, got)
				}
			}
		})
	}
}

func TestWriteGoFileRejectsMissingGoDoc(t *testing.T) {
	dir := t.TempDir()
	err := writeGoFile(dir, "undocumented.go", []byte("package rpc\nfunc Undocumented() {}\n"))
	if err == nil || !strings.Contains(err.Error(), "Undocumented") {
		t.Fatalf("expected missing GoDoc error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "undocumented.go")); !os.IsNotExist(err) {
		t.Fatalf("undocumented file must not be written, got %v", err)
	}
}
