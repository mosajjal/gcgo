package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var (
	Version   = "dev"
	GitCommit = "none"
	BuildTime = "unknown"
)

func Info() map[string]string {
	version, gitCommit, buildTime := resolvedMetadata(debug.ReadBuildInfo())

	return map[string]string{
		"version":    version,
		"git_commit": gitCommit,
		"build_time": buildTime,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}

func resolvedMetadata(bi *debug.BuildInfo, ok bool) (string, string, string) {
	version := Version
	gitCommit := GitCommit
	buildTime := BuildTime

	if version == "" || version == "dev" {
		if ok && bi != nil && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			version = bi.Main.Version
		}
	}
	if gitCommit == "" || gitCommit == "none" {
		if value, found := buildSetting(bi, ok, "vcs.revision"); found {
			gitCommit = value
		}
	}
	if buildTime == "" || buildTime == "unknown" {
		if value, found := buildSetting(bi, ok, "vcs.time"); found {
			buildTime = value
		}
	}
	if (gitCommit == "" || gitCommit == "none") || (buildTime == "" || buildTime == "unknown") {
		if pseudoCommit, pseudoBuildTime, found := pseudoVersionMetadata(version); found {
			if gitCommit == "" || gitCommit == "none" {
				gitCommit = pseudoCommit
			}
			if buildTime == "" || buildTime == "unknown" {
				buildTime = pseudoBuildTime
			}
		}
	}

	return version, gitCommit, buildTime
}

func buildSetting(bi *debug.BuildInfo, ok bool, key string) (string, bool) {
	if !ok || bi == nil {
		return "", false
	}
	for _, setting := range bi.Settings {
		if setting.Key == key && setting.Value != "" {
			return setting.Value, true
		}
	}
	return "", false
}

func pseudoVersionMetadata(version string) (string, string, bool) {
	version = strings.TrimSuffix(version, "+incompatible")
	parts := strings.Split(version, "-")
	if len(parts) < 3 {
		return "", "", false
	}

	timestamp := parts[len(parts)-2]
	revision := parts[len(parts)-1]
	if len(timestamp) != 14 || revision == "" {
		return "", "", false
	}

	builtAt, err := time.Parse("20060102150405", timestamp)
	if err != nil {
		return "", "", false
	}

	return revision, builtAt.UTC().Format(time.RFC3339), true
}
