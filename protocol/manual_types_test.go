package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestManualThreadResumeResponseRoundTripsCompletePayload(t *testing.T) {
	payload := []byte(`{
        "threadId":"legacy-thread-id",
        "thread":{
            "id":"thr_1","extra":{"future":{"value":1}},"sessionId":"session_1",
            "forkedFromId":"thr_0","parentThreadId":"thr_parent","preview":"preview",
            "ephemeral":true,"historyMode":"paginated","modelProvider":"openai",
            "createdAt":10,"updatedAt":20,"recencyAt":19,
            "status":{"type":"idle"},"path":"/tmp/thread.jsonl","cwd":"/tmp/project",
            "cliVersion":"0.144.6","source":{"type":"cli"},"threadSource":"cli",
            "agentNickname":"Ada","agentRole":"reviewer",
            "gitInfo":{"sha":"abc","branch":"main","originUrl":"https://example.test/repo"},
            "name":"Full payload","turns":[{
                "id":"turn_1","items":[{"type":"agentMessage","id":"item_1","text":"done"}],
                "itemsView":"full","status":"failed",
                "error":{"message":"boom","additionalDetails":"details","codexErrorInfo":{"code":"future"}},
                "startedAt":100,"completedAt":101,"durationMs":1000
            }]
        },
        "model":"gpt-test","modelProvider":"openai","serviceTier":"priority","cwd":"/tmp/project",
        "runtimeWorkspaceRoots":["/tmp/project"],"instructionSources":["/tmp/AGENTS.md"],
        "approvalPolicy":"never","approvalsReviewer":{"type":"user"},
        "sandbox":{"type":"readOnly"},"activePermissionProfile":{"name":"readonly"},
        "reasoningEffort":"high","multiAgentMode":"explicitRequestOnly",
        "initialTurnsPage":{"data":[{"id":"turn_2","items":[],"itemsView":"summary","status":"completed"}],"nextCursor":"next","backwardsCursor":"back"}
    }`)

	var response ThreadResumeResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("decode complete response: %v", err)
	}
	if response.Thread == nil || len(response.Thread.Turns) != 1 || response.InitialTurnsPage == nil || len(response.InitialTurnsPage.Data) != 1 {
		t.Fatalf("complete typed fields were not retained: %#v", response)
	}
	roundTrip, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("encode complete response: %v", err)
	}
	var before, after any
	if err := json.Unmarshal(payload, &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(roundTrip, &after); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("manual response lost wire fields\nbefore: %s\nafter:  %s", payload, roundTrip)
	}
}

func TestApprovalDecisionDecodingPreservesUnknownVariants(t *testing.T) {
	var review ReviewDecision
	if err := json.Unmarshal([]byte(`"future_review"`), &review); err != nil {
		t.Fatalf("decode future review decision: %v", err)
	}
	if review.IsKnown() || review.Kind() != "future_review" {
		t.Fatalf("unexpected future review decision: %#v", review)
	}
	if _, err := NewReviewDecision("future_review"); err == nil {
		t.Fatal("checked review constructor accepted a future variant")
	}

	var command CommandExecutionApprovalDecision
	if err := json.Unmarshal([]byte(`{"futureDecision":{"value":1}}`), &command); err != nil {
		t.Fatalf("decode future command decision: %v", err)
	}
	if command.IsKnown() || command.Kind() != "futureDecision" {
		t.Fatalf("unexpected future command decision: %#v", command)
	}
	if _, err := NewCommandExecutionApprovalDecision(map[string]any{"acceptWithExecpolicyAmendment": map[string]any{}}); err == nil {
		t.Fatal("checked command constructor accepted a malformed known variant")
	}
}
