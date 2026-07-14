package wrkcli

import (
	_ "embed"
	"strings"
)

//go:embed VERSION.txt
var versionFile string

// Version returns the embedded build version (e.g. "v0.0.1").
func Version() string {
	return "v" + strings.TrimSpace(versionFile)
}