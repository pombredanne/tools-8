package config

import (
	"fmt"
)

// Stub file used to track and update versions
const (
	buildhash = "fe0a8a6"
	buildtime = "28 Aug 2015 22:35 UTC"
	semver    = "0.1.4"
)


func Version() string {
	return fmt.Sprintf(
		"ogload-%s (Build: %s, %s)", 
		semver, buildhash, buildtime,
	)
}
