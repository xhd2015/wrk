package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"golang.org/x/term"
)

// ghPR is the subset of gh pr list --json fields used by --pr.
type ghPR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// runPR implements bare wrk --pr --title T --comment C from a linked worktree.
// It ensures the head branch exists on origin (push only when missing), then
// creates or attaches a GitHub PR via gh and always adds an additive comment.
func runPR(workDir, title, comment string, colorFlag bool) error {
	title = strings.TrimSpace(title)
	comment = strings.TrimSpace(comment)
	if title == "" {
		return fmt.Errorf("wrk: --title must not be empty")
	}
	if comment == "" {
		return fmt.Errorf("wrk: --comment must not be empty")
	}

	checkoutRoot, err := requireLinkedWorktree(workDir, "--pr")
	if err != nil {
		return err
	}

	// origin must be github.com (fetch URL; push may use pushurl to a bare).
	originURL, err := gitOutputDir(checkoutRoot, "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("wrk: --pr requires an origin remote (github.com): %w", err)
	}
	originURL = strings.TrimSpace(originURL)
	if !isGitHubRemoteURL(originURL) {
		return fmt.Errorf("wrk: --pr requires a github.com origin remote (got %q)", originURL)
	}

	headBranch, err := worktree.ReadBranch(checkoutRoot)
	if err != nil {
		return fmt.Errorf("wrk: resolve current branch for --pr: %w", err)
	}
	headBranch = strings.TrimSpace(headBranch)
	if headBranch == "" || headBranch == "HEAD" {
		return fmt.Errorf("wrk: --pr cannot run on a detached HEAD")
	}

	mainRepo, err := worktree.ResolveMainRepo(checkoutRoot)
	if err != nil {
		return err
	}
	baseBranch, err := worktree.ReadBranch(mainRepo)
	if err != nil {
		return fmt.Errorf("wrk: resolve main branch for --pr: %w", err)
	}
	baseBranch = strings.TrimSpace(baseBranch)
	if baseBranch == "" || baseBranch == "HEAD" {
		return fmt.Errorf("wrk: --pr requires main repository to be on a named branch (not detached HEAD)")
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("Error: wrk: --pr requires the GitHub CLI (gh); install from https://cli.github.com/")
	}

	if err := ensureOriginHeadForPR(checkoutRoot, headBranch); err != nil {
		return err
	}

	existing, err := listOpenPRForHead(checkoutRoot, headBranch)
	if err != nil {
		return err
	}

	colorEnabled := stdoutColorEnabled(colorFlag)
	var prURL string
	var prNumber int

	if existing != nil {
		// Title ignored when PR already exists; still add comment.
		warn := fmt.Sprintf("warning: title ignored (PR already exists); existing title: %s", existing.Title)
		fmt.Fprintln(os.Stderr, FormatStderrWarning(warn))
		prURL = existing.URL
		prNumber = existing.Number
	} else {
		url, number, err := createPR(checkoutRoot, title, baseBranch, headBranch)
		if err != nil {
			return err
		}
		prURL = url
		prNumber = number
		fmt.Println(prSuccessToken("PR created", colorEnabled))
		titleLine := prSuccessToken("title set", colorEnabled) + ": " + title
		fmt.Println(titleLine)
	}

	if err := commentPR(checkoutRoot, prNumber, comment); err != nil {
		return err
	}
	fmt.Println(prSuccessToken("comment added", colorEnabled))
	fmt.Println(prURL)
	return nil
}

// ensureOriginHeadForPR pushes HEAD to origin/<branch> only when the remote
// head ref is missing. When the remote already has the branch, no push runs
// even if local is ahead.
//
// Existence is checked with ls-remote against origin's push URL when set
// (remote.origin.pushurl), else the normal origin remote. This matches hermetic
// fixtures that use a github.com fetch URL for host checks and a local bare
// pushurl for real push/ls-remote, and avoids contacting github.com for the
// ensure-head probe.
func ensureOriginHeadForPR(repo, branch string) error {
	exists, err := originHeadExists(repo, branch)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if combined, pushErr := gitCombinedRunDir(repo, nil, "push", "origin", "HEAD:refs/heads/"+branch); pushErr != nil {
		msg := strings.TrimSpace(string(combined))
		if msg != "" {
			return fmt.Errorf("wrk: git push origin %s failed: %s", branch, msg)
		}
		return fmt.Errorf("wrk: git push origin %s failed: %w", branch, pushErr)
	}
	fmt.Printf("pushed %s → origin/%s\n", branch, branch)
	return nil
}

// originHeadExists reports whether origin already has refs/heads/<branch>.
func originHeadExists(repo, branch string) (bool, error) {
	// Prefer pushurl when configured so github.com-shaped fetch URLs with a
	// local bare pushurl work offline (tests and redirect remotes).
	remote := "origin"
	if pushURL, err := gitOutputDir(repo, "config", "--get", "remote.origin.pushurl"); err == nil {
		if u := strings.TrimSpace(pushURL); u != "" {
			remote = u
		}
	}
	// Capture stderr so a failed probe does not leak git "fatal:" onto the CLI.
	out, _, err := gitOutputDirCapture(repo, "ls-remote", "--heads", remote, branch)
	if err != nil {
		out2, _, err2 := gitOutputDirCapture(repo, "ls-remote", "--heads", remote, "refs/heads/"+branch)
		if err2 != nil {
			return false, fmt.Errorf("wrk: git ls-remote origin %s: %w", branch, err)
		}
		out = out2
	}
	return strings.TrimSpace(out) != "", nil
}

// runGh runs gh with args in repo, capturing stdout/stderr. On success stdout
// is returned trimmed of a single trailing newline. Always best-effort flushes
// a FAKE_GH_LOG record separator after the child exits so multi-call argv logs
// stay parseable when a fake gh logger fails to write "---" (macOS /bin/sh
// builtin printf treats printf '---\n' as invalid options).
func runGh(repo string, args ...string) (stdout string, err error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = repo
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	flushFakeGhLogSeparator()
	if runErr != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = runErr.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return strings.TrimSpace(outBuf.String()), nil
}

// flushFakeGhLogSeparator appends "---\n" to FAKE_GH_LOG when set. No-op when
// unset or unreadable. Harmless if the child already wrote a separator.
func flushFakeGhLogSeparator() {
	logPath := os.Getenv("FAKE_GH_LOG")
	if logPath == "" {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return
	}
	_, _ = f.WriteString("---\n")
	_ = f.Close()
}

// listOpenPRForHead returns the first open PR for headBranch, or nil if none.
func listOpenPRForHead(repo, headBranch string) (*ghPR, error) {
	raw, err := runGh(repo, "pr", "list",
		"--head", headBranch,
		"--state", "open",
		"--json", "number,title,url",
	)
	if err != nil {
		return nil, fmt.Errorf("wrk: gh pr list failed: %s", err.Error())
	}
	if raw == "" {
		raw = "[]"
	}
	var list []ghPR
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil, fmt.Errorf("wrk: parse gh pr list JSON: %w", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	pr := list[0]
	return &pr, nil
}

// createPR runs gh pr create and returns the PR URL and number (parsed from URL
// when gh only prints the link).
func createPR(repo, title, baseBranch, headBranch string) (url string, number int, err error) {
	// Non-interactive gh requires --body (or --fill). Use empty body; wrk always
	// posts the user comment via `gh pr comment` separately.
	out, runErr := runGh(repo, "pr", "create",
		"--title", title,
		"--body", "",
		"--base", baseBranch,
		"--head", headBranch,
	)
	if runErr != nil {
		return "", 0, fmt.Errorf("wrk: gh pr create failed: %s", runErr.Error())
	}
	url = out
	// Prefer the last non-empty line (gh may print notices before the URL).
	if lines := strings.Split(url, "\n"); len(lines) > 0 {
		for i := len(lines) - 1; i >= 0; i-- {
			if s := strings.TrimSpace(lines[i]); s != "" {
				url = s
				break
			}
		}
	}
	if url == "" {
		return "", 0, fmt.Errorf("wrk: gh pr create returned empty URL")
	}
	if n, ok := parsePRNumberFromURL(url); ok {
		number = n
	}
	if number == 0 {
		return "", 0, fmt.Errorf("wrk: could not determine PR number after create (url=%s)", url)
	}
	return url, number, nil
}

// commentPR adds an additive comment via gh.
func commentPR(repo string, number int, body string) error {
	if _, err := runGh(repo, "pr", "comment", fmt.Sprintf("%d", number), "--body", body); err != nil {
		return fmt.Errorf("wrk: gh pr comment failed: %s", err.Error())
	}
	return nil
}

func parsePRNumberFromURL(url string) (int, bool) {
	// .../pull/42 or .../pull/42/
	const marker = "/pull/"
	i := strings.LastIndex(url, marker)
	if i < 0 {
		return 0, false
	}
	rest := url[i+len(marker):]
	rest = strings.Trim(rest, "/")
	// stop at extra path/query
	for j, c := range rest {
		if c < '0' || c > '9' {
			rest = rest[:j]
			break
		}
	}
	if rest == "" {
		return 0, false
	}
	n := 0
	for _, c := range rest {
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return 0, false
	}
	return n, true
}

func prSuccessToken(token string, colorEnabled bool) string {
	if colorEnabled {
		return colorize(token, ansiGreen)
	}
	return token
}

// stdoutColorEnabled mirrors stderr color policy for success tokens on stdout.
func stdoutColorEnabled(force bool) bool {
	if force {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}


