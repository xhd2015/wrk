package wrkcli

import (
	"fmt"
	"strconv"
	"strings"
)

// dashboardStagePreview returns a cheap one-line preview and any captured git
// diagnostics for the Log panel. Soft-fails preview to "" on error; stderr lines
// are never discarded and never written to the real tty (caller feeds Logs to TUI).
func dashboardStagePreview(workDir, stageID string) (preview string, logs []string) {
	workDir = strings.TrimSpace(workDir)
	stageID = strings.TrimSpace(stageID)
	if workDir == "" || stageID == "" {
		return "", nil
	}
	switch stageID {
	case "add-changes", "gen-commit-msg", "commit":
		return previewPorcelainFiles(workDir)
	case "merge-back", "done":
		return previewAheadOfMain(workDir)
	case "push":
		return previewPushUpstream(workDir)
	case "sync", "tag-next", "reinstall-local":
		// Leave empty: sync needs a full worktree scan; tag-next must not create tags;
		// reinstall-local module scan is optional and not free enough for first paint.
		return "", nil
	default:
		return "", nil
	}
}

// captureGit runs git with stdout+stderr captured (no tty inherit). Appends stderr
// lines to *logs as normal log content.
func captureGit(workDir string, logs *[]string, args ...string) (stdout string, err error) {
	out, stderr, err := gitOutputDirCapture(workDir, args...)
	if logs != nil {
		*logs = append(*logs, splitCapturedLogLines(stderr)...)
	}
	return out, err
}

// previewPorcelainFiles summarizes `git status --porcelain` as "N files" or "clean".
func previewPorcelainFiles(workDir string) (string, []string) {
	var logs []string
	out, err := captureGit(workDir, &logs, "status", "--porcelain")
	if err != nil {
		return "", logs
	}
	n := countNonEmptyLines(out)
	return formatFileCountPreview(n), logs
}

// formatFileCountPreview is a pure helper for porcelain summary text.
func formatFileCountPreview(n int) string {
	if n <= 0 {
		return "clean"
	}
	if n == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", n)
}

func countNonEmptyLines(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		n++
	}
	return n
}

// previewAheadOfMain reports commits ahead of main checkout branch, or "on main".
func previewAheadOfMain(workDir string) (string, []string) {
	var logs []string
	if dashboardIsMainCheckout(workDir) {
		return "on main", nil
	}
	mainRepo, err := resolveMainRepoForWorkDir(workDir)
	if err != nil || mainRepo == "" {
		return "", logs
	}
	mainBranch, err := captureGit(mainRepo, &logs, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", logs
	}
	mainBranch = strings.TrimSpace(mainBranch)
	if mainBranch == "" || mainBranch == "HEAD" {
		return "", logs
	}
	// Shared object DB with linked worktrees; count commits on HEAD not in main branch.
	out, err := captureGit(workDir, &logs, "rev-list", "--count", mainBranch+"..HEAD")
	if err != nil {
		return "", logs
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return "", logs
	}
	return formatAheadPreview(n), logs
}

// formatAheadPreview is a pure helper for merge-back/done preview text.
func formatAheadPreview(n int) string {
	if n <= 0 {
		return "ahead 0"
	}
	return fmt.Sprintf("ahead %d", n)
}

// previewPushUpstream reports cheap @{u} ahead/behind without fetching.
// Missing upstream: empty preview; git diagnostics go to logs (not the real tty).
func previewPushUpstream(workDir string) (string, []string) {
	var logs []string
	// Probe upstream first; capture any "fatal: no upstream…" into logs.
	u, err := captureGit(workDir, &logs, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil || strings.TrimSpace(u) == "" {
		return "", logs
	}
	// left = commits on upstream not in HEAD (behind), right = commits on HEAD not in upstream (ahead).
	out, err := captureGit(workDir, &logs, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if err != nil {
		return "", logs
	}
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) != 2 {
		return "", logs
	}
	behind, err1 := strconv.Atoi(fields[0])
	ahead, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil {
		return "", logs
	}
	return formatPushPreview(ahead, behind), logs
}

// formatPushPreview is a pure helper for push stage preview text.
func formatPushPreview(ahead, behind int) string {
	switch {
	case ahead == 0 && behind == 0:
		return "up to date"
	case behind == 0:
		return fmt.Sprintf("ahead %d", ahead)
	case ahead == 0:
		return fmt.Sprintf("behind %d", behind)
	default:
		return fmt.Sprintf("ahead %d, behind %d", ahead, behind)
	}
}
