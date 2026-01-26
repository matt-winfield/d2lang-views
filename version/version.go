// Package version provides version information for the d2lang-views tool.
package version

import (
	"runtime/debug"
)

// RepoURL is the GitHub repository URL for this tool.
const RepoURL = "github.com/matt-winfield/d2lang-views"

// version can be set at build time using ldflags:
//
//	go build -ldflags "-X github.com/matt-winfield/d2lang-views/version.version=v1.0.0"
//
// If not set, it will be detected from build info or default to "dev".
var version = ""

// Version returns the current version of the tool.
// It attempts to determine the version in this order:
// 1. Value set via ldflags at build time
// 2. Module version from debug.ReadBuildInfo() (works with `go install`)
// 3. VCS revision from build settings
// 4. Falls back to "dev"
func Version() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	// Check if main module has a version (set when installed via `go install`)
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// Try to get VCS revision from build settings
	var vcsRevision string
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			vcsRevision = setting.Value
			break
		}
	}

	if vcsRevision != "" {
		// Return short hash (first 7 chars)
		if len(vcsRevision) > 7 {
			return vcsRevision[:7]
		}
		return vcsRevision
	}

	return "dev"
}

// GeneratedFileHeader returns the header comment to be placed at the start of generated D2 files.
func GeneratedFileHeader() string {
	return "# Generated using " + RepoURL + " " + Version() + ". DO NOT EDIT MANUALLY.\n"
}

// GeneratedViewHeader returns the comment to be placed at the start of each generated view.
func GeneratedViewHeader() string {
	return "# Generated using " + RepoURL + " " + Version() + ". DO NOT EDIT MANUALLY.\n"
}
