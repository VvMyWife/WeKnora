package session

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tencent/WeKnora/internal/anonymousapi"
	"github.com/Tencent/WeKnora/internal/archivalcitations"
	"github.com/Tencent/WeKnora/internal/event"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type anonUsers struct{ interfaces.UserService }

func (anonUsers) GetUserByID(context.Context, string) (*types.User, error) {
	return &types.User{IsActive: true}, nil
}

type anonMembers struct {
	interfaces.TenantMemberService
	role types.TenantRole
}

func (m *anonMembers) GetMembership(context.Context, string, uint64) (*types.TenantMember, error) {
	return &types.TenantMember{Role: m.role, Status: types.TenantMemberStatusActive}, nil
}

type anonTenants struct{ interfaces.TenantService }

func (anonTenants) GetTenantByID(context.Context, uint64) (*types.Tenant, error) {
	return &types.Tenant{ID: 1}, nil
}

type anonKBs struct {
	interfaces.KnowledgeBaseService
}

func (anonKBs) GetKnowledgeBaseByID(_ context.Context, id string) (*types.KnowledgeBase, error) {
	if id != "kb" {
		return nil, fmt.Errorf("denied")
	}
	return &types.KnowledgeBase{ID: id, TenantID: 1}, nil
}

type anonAgents struct{ interfaces.CustomAgentService }

func (anonAgents) GetAgentByID(context.Context, string) (*types.CustomAgent, error) {
	return &types.CustomAgent{ID: "agent", TenantID: 1, Config: types.CustomAgentConfig{AgentMode: types.AgentModeQuickAnswer, AccessLevel: types.AssistantAccessAutoWrite, MCPSelectionMode: "all", SkillsSelectionMode: "all", WebSearchEnabled: true}}, nil
}

type anonMessages struct {
	interfaces.MessageService
	rows  map[string]*types.Message
	order []*types.Message
}

func (m *anonMessages) CreateMessage(_ context.Context, v *types.Message) (*types.Message, error) {
	m.rows[v.ID] = v
	m.order = append(m.order, v)
	return v, nil
}
func (m *anonMessages) UpdateMessage(_ context.Context, v *types.Message) error {
	m.rows[v.ID] = v
	return nil
}
func (m *anonMessages) GetMessage(_ context.Context, s, id string) (*types.Message, error) {
	v := m.rows[id]
	if v == nil || v.SessionID != s {
		return nil, fmt.Errorf("missing")
	}
	return v, nil
}

type anonSessions struct {
	interfaces.SessionService
	seen      []*types.QARequest
	cancelled bool
	wait      bool
	fail      bool
}

func (s *anonSessions) CreateSession(_ context.Context, v *types.Session) (*types.Session, error) {
	return v, nil
}
func (s *anonSessions) KnowledgeQA(ctx context.Context, r *types.QARequest, b *event.EventBus) error {
	s.seen = append(s.seen, r)
	if r.CustomAgent.Config.AccessLevel != types.AssistantAccessReadOnly || r.CustomAgent.Config.MCPSelectionMode != "none" || r.CustomAgent.Config.SkillsSelectionMode != "none" || r.LocalBrowserEnabled || r.WebSearchEnabled {
		return fmt.Errorf("unsafe options")
	}
	if s.fail {
		b.Emit(ctx, event.Event{Type: event.EventAgentFinalAnswer, Data: event.AgentFinalAnswerData{Content: "partial"}})
		b.Emit(ctx, event.Event{Type: event.EventError})
		return nil
	}
	// Match the real quick-answer pipeline, which returns before its stream finishes.
	go func() {
		if s.wait {
			<-ctx.Done()
			s.cancelled = true
			return
		}
		time.Sleep(10 * time.Millisecond)
		b.Emit(ctx, event.Event{Type: event.EventAgentFinalAnswer, Data: event.AgentFinalAnswerData{Content: "answer", Done: true}})
	}()
	return nil
}
func anonymousFixture(t *testing.T) (*Handler, *gin.Engine, string, *anonMessages, *anonSessions) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "true")
	d := t.TempDir()
	db, e := gorm.Open(sqlite.Open(filepath.Join(d, "test.db")), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&archivalcitations.ManifestRecord{}, &archivalcitations.Record{}, &anonymousResponse{}); e != nil {
		t.Fatal(e)
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	m := &anonMessages{rows: map[string]*types.Message{}}
	s := &anonSessions{}
	h := &Handler{citationStore: archivalcitations.NewStore(db), userService: anonUsers{}, memberService: &anonMembers{role: types.TenantRoleViewer}, tenantService: anonTenants{}, knowledgebaseService: anonKBs{}, customAgentService: anonAgents{}, messageService: m, sessionService: s, redis: rdb}
	cfg := anonymousapi.Config{Enabled: true, TenantID: 1, UserID: "viewer", Profiles: []anonymousapi.Profile{{ID: "qa", AgentID: "agent", KnowledgeBaseIDs: []string{"kb"}, MaxTokens: 100}}}
	b, _ := json.Marshal(cfg)
	p := filepath.Join(d, "config.json")
	os.WriteFile(p, b, 0600)
	t.Setenv("WEKNORA_ANONYMOUS_API_CONFIG", p)
	router := gin.New()
	h.RegisterAnonymousAPI(router)
	return h, router, p, m, s
}
func anonCall(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	q := httptest.NewRequest(method, path, strings.NewReader(body))
	q.RemoteAddr = "127.0.0.1:1234"
	q.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, q)
	return w
}
func TestAnonymousSyncHistoryAndIsolation(t *testing.T) {
	_, r, p, m, s := anonymousFixture(t)
	body := `{"model":"qa","messages":[{"role":"user","content":"old"},{"role":"assistant","content":"oldanswer"},{"role":"user","content":"new"}]}`
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var v map[string]any
	json.Unmarshal(w.Body.Bytes(), &v)
	if m.order[0].RequestID != m.order[1].RequestID || m.order[0].RequestID == m.order[2].RequestID || m.order[2].RequestID != m.order[3].RequestID {
		t.Fatal("history turns collapse or current turn splits")
	}
	if len(m.order[0].RequestID) != 36 {
		t.Fatal("internal request ID exceeds PostgreSQL schema")
	}
	if len(m.order) != 4 || len(s.seen) != 1 || s.seen[0].Query != "new" {
		t.Fatal("history duplicated or lost")
	}
	if !strings.Contains(w.Body.String(), `"content":"answer"`) {
		t.Fatal(w.Body.String())
	}
	id := v["id"].(string)
	if x := anonCall(r, "GET", "/api/openai/v1/responses/"+id+"/citations", ""); x.Code != 200 {
		t.Fatal(x.Body.String())
	}
	if x := anonCall(r, "GET", "/api/openai/v1/responses/private-web-message/citations", ""); x.Code != 404 {
		t.Fatal("private message exposed")
	}
	os.WriteFile(p, []byte(`{"enabled":true,"tenant_id":1,"user_id":"viewer","profiles":[{"id":"qa","agent_id":"agent","knowledge_base_ids":["other"]}]}`), 0600)
	if x := anonCall(r, "GET", "/api/openai/v1/responses/"+id+"/citations", ""); x.Code != 404 {
		t.Fatal("revoked scope accessible")
	}
}
func TestAnonymousSSEAndStrictParameters(t *testing.T) {
	_, r, _, _, _ := anonymousFixture(t)
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}],"stream":true}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "data: [DONE]") || !strings.Contains(w.Body.String(), `"final_content":"answer"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, extra := range []string{`,"tools":[]`, `,"max_tokens":101`} {
		x := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}]`+extra+`}`)
		if x.Code != 400 {
			t.Fatal(x.Code, x.Body.String())
		}
	}
}
func TestAnonymousRejectsElevatedExecutionIdentity(t *testing.T) {
	h, r, _, _, _ := anonymousFixture(t)
	h.memberService = &anonMembers{role: types.TenantRoleOwner}
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}]}`)
	if w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestAnonymousProjectionSplitTags(t *testing.T) {
	p := newAnonymousProjector()
	out := ""
	for _, s := range []string{"fact <k", "b doc=\"A\" chunk_", "id=\"c1\" /> next <kb chunk_id=\"c1\" />"} {
		out += p.Feed(s)
	}
	if out != "fact [1] next [1]" {
		t.Fatal(out)
	}
	if got := anonymousText(`fact <kb chunk_id="c1" citation_id="x" footnote="1" />`); got != "fact [1]" {
		t.Fatal(got)
	}
}
func TestAnonymousCancellation(t *testing.T) {
	_, r, p, _, s := anonymousFixture(t)
	s.wait = true
	var cfg anonymousapi.Config
	b, _ := os.ReadFile(p)
	json.Unmarshal(b, &cfg)
	cfg.TimeoutSeconds = 1
	b, _ = json.Marshal(cfg)
	os.WriteFile(p, b, 0600)
	start := time.Now()
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}]}`)
	if w.Code != 504 || time.Since(start) > 3*time.Second {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestAnonymousNoAuthBypass(t *testing.T) {
	h, _, _, _, _ := anonymousFixture(t)
	r := gin.New()
	h.RegisterAnonymousAPI(r)
	r.Use(func(c *gin.Context) { c.AbortWithStatus(401) })
	r.GET("/api/v1/tenants", func(c *gin.Context) { c.Status(200) })
	if w := anonCall(r, "GET", "/api/v1/tenants", ""); w.Code != 401 {
		t.Fatal("existing authentication bypassed")
	}
	if w := anonCall(r, "GET", "/api/openai/v1/models", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
}
func TestAnonymousProjectionPreservesLiteralCode(t *testing.T) {
	in := "`<kb chunk_id=\"example\"/>`\n```xml\n<kb chunk_id=\"example2\"/>\n```\nFact <kb chunk_id=\"c1\" footnote=\"1\" />"
	out := anonymousText(in)
	if !strings.Contains(out, "example2") || !strings.Contains(out, "example") || !strings.HasSuffix(out, "Fact [1]") {
		t.Fatal(out)
	}
}
func TestAnonymousRestartAndProfileRevocation(t *testing.T) {
	h, r, _, _, _ := anonymousFixture(t)
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}]}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	id := response["id"].(string)
	replacement := gin.New()
	h.RegisterAnonymousAPI(replacement)
	w = anonCall(replacement, "GET", "/api/openai/v1/responses/"+id+"/citations", "")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAnonymousModelFailure(t *testing.T) {
	_, r, _, _, s := anonymousFixture(t)
	s.fail = true
	w := anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}]}`)
	if w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = anonCall(r, "POST", "/api/openai/v1/chat/completions", `{"model":"qa","messages":[{"role":"user","content":"q"}],"stream":true}`)
	if !strings.Contains(w.Body.String(), `"error"`) || !strings.Contains(w.Body.String(), "data: [DONE]") || strings.Contains(w.Body.String(), `"final_content"`) {
		t.Fatal(w.Body.String())
	}
}
