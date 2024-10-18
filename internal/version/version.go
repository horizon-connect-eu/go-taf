package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Version string
	Build   string
)

func init() {
	Version = "development-release"

	if GetVCSModified() == "true" {
		Build = fmt.Sprintf("%s: %s (modified)", GetVCS(), GetVCSHash())
	} else {
		Build = fmt.Sprintf("%s: %s", GetVCS(), GetVCSHash())
	}
}

func GetVCSHash() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}

func GetVCSModified() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.modified" {
				return setting.Value
			}
		}
	}
	return ""
}

func GetVCS() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs" {
				return setting.Value
			}
		}
	}
	return ""
}
