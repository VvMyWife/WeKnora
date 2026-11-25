package externalchunks

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"rsc.io/pdf"
)

const PhysicalPDFMapping = "physical-v1"

func (m Manifest) HasPhysicalPDFMapping() bool {
	var recipe struct {
		Mapping string `json:"pdf_page_mapping"`
	}
	return json.Unmarshal(m.GenerationConfig, &recipe) == nil && recipe.Mapping == PhysicalPDFMapping
}

func pdfMemberPages(member *zip.File) (count int, err error) {
	defer func() {
		if recover() != nil {
			count, err = 0, fmt.Errorf("invalid PDF page tree")
		}
	}()
	input, err := member.Open()
	if err != nil {
		return 0, err
	}
	defer input.Close()
	tmp, err := os.CreateTemp("", "weknora-import-pdf-*")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	size, err := io.Copy(tmp, input)
	if err != nil {
		return 0, err
	}
	document, err := pdf.NewReader(tmp, size)
	if err != nil {
		return 0, err
	}
	count = document.NumPage()
	if count < 1 {
		return 0, fmt.Errorf("PDF has no pages")
	}
	for page := 1; page <= count; page++ {
		if document.Page(page).V.IsNull() {
			return 0, fmt.Errorf("invalid PDF page %d", page)
		}
	}
	return count, nil
}

func (unit *SourceUnit) UnmarshalJSON(data []byte) error {
	type sourceUnit SourceUnit
	value := struct {
		*sourceUnit
		Page *int `json:"page_index"`
	}{sourceUnit: (*sourceUnit)(unit)}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	unit.pageIndexPresent = value.Page != nil
	if value.Page != nil {
		unit.PageIndex = *value.Page
	}
	return nil
}
