package version

import (
	"strings"
	"testing"
)

func TestGeneratedFileHeader(t *testing.T) {
	header := GeneratedFileHeader()

	if !strings.HasPrefix(header, "# Generated using github.com/matt-winfield/d2lang-views") {
		t.Errorf("header should start with correct prefix, got: %s", header)
	}

	if !strings.Contains(header, Version()) {
		t.Errorf("header should contain version, got: %s", header)
	}

	if !strings.Contains(header, "DO NOT EDIT MANUALLY") {
		t.Errorf("header should contain 'DO NOT EDIT MANUALLY', got: %s", header)
	}

	if !strings.HasSuffix(header, "\n") {
		t.Errorf("header should end with newline, got: %s", header)
	}
}

func TestGeneratedViewHeader(t *testing.T) {
	header := GeneratedViewHeader()

	if !strings.HasPrefix(header, "# Generated using github.com/matt-winfield/d2lang-views") {
		t.Errorf("header should start with correct prefix, got: %s", header)
	}

	if !strings.Contains(header, Version()) {
		t.Errorf("header should contain version, got: %s", header)
	}

	if !strings.Contains(header, "DO NOT EDIT MANUALLY") {
		t.Errorf("header should contain 'DO NOT EDIT MANUALLY', got: %s", header)
	}

	if !strings.HasSuffix(header, "\n") {
		t.Errorf("header should end with newline, got: %s", header)
	}
}

func TestRepoURL(t *testing.T) {
	expected := "github.com/matt-winfield/d2lang-views"
	if RepoURL != expected {
		t.Errorf("expected RepoURL to be %q, got %q", expected, RepoURL)
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version() should not return empty string")
	}
	// Version should be either "dev", a semantic version, or a commit hash
	// Just verify it's non-empty and doesn't contain unexpected characters
	if strings.ContainsAny(v, "\n\t ") {
		t.Errorf("Version() should not contain whitespace, got: %q", v)
	}
}
