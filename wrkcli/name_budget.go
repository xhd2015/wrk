package wrkcli

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Path/branch last-component budget (git/OS ENAMETOOLONG class).
const (
	nameMaxComponentBytes  = 255
	nameSuffixReserveBytes = 3 // reserve for "-99" collision suffix
)

// fitTaskSlugForNames shortens slug so preferred path basename and branch
// (without -N) fit within nameMaxComponentBytes-nameSuffixReserveBytes.
// Shortens the slug only; never chops basename/token/date.
// If the prefix alone (no slug) exceeds the budget, returns a clear wrk error.
// Empty slug still validates the no-slug prefix.
func fitTaskSlugForNames(basename, pathToken, date, slug string) (string, error) {
	maxLen := nameMaxComponentBytes - nameSuffixReserveBytes
	pathBase := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	branchBase := pathToken + "-" + date

	if len(pathBase) > maxLen {
		return "", fmt.Errorf("wrk: worktree path name too long: component would exceed %d bytes (prefix %d bytes exceeds budget of %d with -N reserve; basename/token cannot be shortened)",
			nameMaxComponentBytes, len(pathBase), maxLen)
	}
	if len(branchBase) > maxLen {
		return "", fmt.Errorf("wrk: branch name too long: would exceed %d bytes (prefix %d bytes exceeds budget of %d with -N reserve)",
			nameMaxComponentBytes, len(branchBase), maxLen)
	}
	if slug == "" {
		return "", nil
	}

	// path: pathBase + "-" + slug ≤ maxLen
	maxSlugPath := maxLen - len(pathBase) - 1
	maxSlugBranch := maxLen - len(branchBase) - 1
	maxSlug := maxSlugPath
	if maxSlugBranch < maxSlug {
		maxSlug = maxSlugBranch
	}
	if maxSlug <= 0 {
		// No room for a slug segment; drop it so names still fit.
		return "", nil
	}
	fitted := truncateToMaxBytes(slug, maxSlug)
	fitted = strings.Trim(fitted, "-")
	return fitted, nil
}

// truncateToMaxBytes returns the longest valid UTF-8 prefix of s with len ≤ maxBytes.
func truncateToMaxBytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	// Walk runes so we never split a multi-byte character.
	n := 0
	for i := 0; i < len(s); {
		_, size := utf8.DecodeRuneInString(s[i:])
		if n+size > maxBytes {
			return s[:n]
		}
		n += size
		i += size
	}
	return s[:n]
}
