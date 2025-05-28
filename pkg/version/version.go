package version

// Version holds the current version of Eph.
// This is set via ldflags during build.
var Version = "dev"

// GetVersion returns the current version of the application.
func GetVersion() string {
	return Version
}
