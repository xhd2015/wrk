package wrkcli

import (
	"bytes"
	"io"
	"net/url"
	"strings"
)

// isGitHubRemoteURL reports whether remoteURL points at github.com
// (case-insensitive host). Supports https, http, scp-like ssh
// (git@github.com:owner/repo.git), and ssh:// forms.
func isGitHubRemoteURL(remoteURL string) bool {
	s := strings.TrimSpace(remoteURL)
	if s == "" {
		return false
	}
	// scp-like: git@github.com:owner/repo.git (optional user other than git)
	if !strings.Contains(s, "://") {
		if at := strings.Index(s, "@"); at >= 0 {
			rest := s[at+1:]
			if colon := strings.Index(rest, ":"); colon >= 0 {
				host := rest[:colon]
				return strings.EqualFold(host, "github.com")
			}
		}
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return false
	}
	host := u.Hostname()
	return strings.EqualFold(host, "github.com")
}

// projectHasGitHubOrigin reports whether mainRepo's origin remote URL is github.com.
// Missing origin, git errors, or non-github hosts return false (quiet omit).
// Git stderr is discarded so non-matches do not pollute the CLI.
func projectHasGitHubOrigin(mainRepo string) bool {
	cmd := gitCommandDir(mainRepo, "remote", "get-url", "origin")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return false
	}
	return isGitHubRemoteURL(strings.TrimSpace(stdout.String()))
}

// filterGitHubProjectPaths keeps only paths whose origin is github.com.
// Order of paths is preserved.
func filterGitHubProjectPaths(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if projectHasGitHubOrigin(p) {
			out = append(out, p)
		}
	}
	return out
}
