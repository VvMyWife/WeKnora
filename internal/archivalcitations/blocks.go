package archivalcitations

import (
	"regexp"
	"strings"
	"unicode"
)

type answerBlock struct {
	start, end int
	kind, text string
}

var listStart = regexp.MustCompile(`^\s*(?:[-+*]|\d+[.)、])\s+`)
var headingLine = regexp.MustCompile(`^#{1,6}\s+`)

// Only explicit non-factual scaffolding is exempt. No semantic guesses, broad
// prefix matching, or "short text" exemption: unfamiliar content needs review.
func structuralText(text string) bool {
	v := strings.TrimSpace(text)
	v = strings.TrimSpace(strings.Trim(v, "*"))
	v = strings.TrimRight(v, ":：")
	switch v {
	case "具体要求如下", "以下是相关要求", "根据检索结果，相关要求如下", "数据备份系统要求", "备用网络系统要求", "总结", "小结", "结论", "下面分别说明", "以上为相关资料的整理", "需要我进一步说明吗？":
		return true
	}
	return false
}

// Recognize paragraphs, list items (including indented continuations) and
// GFM table rows. A soft line break is not a new factual block.
func answerBlocks(answer string) []answerBlock {
	lines := strings.Split(answer, "\n")
	out := []answerBlock{}
	fence := ""
	for i := 0; i < len(lines); {
		v := strings.TrimSpace(lines[i])
		if strings.HasPrefix(v, "```") || strings.HasPrefix(v, "~~~") {
			ch := v[:1]
			n := 0
			for n < len(v) && v[n:n+1] == ch {
				n++
			}
			if fence == "" {
				fence = v[:n]
			} else if strings.HasPrefix(v, fence) {
				fence = ""
			}
			i++
			continue
		}
		if fence != "" || v == "" || headingLine.MatchString(v) || tableDivider.MatchString(v) || v == "---" || v == "***" || structuralText(v) {
			i++
			continue
		}
		if i+1 < len(lines) && tableDivider.MatchString(lines[i+1]) {
			i++
			continue
		}
		start, kind := i, "paragraph"
		if strings.HasPrefix(v, "|") && strings.HasSuffix(v, "|") {
			kind = "table_row"
		} else if listStart.MatchString(lines[i]) {
			kind = "list_item"
		}
		i++
		for kind != "table_row" && i < len(lines) {
			next := strings.TrimSpace(lines[i])
			if next == "" || headingLine.MatchString(next) || listStart.MatchString(lines[i]) || strings.HasPrefix(next, "|") || strings.HasPrefix(next, "```") || strings.HasPrefix(next, "~~~") || next == "---" || next == "***" || structuralText(next) {
				break
			}
			if kind == "list_item" && lines[i] == strings.TrimLeft(lines[i], " \t") {
				break
			}
			i++
		}
		out = append(out, answerBlock{start, i - 1, kind, strings.Join(lines[start:i], "\n")})
	}
	return out
}

func citationTagsOutsideCode(text string) []string {
	return kbTag.FindAllString(inlineCode.ReplaceAllString(text, ""), -1)
}

// A reference cannot cover factual text that follows it. Detect such tails
// locally, without interpreting what the cited source semantically proves.
func hasUncitedTail(text string) bool {
	v := inlineCode.ReplaceAllString(text, "")
	all := regexp.MustCompile(`(?is)<(?:kb|web)\b[^>]*>`).FindAllStringIndex(v, -1)
	if len(all) == 0 {
		return false
	}
	tail := v[all[len(all)-1][1]:]
	for _, r := range tail {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func reviewReason(code string) string {
	switch code {
	case "invalid_reference":
		return "引用标识无效或不属于本轮已登记证据；不代表知识库没有来源"
	case "evidence_capture_failed":
		return "本轮存在证据登记失败，此内容尚未建立可验证绑定"
	default:
		return "此内容未提供有效的生成时引用；不代表知识库没有来源"
	}
}
