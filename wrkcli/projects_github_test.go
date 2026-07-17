package wrkcli

import "testing"

func TestIsGitHubRemoteURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://github.com/owner/repo.git", true},
		{"https://github.com/owner/repo", true},
		{"http://github.com/owner/repo", true},
		{"https://GitHub.com/owner/repo.git", true},
		{"git@github.com:owner/repo.git", true},
		{"git@GitHub.com:owner/repo", true},
		{"ssh://git@github.com/owner/repo.git", true},
		{"ssh://git@github.com:22/owner/repo.git", true},
		{"  https://github.com/o/r  ", true},
		{"", false},
		{"https://gitlab.com/owner/repo", false},
		{"https://github.mycorp.com/owner/repo", false},
		{"git@gitlab.com:owner/repo.git", false},
		{"/path/to/local.git", false},
		{"file:///tmp/repo.git", false},
		{"https://notgithub.com/github.com/x", false},
		{"https://evil.com/github.com/owner/repo", false},
	}
	for _, tc := range cases {
		if got := isGitHubRemoteURL(tc.in); got != tc.want {
			t.Errorf("isGitHubRemoteURL(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
