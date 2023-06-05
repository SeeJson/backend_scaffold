package version

import (
	"fmt"
	"os"
)

// Version information.
var (
	ReleaseVersion = "None"
	BuildTS        = "None"
	GitHash        = "None"
	GitBranch      = "None"
	GoVersion      = "None"
)

func Version() string {
	return fmt.Sprintf("Release Version: %s\n"+
		"Git Commit Hash: %s\n"+
		"Git Branch: %s\n"+
		"UTC Build Time: %s\n"+
		"GoVersion: %s\n",
		ReleaseVersion,
		GitHash,
		GitBranch,
		BuildTS,
		GoVersion,
	)
}

func HandleVersion() {
	for _, arg := range os.Args[1:] {
		if arg == "-V" || arg == "--version" {
			print(Version())
			os.Exit(0)
		}
	}
}
