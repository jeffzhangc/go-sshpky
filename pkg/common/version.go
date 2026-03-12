package common

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// VersionInfo holds all the version-related information.
type VersionInfo struct {
	Version     string    `json:"version"`      // Version number
	BuildTime   time.Time `json:"build_time"`   // Build timestamp
	GitCommit   string    `json:"git_commit"`   // Git commit hash
	GoVersion   string    `json:"go_version"`   // Go compiler version
	Platform    string    `json:"platform"`     // Build platform
	BuildNumber string    `json:"build_number"` // Build number
	TreeState   string    `json:"tree_state"`   // Git tree state (e.g., "clean" or "dirty")
}

// Build-time variables, injected via ldflags
var (
	version     = "1.0.0"   // Default version, overridden at build time
	buildTime   = "unknown" // Build time
	gitCommit   = "unknown" // Git commit hash
	buildNumber = "unknown" // Build number
	treeState   = "unknown" // Git tree state
)

// GetVersionInfo parses build-time variables and returns a VersionInfo struct.
func GetVersionInfo() VersionInfo {
	var parsedBuildTime time.Time
	if buildTime != "unknown" {
		if t, err := time.Parse(time.RFC3339, buildTime); err == nil {
			parsedBuildTime = t
		} else if t, err := time.Parse("2006-01-02T15:04:05Z", buildTime); err == nil {
			parsedBuildTime = t
		}
	}

	return VersionInfo{
		Version:     version,
		BuildTime:   parsedBuildTime,
		GitCommit:   gitCommit,
		GoVersion:   runtime.Version(),
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		BuildNumber: buildNumber,
		TreeState:   treeState,
	}
}

// String returns a formatted, colorful string representation of the version info.
func (v VersionInfo) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Helper to add a row with color
	addRow := func(key, value, color string) {
		fmt.Fprintf(w, "  %s%s:\t%s%s%s\n", color, key, colorReset, color, value)
	}

	// Title
	fmt.Fprintf(w, "%s%s%s\n", colorGreen, "SSHPKY Version Information", colorReset)
	fmt.Fprintf(w, "%s%s%s\n", colorGreen, "--------------------------", colorReset)

	// Version
	addRow("Version", v.Version, colorCyan)

	// Build Time
	buildTimeStr := "unknown"
	if !v.BuildTime.IsZero() {
		buildTimeStr = v.BuildTime.Format("2006-01-02 15:04:05 MST")
	}
	addRow("Build Time", buildTimeStr, colorPurple)

	// Git Commit
	commit := v.GitCommit
	if len(commit) > 12 {
		commit = commit[:12]
	}
	addRow("Git Commit", commit, colorYellow)

	// Git Tree State
	treeStateColor := colorGreen
	if v.TreeState == "dirty" {
		treeStateColor = colorRed
	}
	addRow("Git Tree", v.TreeState, treeStateColor)

	// Go Version
	addRow("Go Version", v.GoVersion, colorBlue)

	// Platform
	addRow("Platform", v.Platform, colorWhite)

	// Build Number
	addRow("Build Number", v.BuildNumber, colorWhite)

	w.Flush()
	return buf.String()
}

// ShortString returns a compact, single-line version string.
func (v VersionInfo) ShortString() string {
	buildTimeStr := "unknown"
	if !v.BuildTime.IsZero() {
		buildTimeStr = v.BuildTime.Format("2006-01-02")
	}

	commit := v.GitCommit
	if len(commit) > 8 {
		commit = commit[:8]
	}

	return fmt.Sprintf("%s (Build: %s, Commit: %s)",
		v.Version,
		buildTimeStr,
		commit)
}

// IsDevelopment checks if the version is a development build.
func (v VersionInfo) IsDevelopment() bool {
	return v.Version == "dev" || v.Version == "unknown" || strings.HasSuffix(v.Version, "-dev")
}

// GetVersion returns the raw version string.
func GetVersion() string {
	return version
}

// GetBuildTime returns the raw build time string.
func GetBuildTime() string {
	return buildTime
}

// GetGitCommit returns the raw Git commit hash.
func GetGitCommit() string {
	return gitCommit
}

// GetBuildNumber returns the raw build number.
func GetBuildNumber() string {
	return buildNumber
}
