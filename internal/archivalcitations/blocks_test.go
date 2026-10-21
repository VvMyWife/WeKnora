package archivalcitations

import (
	"context"
	"strings"
	"testing"
)

func TestBlocksSeparateListAndTableAndKeepParagraphSoftBreaks(t *testing.T) {
	answer := "# 标题\n具体要求如下：\n\n事实第一行\n事实第二行\n\n- 条目一\n  继续一\n- 条目二\n\n| 名称 | 数值 |\n| --- | --- |\n| 甲 | 12 |\n| 乙 | 13 |\n\n```go\n秘密不应成为事实\n```"
	b := answerBlocks(answer)
	if len(b) != 5 || b[0].start != 3 || b[0].end != 4 || b[1].kind != "list_item" || b[1].end != 7 || b[3].kind != "table_row" {
		t.Fatalf("%+v", b)
	}
}

func TestBlockBindingKeepsVersionsAndDoesNotCoverUncitedTail(t *testing.T) {
	s := requestFixture(t)
	answer := "具体要求如下：\n\n第一行事实\n第二行事实<kb chunk_id=\"c1\" />\n\n- 已有事实<kb chunk_id=\"c2\" />\n- 没有依据\n\n已引事实<kb chunk_id=\"c1\" />后续未引用事实"
	out, m, err := s.Finalize(context.Background(), answer)
	if err != nil || len(m.Citations) != 2 || len(m.Bindings) != 4 || len(m.Unbound) != 2 || m.RepairAttempted {
		t.Fatalf("%s %+v %v", out, m, err)
	}
	if m.Bindings[0].EndIndex != 3 || len(m.Bindings[0].CitationIDs) != 1 {
		t.Fatal(m.Bindings)
	}
	if strings.Count(out, "引用待核验") != 2 || !strings.Contains(out, "具体要求如下：\n\n") {
		t.Fatal(out)
	}
}

func TestInvalidSourceAndCaptureFailureHaveDifferentReasons(t *testing.T) {
	s := requestFixture(t)
	_, m, err := s.Finalize(context.Background(), `事实<kb chunk_id="not-offered" />`)
	if err != nil || len(m.Unbound) != 1 || m.Unbound[0].ReasonCode != "invalid_reference" {
		t.Fatal(m, err)
	}
	s = requestFixture(t)
	s.evidence = map[string]offeredEvidence{}
	s.NoteCaptureFailure("failed-source")
	_, m, err = s.Finalize(context.Background(), "未成功登记的事实")
	if err != nil || m.State == "not_applicable" || m.Unbound[0].ReasonCode != "evidence_capture_failed" {
		t.Fatal(m, err)
	}
}

func TestCodeCitationDoesNotBindProseAndUnknownTransitionsStayConservative(t *testing.T) {
	s := requestFixture(t)
	_, m, err := s.Finalize(context.Background(), "这句话没有证据。`<kb chunk_id=\"c1\" />`\n\n依据标准，此项必须每日执行。")
	if err != nil || len(m.Citations) != 0 || len(m.Unbound) != 2 {
		t.Fatal(m, err)
	}
}
