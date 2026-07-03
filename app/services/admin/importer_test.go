package admin

import (
	"strings"
	"testing"
)

func TestParseCSVImportRowsReportsValidAndInvalidRows(t *testing.T) {
	csvData := "isbn,title,author,publisher,publish_year,source,min_price,avg_price,max_price,sample_count,confidence\n" +
		"9787111128069,示例书,作者,出版社,2018,manual,10.00,20.00,30.00,5,HIGH\n" +
		"bad-isbn,坏数据,,,,manual,1.00,abc,2.00,0,NONE\n"

	result, err := ParseCSVImportRows(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ParseCSVImportRows err=%v", err)
	}
	if result.Total != 2 {
		t.Fatalf("total = %d, want 2", result.Total)
	}
	if result.ValidCount != 1 || result.InvalidCount != 1 {
		t.Fatalf("valid/invalid = %d/%d, want 1/1", result.ValidCount, result.InvalidCount)
	}
	if !result.Rows[0].Valid {
		t.Fatalf("first row should be valid: %s", result.Rows[0].ErrorMessage)
	}
	if result.Rows[1].Valid {
		t.Fatal("second row should be invalid")
	}
}
