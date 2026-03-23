package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// RemotePathInfo holds parsed information about a remote path argument.
type RemotePathInfo struct {
	HostAlias string // SSH config host alias
	Path      string // The path part
	Original  string // The original string argument
	IsRemote  bool   // True if it's a remote path
	User      string // Optional username
}

// ParseRemotePath parses a path argument like "my-server:/remote/path" or "user@my-server:/remote/path".
func ParseRemotePath(pathArg string) *RemotePathInfo {
	info := &RemotePathInfo{Original: pathArg}
	parts := strings.SplitN(pathArg, ":", 2)

	// Not a remote path if:
	// - no colon
	// - colon is the first character (e.g. ":/path")
	// - part before colon contains a slash (likely a windows path like "C:/Users")
	if len(parts) != 2 || parts[0] == "" || strings.Contains(parts[0], "/") {
		info.IsRemote = false
		info.Path = pathArg
		return info
	}

	info.IsRemote = true
	info.Path = parts[1]

	hostPart := parts[0]
	if userAtIdx := strings.LastIndex(hostPart, "@"); userAtIdx != -1 {
		info.User = hostPart[:userAtIdx]
		info.HostAlias = hostPart[userAtIdx+1:]
	} else {
		info.HostAlias = hostPart
	}

	return info
}

// ExpandPath expands a leading "~/" to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
