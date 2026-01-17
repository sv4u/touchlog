package cli

import (
	"testing"
)

// TestParseCSV tests the parseCSV function behavior
func TestParseCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single value",
			input:    "value1",
			expected: []string{"value1"},
		},
		{
			name:     "multiple values",
			input:    "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "values with spaces",
			input:    "value1, value2 , value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "empty values filtered",
			input:    "value1,,value2, ,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "only spaces",
			input:    " , , ",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "value1,value2,",
			expected: []string{"value1", "value2"},
		},
		{
			name:     "leading comma",
			input:    ",value1,value2",
			expected: []string{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCSV(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d values, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected value[%d] = %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}
