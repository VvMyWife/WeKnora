package docparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"rsc.io/pdf"
)

// blankPDF creates a standards-compliant, text-free PDF. Blank/scanned pages
// must still count even though they may produce no Markdown.
func blankPDF(count int) []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>"}
	kids := make([]string, count)
	for i := range kids {
		kids[i] = fmt.Sprintf("%d 0 R", i+3)
	}
	objects = append(objects, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), count))
	for range count {
		objects = append(objects, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] >>")
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return out.Bytes()
}

func TestPaddleOCRVLSourcePDFCountBeyondTen(t *testing.T) {
	content := blankPDF(12)
	reader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	if got := reader.NumPage(); got != 12 {
		t.Fatalf("source pages = %d, want 12", got)
	}
}

func TestPaddleOCRVLPageIntegrity(t *testing.T) {
	for _, n := range []int{1, 16, 24} {
		var response paddleOCRVLResponse
		entries := []map[string]any{}
		for i := 0; i < n; i++ {
			entries = append(entries, map[string]any{"prunedResult": map[string]int{"page_index": i}, "markdown": map[string]any{"text": ""}})
		}
		data, _ := json.Marshal(map[string]any{"result": map[string]any{"dataInfo": map[string]int{"numPages": n, "processedPages": n}, "layoutParsingResults": entries}})
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		_, _, pages, _, err := mergeLayoutParsingResult(&response, n)
		if err != nil || pages != n {
			t.Fatalf("blank pages: %d %v", pages, err)
		}
		_, _, _, _, err = mergeLayoutParsingResult(&response, n+1)
		var pe *PDFPageIntegrityError
		if !errors.As(err, &pe) {
			t.Fatal("missing page accepted")
		}
		if n > 1 {
			response.Result.LayoutParsingResults[1].PrunedResult = response.Result.LayoutParsingResults[0].PrunedResult
			_, _, _, _, err = mergeLayoutParsingResult(&response, n)
			if !errors.As(err, &pe) {
				t.Fatal("duplicate page accepted")
			}
		}
		response.Result.LayoutParsingResults = nil
		if _, _, _, _, err = mergeLayoutParsingResult(&response, n); err == nil {
			t.Fatal("empty response accepted")
		}
	}
}
func TestPDFInvalidBeforeOCR(t *testing.T) {
	if _, err := CountPDFPages([]byte("broken pdf")); err == nil {
		t.Fatal("invalid PDF accepted")
	}
}
