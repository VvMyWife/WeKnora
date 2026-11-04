package session

import (
	"fmt"
	"github.com/Tencent/WeKnora/internal/anonymousapi"
	"github.com/Tencent/WeKnora/internal/archivalcitations"
	"github.com/gin-gonic/gin"
	"html"
	"regexp"
	"strings"
)

var anonKB = regexp.MustCompile(`(?is)<kb\b[^>]*>`)
var anonNumber = regexp.MustCompile(`\bfootnote="([0-9]+)"`)
var anonChunk = regexp.MustCompile(`\bchunk_id="([^"]+)"`)

func anonymousText(s string) string {
	return archivalcitations.ProjectCitationTags(s, func(tag string) string {
		m := anonNumber.FindStringSubmatch(tag)
		if len(m) == 2 {
			return "[" + m[1] + "]"
		}
		return ""
	})
}

type anonymousProjector struct {
	pending string
	numbers map[string]int
}

func newAnonymousProjector() *anonymousProjector {
	return &anonymousProjector{numbers: map[string]int{}}
}

// Only complete citation tags are projected; tag fragments never leave the API.
func (p *anonymousProjector) Feed(s string) string {
	p.pending += s
	var out strings.Builder
	for len(p.pending) > 0 {
		i := strings.IndexByte(p.pending, '<')
		if i < 0 {
			out.WriteString(p.pending)
			p.pending = ""
			break
		}
		out.WriteString(p.pending[:i])
		p.pending = p.pending[i:]
		low := strings.ToLower(p.pending)
		if len(low) < 4 && strings.HasPrefix("<kb ", low) {
			break
		}
		if !strings.HasPrefix(low, "<kb ") && !strings.HasPrefix(low, "<kb\t") && !strings.HasPrefix(low, "<kb\n") {
			out.WriteByte('<')
			p.pending = p.pending[1:]
			continue
		}
		end := strings.IndexByte(p.pending, '>')
		if end < 0 {
			if len(p.pending) > 16384 {
				p.pending = ""
			}
			break
		}
		tag := p.pending[:end+1]
		p.pending = p.pending[end+1:]
		m := anonChunk.FindStringSubmatch(tag)
		if len(m) == 2 {
			key := html.UnescapeString(m[1])
			n := p.numbers[key]
			if n == 0 {
				n = len(p.numbers) + 1
				p.numbers[key] = n
			}
			out.WriteString(fmt.Sprintf("[%d]", n))
		}
	}
	return out.String()
}
func publicAnonymousCitations(responseID string, m *archivalcitations.Manifest) []gin.H {
	out := []gin.H{}
	if m == nil {
		return out
	}
	for _, s := range m.Citations {
		out = append(out, gin.H{"id": s.ID, "number": s.Number, "title": s.Title, "source_kind": s.SourceKind, "location": s.Location, "content_revision": s.ContentRevision, "content_sha256": s.ContentSHA256, "archive_version_id": s.ArchiveVersionID, "archive_file_name": s.ArchiveFileName, "detail_url": "/api/openai/v1/responses/" + responseID + "/citations/" + s.ID})
	}
	return out
}
func (h *Handler) anonymousCitations(c *gin.Context) {
	cfg := c.MustGet("anonymous_config").(*anonymousapi.Config)
	if h.citationStore == nil {
		anonError(c, 503, "Citation service unavailable")
		return
	}
	var r anonymousResponse
	if e := h.citationStore.DB.WithContext(c.Request.Context()).Where("id=? AND tenant_id=? AND user_id=?", c.Param("response_id"), cfg.TenantID, cfg.UserID).Take(&r).Error; e != nil {
		anonError(c, 404, "Response unavailable")
		return
	}
	p, ok := cfg.Profile(r.ProfileID)
	if !ok || profileHash(p) != r.ProfileSHA256 {
		anonError(c, 404, "Response is no longer published")
		return
	}
	ctx, _, e := h.anonymousIdentity(c.Request.Context(), cfg, p)
	if e != nil {
		anonError(c, 404, "Response is no longer published")
		return
	}
	m, e := h.messageService.GetMessage(ctx, r.SessionID, r.MessageID)
	if e != nil {
		anonError(c, 404, "Response unavailable")
		return
	}
	manifest, e := h.citationStore.Manifest(ctx, cfg.TenantID, r.SessionID, r.MessageID, m.Content)
	if e != nil {
		anonError(c, 404, "Citations unavailable")
		return
	}
	allowed := map[string]bool{}
	for _, k := range p.KnowledgeBaseIDs {
		allowed[k] = true
	}
	for _, s := range manifest.Citations {
		if !allowed[s.KnowledgeBaseID] {
			anonError(c, 404, "Citation scope no longer published")
			return
		}
	}
	if c.Param("citation_id") == "" {
		c.JSON(200, gin.H{"id": r.ID, "citations": publicAnonymousCitations(r.ID, &manifest), "citation_status": gin.H{"state": manifest.State, "unbound_count": len(manifest.Unbound)}})
		return
	}
	row, s, e := h.citationStore.Evidence(ctx, cfg.TenantID, r.SessionID, r.MessageID, c.Param("citation_id"))
	if e != nil || !allowed[s.KnowledgeBaseID] {
		anonError(c, 404, "Citation unavailable")
		return
	}
	var count int64
	if s.WikiPageID != "" {
		e = h.citationStore.DB.WithContext(ctx).Table("wiki_pages").Where("id=? AND tenant_id=? AND knowledge_base_id=? AND deleted_at IS NULL", s.WikiPageID, cfg.TenantID, s.KnowledgeBaseID).Count(&count).Error
	} else {
		e = h.citationStore.DB.WithContext(ctx).Table("chunks").Where("id=? AND tenant_id=? AND knowledge_id=? AND knowledge_base_id=? AND deleted_at IS NULL", s.ChunkID, cfg.TenantID, s.KnowledgeID, s.KnowledgeBaseID).Count(&count).Error
	}
	if e != nil || count != 1 {
		anonError(c, 404, "Citation source unavailable")
		return
	}
	c.JSON(200, gin.H{"id": s.ID, "number": s.Number, "title": s.Title, "source_kind": s.SourceKind, "location": s.Location, "snapshot": row.Snapshot, "snapshot_at": row.CreatedAt, "content_revision": s.ContentRevision, "content_sha256": s.ContentSHA256, "archive_version_id": s.ArchiveVersionID, "archive_file_name": s.ArchiveFileName})
}
