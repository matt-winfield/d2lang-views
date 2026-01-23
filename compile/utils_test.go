package compile

import (
	"strings"
	"testing"
)

func TestFindObjectById_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		searchId      string
		expectedId    string
		expectedLabel string
		shouldFind    bool
	}{
		{
			name: "exact match",
			content: `
Test: "Test Label"
`,
			searchId:      "Test",
			expectedId:    "Test",
			expectedLabel: "Test Label",
			shouldFind:    true,
		},
		{
			name: "lowercase search for uppercase object",
			content: `
Test: "Test Label"
`,
			searchId:      "test",
			expectedId:    "Test",
			expectedLabel: "Test Label",
			shouldFind:    true,
		},
		{
			name: "uppercase search for lowercase object",
			content: `
test: "Test Label"
`,
			searchId:      "TEST",
			expectedId:    "test",
			expectedLabel: "Test Label",
			shouldFind:    true,
		},
		{
			name: "mixed case search",
			content: `
SomeThing: "Something Label"
`,
			searchId:      "something",
			expectedId:    "SomeThing",
			expectedLabel: "Something Label",
			shouldFind:    true,
		},
		{
			name: "nested object case insensitive",
			content: `
Parent: "Parent Label" {
    Child: "Child Label"
}
`,
			searchId:      "parent.child",
			expectedId:    "Child",
			expectedLabel: "Child Label",
			shouldFind:    true,
		},
		{
			name: "deeply nested case insensitive",
			content: `
First: {
    Second: {
        Third: "Deep Label"
    }
}
`,
			searchId:      "first.second.third",
			expectedId:    "Third",
			expectedLabel: "Deep Label",
			shouldFind:    true,
		},
		{
			name: "not found",
			content: `
Test: "Test Label"
`,
			searchId:   "nonexistent",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			graph, _, err := CompileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			obj, err := FindObjectById(graph, tt.searchId)

			if tt.shouldFind {
				if err != nil {
					t.Fatalf("expected to find object, but got error: %v", err)
				}
				if obj.ID != tt.expectedId {
					t.Fatalf("expected object ID '%s', got '%s'", tt.expectedId, obj.ID)
				}
				if obj.Label.Value != tt.expectedLabel {
					t.Fatalf("expected object label '%s', got '%s'", tt.expectedLabel, obj.Label.Value)
				}
			} else {
				if err == nil {
					t.Fatalf("expected not to find object, but found: %s", obj.ID)
				}
			}
		})
	}
}
