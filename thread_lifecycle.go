package codex

import (
	"context"
	"errors"
	"fmt"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

// ThreadListOptions configures a thread/list request.
type ThreadListOptions struct {
	Archived *bool
	Cursor   string
	Cwd      any
	// IsPinned is retained for source compatibility with Codex versions before
	// 0.147. Codex 0.147 replaces pinned-thread organization with sections.
	//
	// Deprecated: use SectionID or Unsectioned.
	IsPinned       *bool
	Limit          *int
	ModelProviders []string
	SearchTerm     string
	// SectionID limits results to one persisted section. It cannot be combined
	// with Unsectioned.
	SectionID     string
	SortDirection protocol.SortDirection
	SortKey       protocol.ThreadSortKey
	SourceKinds   []protocol.ThreadSourceKind
	// Unsectioned limits results to threads that do not belong to a section. It
	// cannot be combined with SectionID.
	Unsectioned    bool
	UseStateDBOnly *bool
}

func (o ThreadListOptions) toParams() (protocol.ThreadListParams, error) {
	if o.SortDirection != "" && o.SortDirection != protocol.SortDirectionAsc && o.SortDirection != protocol.SortDirectionDesc {
		return protocol.ThreadListParams{}, fmt.Errorf("invalid thread sort direction %q", o.SortDirection)
	}
	switch o.SortKey {
	case "", protocol.ThreadSortKeyCreatedAt, protocol.ThreadSortKeyUpdatedAt, protocol.ThreadSortKeyRecencyAt, protocol.ThreadSortKeySectionPosition:
	default:
		return protocol.ThreadListParams{}, fmt.Errorf("invalid thread sort key %q", o.SortKey)
	}
	if o.SectionID != "" && o.Unsectioned {
		return protocol.ThreadListParams{}, errors.New("thread section id and unsectioned filter cannot both be set")
	}
	params := protocol.ThreadListParams{
		Archived:       o.Archived,
		Cwd:            o.Cwd,
		Limit:          o.Limit,
		SortDirection:  o.SortDirection,
		SortKey:        o.SortKey,
		UseStateDbOnly: o.UseStateDBOnly,
	}
	if o.Unsectioned {
		var unsectioned *string
		params.SectionID = &unsectioned
	} else if o.SectionID != "" {
		sectionID := o.SectionID
		sectionValue := &sectionID
		params.SectionID = &sectionValue
	}
	if o.ModelProviders != nil {
		modelProviders := protocol.ThreadListParamsModelProviders(o.ModelProviders)
		params.ModelProviders = &modelProviders
	}
	if o.Cursor != "" {
		params.Cursor = stringPtr(o.Cursor)
	}
	if o.SearchTerm != "" {
		params.SearchTerm = stringPtr(o.SearchTerm)
	}
	if o.SourceKinds != nil {
		sourceKinds := protocol.ThreadListParamsSourceKinds(o.SourceKinds)
		params.SourceKinds = &sourceKinds
	}
	return params, nil
}

// ListThreads returns persisted threads visible to the app-server.
func (c *Codex) ListThreads(ctx context.Context, opts ThreadListOptions) (*protocol.ThreadListResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	params, err := opts.toParams()
	if err != nil {
		return nil, err
	}
	return c.client.ThreadList(ctx, params)
}

// ThreadReadOptions configures a thread/read request.
type ThreadReadOptions struct {
	IncludeTurns bool
}

// ReadThread reads a persisted thread by id.
func (c *Codex) ReadThread(ctx context.Context, threadID string, opts ThreadReadOptions) (*protocol.ThreadReadResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if threadID == "" {
		return nil, errors.New("thread id is required")
	}
	params := protocol.ThreadReadParams{ThreadID: threadID}
	if opts.IncludeTurns {
		params.IncludeTurns = boolPtr(opts.IncludeTurns)
	}
	return c.client.ThreadRead(ctx, params)
}

// Read reads this thread from the app-server.
func (t *Thread) Read(ctx context.Context, opts ThreadReadOptions) (*protocol.ThreadReadResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}
	params := protocol.ThreadReadParams{ThreadID: t.id}
	if opts.IncludeTurns {
		params.IncludeTurns = boolPtr(opts.IncludeTurns)
	}
	return t.client.ThreadRead(ctx, params)
}

// SetThreadName sets the display name for a thread by id.
func (c *Codex) SetThreadName(ctx context.Context, threadID, name string) (*protocol.ThreadSetNameResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if threadID == "" {
		return nil, errors.New("thread id is required")
	}
	if name == "" {
		return nil, errors.New("thread name is required")
	}
	return c.client.ThreadNameSet(ctx, protocol.ThreadSetNameParams{ThreadID: threadID, Name: name})
}

// SetName sets the display name for this thread.
func (t *Thread) SetName(ctx context.Context, name string) (*protocol.ThreadSetNameResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("thread name is required")
	}
	return t.client.ThreadNameSet(ctx, protocol.ThreadSetNameParams{ThreadID: t.id, Name: name})
}

// ArchiveThread archives a thread by id.
func (c *Codex) ArchiveThread(ctx context.Context, threadID string) (*protocol.ThreadArchiveResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if threadID == "" {
		return nil, errors.New("thread id is required")
	}
	return c.client.ThreadArchive(ctx, protocol.ThreadArchiveParams{ThreadID: threadID})
}

// Archive archives this thread.
func (t *Thread) Archive(ctx context.Context) (*protocol.ThreadArchiveResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}
	return t.client.ThreadArchive(ctx, protocol.ThreadArchiveParams{ThreadID: t.id})
}

// UnarchiveThread unarchives a thread by id.
func (c *Codex) UnarchiveThread(ctx context.Context, threadID string) (*protocol.ThreadUnarchiveResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if threadID == "" {
		return nil, errors.New("thread id is required")
	}
	return c.client.ThreadUnarchive(ctx, protocol.ThreadUnarchiveParams{ThreadID: threadID})
}

// Unarchive unarchives this thread.
func (t *Thread) Unarchive(ctx context.Context) (*protocol.ThreadUnarchiveResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}
	return t.client.ThreadUnarchive(ctx, protocol.ThreadUnarchiveParams{ThreadID: t.id})
}

// ThreadCompactOptions configures a thread/compact/start request.
type ThreadCompactOptions struct{}

// CompactThread starts compaction for a thread by id.
func (c *Codex) CompactThread(ctx context.Context, threadID string, opts ThreadCompactOptions) (*protocol.ThreadCompactStartResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if threadID == "" {
		return nil, errors.New("thread id is required")
	}
	return c.client.ThreadCompactStart(ctx, protocol.ThreadCompactStartParams{ThreadID: threadID})
}

// Compact starts compaction for this thread.
func (t *Thread) Compact(ctx context.Context, opts ThreadCompactOptions) (*protocol.ThreadCompactStartResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}
	return t.client.ThreadCompactStart(ctx, protocol.ThreadCompactStartParams{ThreadID: t.id})
}

// ThreadForkOptions configures a thread/fork request.
type ThreadForkOptions struct {
	Model                 string
	ModelProvider         string
	ServiceTier           string
	LastTurnID            string
	Cwd                   string
	ApprovalPolicy        any
	Sandbox               any
	Config                map[string]any
	BaseInstructions      string
	DeveloperInstructions string
	Ephemeral             *bool
	// ExcludeTurns returns metadata without hydrating turn history when true.
	ExcludeTurns *bool
}

func (o ThreadForkOptions) toParams(threadID string) (protocol.ThreadForkParams, error) {
	params := protocol.ThreadForkParams{
		ThreadID:     threadID,
		Ephemeral:    o.Ephemeral,
		ExcludeTurns: o.ExcludeTurns,
	}
	if o.Model != "" {
		params.Model = stringPtr(o.Model)
	}
	if o.ModelProvider != "" {
		params.ModelProvider = stringPtr(o.ModelProvider)
	}
	if o.ServiceTier != "" {
		params.ServiceTier = stringPtr(o.ServiceTier)
	}
	if o.LastTurnID != "" {
		params.LastTurnID = stringPtr(o.LastTurnID)
	}
	if o.Cwd != "" {
		params.Cwd = stringPtr(o.Cwd)
	}
	if raw, err := normalizeJSONValue("approvalPolicy", o.ApprovalPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.ApprovalPolicy = raw
	}
	if raw, err := normalizeJSONValue("sandbox", o.Sandbox); err != nil {
		return params, err
	} else if raw != nil {
		params.Sandbox = raw
	}
	if o.Config != nil {
		config := o.Config
		params.Config = &config
	}
	if o.BaseInstructions != "" {
		params.BaseInstructions = stringPtr(o.BaseInstructions)
	}
	if o.DeveloperInstructions != "" {
		params.DeveloperInstructions = stringPtr(o.DeveloperInstructions)
	}
	return params, nil
}

// ForkThread forks a thread by id and returns the newly forked thread.
func (c *Codex) ForkThread(ctx context.Context, threadID string, opts ThreadForkOptions) (*Thread, protocol.ThreadForkResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	if threadID == "" {
		return nil, protocol.ThreadForkResponse{}, errors.New("thread id is required")
	}
	params, err := opts.toParams(threadID)
	if err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	response, err := c.client.ThreadFork(ctx, params)
	if err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	id, err := threadIDFromResponse(response.ThreadID, response.Thread)
	if err != nil {
		return nil, *response, err
	}
	return &Thread{client: c.client, id: id, logger: c.logger}, *response, nil
}

// Fork forks this thread and returns the newly forked thread.
func (t *Thread) Fork(ctx context.Context, opts ThreadForkOptions) (*Thread, protocol.ThreadForkResponse, error) {
	if err := t.ensureReady(); err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	params, err := opts.toParams(t.id)
	if err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	response, err := t.client.ThreadFork(ctx, params)
	if err != nil {
		return nil, protocol.ThreadForkResponse{}, err
	}
	id, err := threadIDFromResponse(response.ThreadID, response.Thread)
	if err != nil {
		return nil, *response, err
	}
	return &Thread{client: t.client, id: id, logger: t.logger}, *response, nil
}
