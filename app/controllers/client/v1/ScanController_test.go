package v1

import "testing"

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback int
		expected int
	}{
		{name: "uses fallback for empty value", input: "", fallback: 20, expected: 20},
		{name: "parses positive integer", input: "25", fallback: 20, expected: 25},
		{name: "trims spaces", input: " 3 ", fallback: 1, expected: 3},
		{name: "uses fallback for zero", input: "0", fallback: 1, expected: 1},
		{name: "uses fallback for negative", input: "-1", fallback: 1, expected: 1},
		{name: "uses fallback for non numeric", input: "abc", fallback: 20, expected: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := parsePositiveInt(tt.input, tt.fallback)
			if actual != tt.expected {
				t.Fatalf("parsePositiveInt(%q, %d) = %d, want %d", tt.input, tt.fallback, actual, tt.expected)
			}
		})
	}
}

func TestValidDecision(t *testing.T) {
	for _, decision := range []string{decisionAccept, decisionReject, decisionNeedReview} {
		if !validDecision(decision) {
			t.Fatalf("validDecision(%q) = false, want true", decision)
		}
	}
	if validDecision("UNKNOWN") {
		t.Fatal("validDecision(\"UNKNOWN\") = true, want false")
	}
}
