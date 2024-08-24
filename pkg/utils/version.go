package utils

import (
	"fmt"
)

var (
	BuildTS   = "None"
	GitHash   string
	GitBranch = "None"
	Version   = "None"
)

// getVersion prints build version.
func getVersion() string {
	if GitHash != "" {
		h := GitHash
		if len(h) > 7 {
			h = h[:7]
		}
		return fmt.Sprintf("%s-%s", GitBranch, h)
	}
	return Version
}

// PrintFullVersion print full version
func PrintFullVersion() {
	fmt.Println("Version:          ", getVersion())
	fmt.Println("Git Branch:       ", GitBranch)
	fmt.Println("Git Commit:       ", GitHash)
	fmt.Println("Build Time (UTC): ", BuildTS)
	fmt.Println("")
}
