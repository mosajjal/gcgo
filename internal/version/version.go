package version

import (
	"runtime"
	"runtime/debug"
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
