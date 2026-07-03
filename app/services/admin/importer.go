package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type ImportParseResult struct {
	Total        int              `json:"total"`
	ValidCount   int              `json:"valid_count"`
	InvalidCount int              `json:"invalid_count"`
	Rows         []ImportRowInput `json:"rows"`
}

type ImportRowInput struct {
	RowNumber    int              `json:"row_number"`
	Isbn         string           `json:"isbn"`
	Title        string           `json:"title"`
	Author       string           `json:"author"`
	Publisher    string           `json:"publisher"`
	PublishYear  string           `json:"publish_year"`
	Source       string           `json:"source"`
	MinPrice     *decimal.Decimal `json:"min_price"`
	AvgPrice     *decimal.Decimal `json:"avg_price"`
	MaxPrice     *decimal.Decimal `json:"max_price"`
	SampleCount  int              `json:"sample_count"`
	Confidence   string           `json:"confidence"`
	Valid        bool             `json:"valid"`
	ErrorMessage string           `json:"error_message"`
	RawPayload   string           `json:"raw_payload"`
}

func ParseCSVImportRows(reader io.Reader) (ImportParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return ImportParseResult{}, err
	}
	result := ImportParseResult{}
	if len(records) <= 1 {
		return result, nil
	}
	for i, record := range records[1:] {
		row := parseImportRecord(i+2, record)
		result.Rows = append(result.Rows, row)
		result.Total++
		if row.Valid {
			result.ValidCount++
		} else {
			result.InvalidCount++
		}
	}
	return result, nil
}

func parseImportRecord(rowNumber int, record []string) ImportRowInput {
	row := ImportRowInput{RowNumber: rowNumber}
	value := func(index int) string {
		if index >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[index])
	}
	row.Isbn = NormalizeISBN(value(0))
	row.Title = value(1)
	row.Author = value(2)
	row.Publisher = value(3)
	row.PublishYear = value(4)
	row.Source = value(5)
	if row.Source == "" {
		row.Source = "manual"
	}
	row.Confidence = strings.ToUpper(value(10))
	if row.Confidence == "" {
		row.Confidence = "NONE"
	}

	var errors []string
	if row.Isbn == "" {
		errors = append(errors, "ISBN格式无效")
	}
	if row.Title == "" {
		errors = append(errors, "书名不能为空")
	}
	row.MinPrice = parseDecimalField(value(6), "最低价", &errors)
	row.AvgPrice = parseDecimalField(value(7), "平均价", &errors)
	row.MaxPrice = parseDecimalField(value(8), "最高价", &errors)
	if sampleValue := value(9); sampleValue != "" {
		sampleCount, err := strconv.Atoi(sampleValue)
		if err != nil || sampleCount < 0 {
			errors = append(errors, "样本数格式无效")
		} else {
			row.SampleCount = sampleCount
		}
	}
	if !validConfidence(row.Confidence) {
		errors = append(errors, "可信度无效")
	}

	payload, _ := json.Marshal(record)
	row.RawPayload = string(payload)
	row.Valid = len(errors) == 0
	row.ErrorMessage = strings.Join(errors, "；")
	return row
}

func parseDecimalField(value, name string, errors *[]string) *decimal.Decimal {
	if value == "" {
		return nil
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("%s格式无效", name))
		return nil
	}
	return &amount
}

func NormalizeISBN(isbn string) string {
	normalized := strings.ToUpper(strings.TrimSpace(isbn))
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if len(normalized) == 10 {
		for i := 0; i < 10; i++ {
			ch := normalized[i]
			if ch >= '0' && ch <= '9' {
				continue
			}
			if i == 9 && ch == 'X' {
				continue
			}
			return ""
		}
		return normalized
	}
	if len(normalized) == 13 {
		for i := 0; i < 13; i++ {
			if normalized[i] < '0' || normalized[i] > '9' {
				return ""
			}
		}
		return normalized
	}
	return ""
}

func validConfidence(confidence string) bool {
	return confidence == "HIGH" || confidence == "MEDIUM" || confidence == "LOW" || confidence == "NONE"
}
