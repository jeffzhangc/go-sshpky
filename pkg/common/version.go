package commom

import (
	"fmt"
	"runtime"
	"time"
)

// VersionInfo 版本信息结构体
type VersionInfo struct {
	Version     string    `json:"version"`      // 版本号
	BuildTime   time.Time `json:"build_time"`   // 构建时间
	GitCommit   string    `json:"git_commit"`   // Git提交哈希
	GoVersion   string    `json:"go_version"`   // Go版本
	Platform    string    `json:"platform"`     // 构建平台
	BuildNumber string    `json:"build_number"` // 构建编号
}

// 编译时注入的变量
var (
	version     = "1.0.0"   // 默认版本号，编译时会被覆盖
	buildTime   = "unknown" // 构建时间
	gitCommit   = "unknown" // Git提交哈希
	buildNumber = "unknown" // 构建编号
	treeState   = "unknown" // Git树状态
)

// GetVersionInfo 获取版本信息
func GetVersionInfo() VersionInfo {
	// 解析构建时间
	var parsedBuildTime time.Time
	if buildTime != "unknown" {
		if t, err := time.Parse(time.RFC3339, buildTime); err == nil {
			parsedBuildTime = t
		} else {
			// 尝试其他格式
			if t, err := time.Parse("2006-01-02T15:04:05Z", buildTime); err == nil {
				parsedBuildTime = t
			}
		}
	}

	return VersionInfo{
		Version:     version,
		BuildTime:   parsedBuildTime,
		GitCommit:   gitCommit,
		GoVersion:   runtime.Version(),
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		BuildNumber: buildNumber,
	}
}

// String 返回版本信息的字符串表示
func (v VersionInfo) String() string {
	buildTimeStr := "unknown"
	if !v.BuildTime.IsZero() {
		buildTimeStr = v.BuildTime.Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf(`Version: %s
Build Time: %s
Git Commit: %s
Go Version: %s
Platform: %s
Build Number: %s`,
		v.Version,
		buildTimeStr,
		v.GitCommit,
		v.GoVersion,
		v.Platform,
		v.BuildNumber)
}

// ShortString 返回简短的版本信息
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
		commit) // 只显示前8位提交哈希
}

// IsDevelopment 判断是否为开发版本
func (v VersionInfo) IsDevelopment() bool {
	return v.Version == "dev" || v.Version == "unknown"
}

// GetVersion 获取版本号
func GetVersion() string {
	return version
}

// GetBuildTime 获取构建时间
func GetBuildTime() string {
	return buildTime
}

// GetGitCommit 获取Git提交哈希
func GetGitCommit() string {
	return gitCommit
}

// GetBuildNumber 获取构建编号
func GetBuildNumber() string {
	return buildNumber
}
