package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/WeKnora/internal/anonymousapi"
	"github.com/Tencent/WeKnora/internal/archivalcitations"
	"github.com/Tencent/WeKnora/internal/event"
	"github.com/Tencent/WeKnora/internal/logger"
	modelapi "github.com/Tencent/WeKnora/internal/models/api"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type anonymousResponse struct {
	ID            string
	TenantID      uint64
	UserID        string
	ProfileID     string
	ProfileSHA256 string
	SessionID     string
	MessageID     string
	CreatedAt     time.Time
}

func (anonymousResponse) TableName() string { return "ljc_anonymous_api_responses" }
func anonError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": gin.H{"message": msg, "type": "invalid_request_error", "param": nil, "code": http.StatusText(status)}})
}
func profileHash(p anonymousapi.Profile) string {
	b, _ := json.Marshal(p)
	v := sha256.Sum256(b)
	return hex.EncodeToString(v[:])
}

// RegisterAnonymousAPI must run before global authentication. No other routes bypass Auth.
func (h *Handler) RegisterAnonymousAPI(r *gin.Engine) {
	g := r.Group("/api/openai/v1")
	g.Use(func(c *gin.Context) {
		cfg, e := anonymousapi.Load(os.Getenv("WEKNORA_ANONYMOUS_API_CONFIG"))
		if e != nil {
			anonError(c, 503, "API configuration unavailable")
			c.Abort()
			return
		}
		if !cfg.Enabled {
			anonError(c, 503, "Anonymous API is disabled")
			c.Abort()
			return
		}
		ip, e := cfg.ClientIP(c.Request)
		if e != nil {
			anonError(c, 403, "Source network not allowed")
			c.Abort()
			return
		}
		c.Set("anonymous_config", cfg)
		c.Set("anonymous_ip", ip.String())
		c.Next()
	})
	g.GET("/models", func(c *gin.Context) {
		cfg := c.MustGet("anonymous_config").(*anonymousapi.Config)
		out := []gin.H{}
		for _, p := range cfg.Profiles {
			if _, _, e := h.anonymousIdentity(c.Request.Context(), cfg, p); e == nil {
				out = append(out, gin.H{"id": p.ID, "object": "model", "created": 0, "owned_by": "weknora"})
			}
		}
		c.JSON(200, gin.H{"object": "list", "data": out})
	})
	g.POST("/chat/completions", h.anonymousChat)
	g.GET("/responses/:response_id/citations", h.anonymousCitations)
	g.GET("/responses/:response_id/citations/:citation_id", h.anonymousCitations)
}
func (h *Handler) anonymousIdentity(ctx context.Context, cfg *anonymousapi.Config, p anonymousapi.Profile) (context.Context, *types.CustomAgent, error) {
	if h.citationStore == nil || h.userService == nil || h.memberService == nil {
		return ctx, nil, errors.New("services unavailable")
	}
	user, e := h.userService.GetUserByID(ctx, cfg.UserID)
	if e != nil || user == nil || !user.IsActive || user.IsSystemAdmin || user.CanAccessAllTenants {
		return ctx, nil, errors.New("execution identity unavailable")
	}
	member, e := h.memberService.GetMembership(ctx, cfg.UserID, cfg.TenantID)
	if e != nil || member == nil || member.Role != types.TenantRoleViewer || member.Status != types.TenantMemberStatusActive {
		return ctx, nil, errors.New("dedicated identity must be an active viewer")
	}
	tenant, e := h.tenantService.GetTenantByID(ctx, cfg.TenantID)
	if e != nil {
		return ctx, nil, e
	}
	ctx = context.WithValue(ctx, types.TenantIDContextKey, cfg.TenantID)
	ctx = context.WithValue(ctx, types.TenantInfoContextKey, tenant)
	ctx = context.WithValue(ctx, types.UserIDContextKey, cfg.UserID)
	ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
	ctx = types.WithPrincipal(ctx, types.Principal{Type: types.PrincipalWebUser, ID: cfg.UserID})
	ctx = types.WithTenantAPIKeyScope(ctx, types.TenantAPIKeyScope{Name: "anonymous-api", FullAccess: false, KnowledgeBaseIDs: types.StringArray(p.KnowledgeBaseIDs), Capabilities: types.StringArray{"chat", "retrieve"}})
	for _, id := range p.KnowledgeBaseIDs {
		kb, e := h.knowledgebaseService.GetKnowledgeBaseByID(ctx, id)
		if e != nil || kb == nil || kb.TenantID != cfg.TenantID {
			return ctx, nil, errors.New("knowledge scope unavailable")
		}
	}
	a, e := h.customAgentService.GetAgentByID(ctx, p.AgentID)
	if e != nil || a == nil {
		return ctx, nil, errors.New("agent unavailable")
	}
	if a.ID == types.BuiltinWeKnoraAssistantID || a.ID == types.BuiltinSkillInstallerID || a.ID == types.BuiltinWikiFixerID {
		return ctx, nil, errors.New("management agents are not allowed")
	}
	raw, e := json.Marshal(a)
	if e != nil {
		return ctx, nil, e
	}
	var copy types.CustomAgent
	if e = json.Unmarshal(raw, &copy); e != nil {
		return ctx, nil, e
	}
	copy.EnsureDefaults()
	copy.Config.KBSelectionMode = "selected"
	copy.Config.KnowledgeBases = append([]string{}, p.KnowledgeBaseIDs...)
	copy.Config.RetrieveKBOnlyWhenMentioned = false
	copy.Config.AccessLevel = types.AssistantAccessReadOnly
	copy.Config.AllowedTools = []string{"search_knowledge", "knowledge_search", "grep_chunks", "read_document", "list_knowledge_chunks", "search_wiki", "read_wiki_page"}
	copy.Config.MCPSelectionMode = "none"
	copy.Config.MCPServices = nil
	copy.Config.SkillsSelectionMode = "none"
	copy.Config.WebSearchEnabled = false
	copy.Config.SandboxConfigID = ""
	copy.Config.DataAnalysisEnabled = false
	no := false
	copy.Config.MemoryEnabled = &no
	copy.Config.WebFetchEnabled = false
	copy.Config.MaxIterations = 6
	copy.Config.LLMCallTimeout = 120
	copy.Config.MaxCompletionTokens = p.MaxTokens
	copy.Config.MultiTurnEnabled = true
	copy.Config.HistoryTurns = 50
	yes := true
	copy.Config.CitationEnabled = &yes
	return ctx, &copy, nil
}
func (h *Handler) anonymousChat(c *gin.Context) {
	cfg := c.MustGet("anonymous_config").(*anonymousapi.Config)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512*1024)
	req, e := anonymousapi.Decode(c.Request.Body)
	if e != nil {
		anonError(c, 400, "Invalid request: "+e.Error())
		return
	}
	p, ok := cfg.Profile(req.Model)
	if !ok {
		anonError(c, 404, "Unknown model alias")
		return
	}
	if req.MaxTokens != nil {
		if *req.MaxTokens > p.MaxTokens {
			anonError(c, 400, "max_tokens exceeds profile limit")
			return
		}
	}
	started := time.Now()
	requestID := uuid.NewString()
	id := "chatcmpl-" + requestID
	ip := c.MustGet("anonymous_ip").(string)
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()
	release, e := anonymousapi.Acquire(ctx, h.redis, cfg, ip, id)
	if e != nil {
		if errors.Is(e, anonymousapi.ErrLimited) {
			c.Header("Retry-After", "60")
			anonError(c, 429, "Request limit exceeded")
		} else {
			anonError(c, 503, "Shared limiter unavailable")
		}
		return
	}
	defer release()
	ctx = logger.WithPrivateContent(ctx)
	var usageMu sync.Mutex
	var observedUsage types.TokenUsage
	firstToken := time.Duration(0)
	defer func() {
		usageMu.Lock()
		defer usageMu.Unlock()
		errorType := "none"
		if e != nil {
			errorType = "execution_error"
		}
		if ctx.Err() != nil {
			errorType = "canceled_or_timeout"
		}
		logger.Infof(c.Request.Context(), "anonymous_api request=%s ip=%s profile=%s first_token_ms=%d duration_ms=%d prompt_tokens=%d completion_tokens=%d total_tokens=%d status=%d error_type=%s", id, ip, p.ID, firstToken.Milliseconds(), time.Since(started).Milliseconds(), observedUsage.PromptTokens, observedUsage.CompletionTokens, observedUsage.TotalTokens, c.Writer.Status(), errorType)
	}()
	ctx = modelapi.WithUsageObserver(ctx, func(u types.TokenUsage) { usageMu.Lock(); defer usageMu.Unlock(); observedUsage.Accumulate(u) })
	ctx, a, e := h.anonymousIdentity(ctx, cfg, p)
	if e != nil {
		anonError(c, 403, "Question-answering profile unavailable")
		return
	}
	if req.MaxTokens != nil {
		a.Config.MaxCompletionTokens = *req.MaxTokens
	}
	if !archivalcitations.Enabled() {
		anonError(c, 503, "Citation support unavailable")
		return
	}
	session, e := h.sessionService.CreateSession(ctx, &types.Session{ID: uuid.NewString(), TenantID: cfg.TenantID, UserID: cfg.UserID, Kind: "anonymous-api", Title: "Anonymous API"})
	if e != nil {
		anonError(c, 503, "Session creation failed")
		return
	}
	historyRequestID := requestID
	for i, m := range req.Messages {
		if m.Role == "user" {
			if i == len(req.Messages)-1 {
				historyRequestID = requestID
			} else {
				historyRequestID = uuid.NewString()
			}
		}
		_, e = h.messageService.CreateMessage(ctx, &types.Message{ID: uuid.NewString(), SessionID: session.ID, RequestID: historyRequestID, Content: m.Content, Role: m.Role, IsCompleted: true, Channel: "anonymous-api"})
		if e != nil {
			anonError(c, 503, "History persistence failed")
			return
		}
	}
	assistant := &types.Message{ID: uuid.NewString(), SessionID: session.ID, RequestID: requestID, Role: "assistant", Channel: "anonymous-api", ModelID: a.Config.ModelID}
	assistant, e = h.messageService.CreateMessage(ctx, assistant)
	if e != nil {
		anonError(c, 503, "Message creation failed")
		return
	}
	ctx = archivalcitations.WithScope(ctx, h.citationStore, cfg.TenantID, session.ID, assistant.ID)
	bus := event.NewEventBus()
	turnDone := make(chan struct{})
	var doneOnce sync.Once
	done := func() { doneOnce.Do(func() { close(turnDone) }) }
	var answer strings.Builder
	var usage *types.TokenUsage
	var runErr error
	var stateMu sync.Mutex
	closed := false
	truncated := false
	projector := newAnonymousProjector()
	streamStarted := false
	send := func(data any) error {
		if !streamStarted {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("X-Accel-Buffering", "no")
			c.Status(200)
			streamStarted = true
		}
		b, e := json.Marshal(data)
		if e != nil {
			return e
		}
		_, e = fmt.Fprintf(c.Writer, "data: %s\n\n", b)
		c.Writer.Flush()
		return e
	}
	chunk := func(delta gin.H, finish any, extra gin.H) error {
		v := gin.H{"id": id, "object": "chat.completion.chunk", "created": started.Unix(), "model": p.ID, "choices": []gin.H{{"index": 0, "delta": delta, "finish_reason": finish}}}
		for k, x := range extra {
			v[k] = x
		}
		return send(v)
	}
	bus.On(event.EventAgentFinalAnswer, func(_ context.Context, ev event.Event) error {
		stateMu.Lock()
		defer stateMu.Unlock()
		if closed {
			return ctx.Err()
		}
		d, ok := ev.Data.(event.AgentFinalAnswerData)
		if !ok {
			return nil
		}
		if d.Replace {
			answer.Reset()
			projector = newAnonymousProjector()
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.Done && a.Config.AgentMode != types.AgentModeSmartReasoning {
			defer done()
		}
		if d.Truncated {
			truncated = true
		}
		answer.WriteString(d.Content)
		if d.Content != "" && firstToken == 0 {
			firstToken = time.Since(started)
		}
		if req.Stream {
			out := projector.Feed(d.Content)
			if out != "" {
				if e := chunk(gin.H{"content": out}, nil, nil); e != nil {
					cancel()
					return e
				}
			}
		}
		return nil
	})
	bus.On(event.EventAgentComplete, func(_ context.Context, ev event.Event) error {
		stateMu.Lock()
		defer stateMu.Unlock()
		if closed {
			return ctx.Err()
		}
		if d, ok := ev.Data.(event.AgentCompleteData); ok {
			if d.FinalAnswer != "" {
				answer.Reset()
				answer.WriteString(d.FinalAnswer)
			}
			usage, _ = d.Usage.(*types.TokenUsage)
			done()
		}
		return nil
	})
	bus.On(event.EventError, func(_ context.Context, ev event.Event) error {
		stateMu.Lock()
		defer stateMu.Unlock()
		if closed {
			return ctx.Err()
		}
		runErr = errors.New("generation failed")
		done()
		return nil
	})
	qa := &types.QARequest{Session: session, Query: req.Messages[len(req.Messages)-1].Content, AssistantMessageID: assistant.ID, CustomAgent: a, KnowledgeBaseIDs: append([]string{}, p.KnowledgeBaseIDs...), Mode: types.RunModeUnattended, AssistantAccess: types.AssistantAccessReadOnly, RunNote: "Read-only anonymous API. Only the explicitly configured knowledge bases may be used. No management operations."}
	if a.Config.AgentMode == types.AgentModeSmartReasoning {
		e = h.sessionService.AgentQA(ctx, qa, bus)
	} else {
		e = h.sessionService.KnowledgeQA(ctx, qa, bus)
	}
	if e == nil {
		select {
		case <-turnDone:
		case <-ctx.Done():
			e = ctx.Err()
		}
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	defer func() { closed = true }()
	if e == nil {
		e = runErr
	}
	if ctx.Err() != nil {
		e = ctx.Err()
	}
	if e != nil {
		status := 502
		if errors.Is(e, context.DeadlineExceeded) {
			status = 504
		}
		if !streamStarted {
			anonError(c, status, "Question answering failed")
		} else {
			send(gin.H{"error": gin.H{"message": "Question answering failed", "type": "server_error", "code": http.StatusText(status)}})
			io.WriteString(c.Writer, "data: [DONE]\n\n")
			c.Writer.Flush()
		}
		return
	}
	persistError := func(message string) {
		if !streamStarted {
			anonError(c, 503, message)
			return
		}
		send(gin.H{"error": gin.H{"message": message, "type": "server_error", "code": "persistence_error"}})
		io.WriteString(c.Writer, "data: [DONE]\n\n")
		c.Writer.Flush()
	}
	final, manifest, e := archivalcitations.FromContext(ctx).Finalize(ctx, answer.String())
	if e != nil {
		persistError("Citation persistence failed")
		return
	}
	usageMu.Lock()
	if observedUsage.TotalTokens > 0 {
		copyUsage := observedUsage
		usage = &copyUsage
	}
	usageMu.Unlock()
	assistant.Content = final
	assistant.IsCompleted = true
	assistant.Usage = usage
	if e = h.messageService.UpdateMessage(ctx, assistant); e != nil {
		persistError("Answer persistence failed")
		return
	}
	record := anonymousResponse{ID: id, TenantID: cfg.TenantID, UserID: cfg.UserID, ProfileID: p.ID, ProfileSHA256: profileHash(p), SessionID: session.ID, MessageID: assistant.ID, CreatedAt: time.Now()}
	if e = h.citationStore.DB.WithContext(ctx).Create(&record).Error; e != nil {
		persistError("Response mapping failed")
		return
	}

	text := anonymousText(final)
	citations := publicAnonymousCitations(id, manifest)
	state := "not_applicable"
	unbound := 0
	if manifest != nil {
		state = manifest.State
		unbound = len(manifest.Unbound)
	}
	finish := "stop"
	if truncated {
		finish = "length"
	}
	extra := gin.H{"citations": citations, "citation_status": gin.H{"state": state, "unbound_count": unbound}, "usage": anonymousUsage(usage)}
	if req.Stream {
		extra["final_content"] = text
		chunk(gin.H{}, finish, extra)
		io.WriteString(c.Writer, "data: [DONE]\n\n")
		c.Writer.Flush()
	} else {
		extra["id"] = id
		extra["object"] = "chat.completion"
		extra["created"] = started.Unix()
		extra["model"] = p.ID
		extra["choices"] = []gin.H{{"index": 0, "message": gin.H{"role": "assistant", "content": text}, "finish_reason": finish}}
		c.JSON(200, extra)
	}
}
func anonymousUsage(u *types.TokenUsage) any {
	if u == nil {
		return nil
	}
	return gin.H{"prompt_tokens": u.PromptTokens, "completion_tokens": u.CompletionTokens, "total_tokens": u.TotalTokens}
}
