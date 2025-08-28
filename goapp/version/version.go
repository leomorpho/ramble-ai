package version

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// Build-time variables (set via -ldflags during build)
var (
	Version   = "dev"           // Application version
	GitCommit = "unknown"       // Git commit hash
	BuildDate = "unknown"       // Build timestamp
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Get returns the current version information
func Get() Info {
	info := Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: strings.TrimPrefix(runtime.Version(), "go"),
	}

	// If version is still "dev", try to get info from build info
	if info.Version == "dev" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			// Try to get version from build info
			for _, setting := range buildInfo.Settings {
				switch setting.Key {
				case "vcs.revision":
					if len(setting.Value) >= 7 {
						info.GitCommit = setting.Value[:7] // Short commit hash
					}
				case "vcs.time":
					info.BuildDate = setting.Value
				}
			}
			
			// If we have git info but no explicit version, create a dev version
			if info.GitCommit != "unknown" && info.GitCommit != "" {
				info.Version = "dev-" + info.GitCommit
			}
		}
	}

	return info
}

// GetVersion returns just the version string
func GetVersion() string {
	return Get().Version
}

// IsDevBuild returns true if this is a development build
func IsDevBuild() bool {
	return strings.HasPrefix(Version, "dev")
}