package unwind

import (
	"fmt"
	"os/exec"
	"strings"

	xgocmd "github.com/xhd2015/xgo/support/cmd"
)

func gitRunDir(repoPath string, args ...string) error {
	h := currentHost()
	if h.GitRun != nil {
		return h.GitRun(repoPath, args...)
	}
	return xgocmd.Dir(repoPath).Run("git", args...)
}

func gitOutputDir(repoPath string, args ...string) (string, error) {
	h := currentHost()
	if h.GitOutput != nil {
		return h.GitOutput(repoPath, args...)
	}
	return xgocmd.Dir(repoPath).Output("git", args...)
}

func shortHEAD(repo string) (string, error) {
	h := currentHost()
	if h.ShortHEAD != nil {
		return h.ShortHEAD(repo)
	}
	out, err := gitOutputDir(repo, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --short=7 HEAD: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func goModTidy(dir string) error {
	h := currentHost()
	if h.GoModTidy != nil {
		return h.GoModTidy(dir)
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("failed to execute go mod tidy: %w\n%s", err, msg)
		}
		return fmt.Errorf("failed to execute go mod tidy: %w", err)
	}
	return nil
}
