package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	htmltomd "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	htmlTagPattern        = regexp.MustCompile(`<[^>]+>`)
	codeBlockPattern      = regexp.MustCompile("(?s)^\\s*```[a-zA-Z]*\\s*\n(.*?)\n\\s*```\\s*$")
	htmlDocPattern        = regexp.MustCompile(`(?i)^\s*(<\!DOCTYPE|<html|<body|<div|<p[\s>]|<table|<h[1-6][\s>])`)
	multipleNewlines      = regexp.MustCompile(`\n{3,}`)
	reasoningBlockPattern = regexp.MustCompile(`(?is)<(?:think|analysis|reasoning)\b[^>]*>.*?</(?:think|analysis|reasoning)\s*>`)
	reasoningOpenPattern  = regexp.MustCompile(`(?is)<(?:think|analysis|reasoning)\b[^>]*>`)
	reasoningClosePattern = regexp.MustCompile(`(?is)</(?:think|analysis|reasoning)\s*>`)
	knownEmptyReplies     = []string{
		"无文字内容",
		"无法识别",
		"no text",
		"no text content",
		"no content",
		"empty",
		"图片中没有文字",
		"图片中没有可识别的文字",
	}
)

// OCRSanitizationResult separates stored evidence from model control text.
type OCRSanitizationResult struct {
	Content  string
	Rejected bool
	Reason   string
}

// sanitizeOCRText retains the legacy simple API for temporary attachments.
func sanitizeOCRText(raw string) string {
	return sanitizeOCRTextWithResult(raw).Content
}

// sanitizeOCRTextWithResult rejects malformed reasoning markup and strips
// closed reasoning blocks before any output is persisted or embedded.
func sanitizeOCRTextWithResult(raw string) OCRSanitizationResult {
	text := strings.TrimSpace(raw)
	if text == "" {
		return OCRSanitizationResult{}
	}
	if reasoningOpenPattern.MatchString(text) {
		if !reasoningClosePattern.MatchString(text) {
			return OCRSanitizationResult{Rejected: true, Reason: "unterminated_reasoning_block"}
		}
		text = reasoningBlockPattern.ReplaceAllString(text, "")
	}
	if reasoningOpenPattern.MatchString(text) || reasoningClosePattern.MatchString(text) {
		return OCRSanitizationResult{Rejected: true, Reason: "reasoning_markup_remaining"}
	}
	return OCRSanitizationResult{Content: sanitizeOCRTextCore(text)}
}

func sanitizeOCRTextCore(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}

	text = stripMarkdownCodeBlock(text)

	// If stripping HTML tags leaves almost no text, the response is useless
	// (e.g. "<html><body><div class="image"><img/></div></body></html>").
	plainText := strings.TrimSpace(htmlTagPattern.ReplaceAllString(text, ""))
	if len(plainText) < 10 && htmlTagPattern.MatchString(text) {
		return ""
	}

	if looksLikeHTML(text) {
		text = ocrHTMLToMarkdown(text)
		text = strings.TrimSpace(text)
		if text == "" {
			return ""
		}
	}

	if isKnownEmptyReply(text) {
		return ""
	}

	text = multipleNewlines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

// writeVLMOCRAudit writes sensitive raw OCR only to an opt-in private mount.
func writeVLMOCRAudit(knowledgeID, raw string, result OCRSanitizationResult) {
	root := strings.TrimSpace(os.Getenv("WEKNORA_VLM_AUDIT_DIR"))
	if root == "" {
		return
	}
	sum := sha256.Sum256([]byte(knowledgeID))
	dir := filepath.Join(root, hex.EncodeToString(sum[:8]))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	record := struct {
		At       time.Time `json:"at"`
		Raw      string    `json:"raw"`
		Clean    string    `json:"clean"`
		Rejected bool      `json:"rejected"`
		Reason   string    `json:"reason,omitempty"`
	}{At: time.Now().UTC(), Raw: raw, Clean: result.Content, Rejected: result.Rejected, Reason: result.Reason}
	data, err := json.Marshal(record)
	if err != nil {
		return
	}
	name := time.Now().UTC().Format("20060102T150405.000000000Z") + ".json"
	tmp := filepath.Join(dir, "."+name+".tmp")
	if err := os.WriteFile(tmp, data, 0o600); err == nil {
		_ = os.Rename(tmp, filepath.Join(dir, name))
	}
}

// stripMarkdownCodeBlock removes a markdown code-fence wrapper that some
// models add around their output (e.g. ```html\n...\n``` or ```markdown\n...\n```).
func stripMarkdownCodeBlock(text string) string {
	if m := codeBlockPattern.FindStringSubmatch(text); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return text
}

// looksLikeHTML returns true when the text appears to be an HTML document
// or contains a significant amount of HTML tags.
func looksLikeHTML(text string) bool {
	if htmlDocPattern.MatchString(text) {
		return true
	}
	tags := htmlTagPattern.FindAllString(text, -1)
	if len(tags) == 0 {
		return false
	}
	tagChars := 0
	for _, t := range tags {
		tagChars += len(t)
	}
	return float64(tagChars)/float64(len(text)) > 0.3
}

// ocrHTMLToMarkdown converts HTML content to markdown, falling back to the
// original text on failure.
func ocrHTMLToMarkdown(content string) string {
	md, err := htmltomd.ConvertString(content)
	if err != nil {
		return content
	}
	return md
}

// isKnownEmptyReply checks whether the text matches a known "no content"
// reply pattern that VLM models produce when the image has no text.
// Trailing punctuation (., !, ?) is stripped before comparison so that
// responses like "No text content." still match "no text content".
func isKnownEmptyReply(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	lower = strings.TrimRight(lower, ".!?。！？")
	for _, phrase := range knownEmptyReplies {
		if lower == strings.ToLower(phrase) {
			return true
		}
	}
	return false
}
