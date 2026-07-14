package wrkcli

import "strings"

// ShellSafeQuote encodes s as a single POSIX shell word using single quotes.
// Empty string becomes ''; embedded single quotes use the '\'' join form.
func ShellSafeQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
