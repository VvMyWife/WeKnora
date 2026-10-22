package modelcontext

import (
	"github.com/Tencent/WeKnora/internal/models/chat"
	"strings"
)

// Static, bounded ASCII text (<128 tokens even under byte-level tokenization).
// One copy in a model-facing tool result, never in canonical history/evidence.
const agentCitationReminder = "\n<cite>Facts need <ref id=\"cN\"/> using the matching chunk id above. No refs for headings. Never invent ids.</cite>"

func (r *Registry) WithAgentCitationReminder(messages []chat.Message) []chat.Message {
	if r == nil || r.sources == nil || !r.sources.citationsEnabled {
		return messages
	}
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		if m.Role != "tool" {
			continue
		}
		if !strings.Contains(m.Content, `<chunk id="c`) && !strings.Contains(m.Content, `<source_binding id="c`) {
			continue
		}
		if strings.Contains(m.Content, agentCitationReminder) {
			return messages
		}
		out := append([]chat.Message(nil), messages...)
		out[i].Content += agentCitationReminder
		return out
	}
	return messages
}
