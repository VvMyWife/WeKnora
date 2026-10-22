package modelcontext

import "strings"

// CitationDiagnostics contains counts only, never model/document text. A decoder
// owns one instance, so provider chunk boundaries cannot double-count a tag.
type CitationDiagnostics struct {
	Raw, Expanded, Malformed, Unknown, Ineligible, Legacy, Disabled int
}

func (r *sourceRegistry) classifyCitation(tag string, stats *CitationDiagnostics) {
	stats.Raw++
	if isNamedTagStart(strings.ToLower(tag), "kb") || isNamedTagStart(strings.ToLower(tag), "web") {
		stats.Legacy++
		return
	}
	m := refTagRE.FindStringSubmatch(tag)
	if len(m) != 2 {
		stats.Malformed++
		return
	}
	if !r.citationsEnabled {
		stats.Disabled++
		return
	}
	handle := strings.ToLower(m[1])
	_, _, chunk := r.chunks.resolve(handle)
	_, _, web := r.webs.resolve(handle)
	if !chunk && !web {
		stats.Unknown++
		return
	}
	if _, ok := r.citable.Load(handle); !ok {
		stats.Ineligible++
		return
	}
	stats.Expanded++
}

// CitationDiagnostics reports this stream only; reasoning uses another decoder.
func (d *StreamDecoder) CitationDiagnostics() CitationDiagnostics {
	if d == nil || d.sources == nil {
		return CitationDiagnostics{}
	}
	return d.sources.diagnostics
}
