package version

import "runtime"

// Version holds the current version of Eph.
// This is set via ldflags during build.
var Version = "dev"

// GitCommit holds the git commit hash.
// This is set via ldflags during build.
var GitCommit = "unknown"

// BuildDate holds the date when the binary was built.
// This is set via ldflags during build.
var BuildDate = "unknown"

// GoVersion holds the Go runtime version.
var GoVersion = runtime.Version()

func GetVersion() string {
	return Version
}
