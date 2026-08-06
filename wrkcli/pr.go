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

// runPRShow implements bare wrk --pr (no --title/--comment): list open PR for
// the linked worktree head and print its URL only, or empty stdout if none.
// No ensure-push, create, or comment.
func runPRShow(workDir string) error {
	checkoutRoot, headBranch, err := prSharedGates(workDir)
	if err != nil {
		return err
	}

	existing, err := listOpenPRForHead(checkoutRoot, headBranch)
	if err != nil {
		return err
	}
	if existing != nil {
		fmt.Println(existing.URL)
	}
	return nil
}

// ghPRView is the subset of gh pr view --json fields used by --pr --status.
type ghPRView struct {
	Number            int           `json:"number"`
	Title             string        `json:"title"`
	URL               string        `json:"url"`
	State             string        `json:"state"`
	IsDraft           bool          `json:"isDraft"`
	ReviewDecision    string        `json:"reviewDecision"`
	StatusCheckRollup []ghCheckNode `json:"statusCheckRollup"`
}

// ghCheckNode is one statusCheckRollup entry (CheckRun or StatusContext).
type ghCheckNode struct {
	Typename   string `json:"__typename"`
	Name       string `json:"name"`
	Status     string `json:"status"`     // CheckRun: QUEUED, IN_PROGRESS, COMPLETED, …
	Conclusion string `json:"conclusion"` // CheckRun: SUCCESS, FAILURE, NEUTRAL, SKIPPED, …
	State      string `json:"state"`      // StatusContext / rollup: SUCCESS, FAILURE, PENDING, …
}

// runPRStatus implements wrk --pr --status: list open PR for the linked
// worktree head, fetch view JSON, and print URL + State/Title/Checks/Reviews.
// Read-only: no push, create, or comment. Exit 0 even when Checks=failure.
// No open PR: exit 0, empty stdout, stderr warning.
func runPRStatus(workDir string) error {
	checkoutRoot, headBranch, err := prSharedGates(workDir)
	if err != nil {
		return err
	}

	existing, err := listOpenPRForHead(checkoutRoot, headBranch)
	if err != nil {
		return err
	}
	if existing == nil {
		warn := fmt.Sprintf("warning: no open pull request for branch %s", headBranch)
		fmt.Fprintln(os.Stderr, FormatStderrWarning(warn))
		return nil
	}

	raw, err := runGh(checkoutRoot, "pr", "view", fmt.Sprintf("%d", existing.Number),
		"--json", "number,title,url,state,isDraft,reviewDecision,statusCheckRollup",
	)
	if err != nil {
		return fmt.Errorf("wrk: gh pr view failed: %s", err.Error())
	}
	if raw == "" {
		raw = "{}"
	}

	// Decoder (not Unmarshal): some hermetic fake-gh scripts expand
	// ${FAKE_GH_VIEW_JSON:-{}} which appends a trailing "}" when the env is set.
	// Accept the first JSON value and ignore trailing junk.
	var view ghPRView
	dec := json.NewDecoder(strings.NewReader(raw))
	if err := dec.Decode(&view); err != nil {
		return fmt.Errorf("wrk: parse gh pr view JSON: %w", err)
	}

	state := formatPRState(view.IsDraft, view.State)
	checks := formatChecksRollup(view.StatusCheckRollup)
	reviews := formatReviewDecision(view.ReviewDecision)
	url := view.URL
	if url == "" {
		url = existing.URL
	}
	title := view.Title
	if title == "" {
		title = existing.Title
	}

	// Exact field labels + column spacing (State:/Title:/Checks:/Reviews: → col 11).
	fmt.Printf("%s\nState:     %s\nTitle:     %s\nChecks:    %s\nReviews:   %s\n",
		url, state, title, checks, reviews)
	return nil
}

// formatPRState maps isDraft + gh state → open|draft|… (lowercased).
func formatPRState(isDraft bool, state string) string {
	if isDraft {
		return "draft"
	}
	s := strings.ToLower(strings.TrimSpace(state))
	if s == "" {
		return "open"
	}
	return s
}

// formatReviewDecision maps gh reviewDecision → human lower case.
func formatReviewDecision(d string) string {
	switch strings.ToUpper(strings.TrimSpace(d)) {
	case "":
		return "none"
	case "APPROVED":
		return "approved"
	case "CHANGES_REQUESTED":
		return "changes requested"
	case "REVIEW_REQUIRED":
		return "review required"
	default:
		// Fallback: snake_case → spaces, lower.
		return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(d), "_", " "))
	}
}

// formatChecksRollup maps statusCheckRollup → success|failure|pending|none|mixed.
//
// Rules:
//   - empty / missing → none
//   - any FAILURE/ERROR/TIMED_OUT/CANCELLED/ACTION_REQUIRED → failure
//   - any PENDING/EXPECTED/IN_PROGRESS/QUEUED/WAITING/REQUESTED without failure → pending
//   - all SUCCESS (and NEUTRAL/SKIPPED) → success
//   - otherwise mixed (e.g. success + neutral-only edge cases already covered)
func formatChecksRollup(nodes []ghCheckNode) string {
	if len(nodes) == 0 {
		return "none"
	}

	var hasSuccess, hasPending, hasFailure, hasOther bool
	for _, n := range nodes {
		switch classifyCheckNode(n) {
		case "failure":
			hasFailure = true
		case "pending":
			hasPending = true
		case "success":
			hasSuccess = true
		default:
			hasOther = true
		}
	}
	// Failure dominates.
	if hasFailure {
		return "failure"
	}
	if hasPending {
		// success+pending without failure → pending (in-progress overall).
		return "pending"
	}
	if hasSuccess && !hasOther {
		return "success"
	}
	if hasSuccess && hasOther {
		return "mixed"
	}
	if hasOther {
		return "mixed"
	}
	return "none"
}

// classifyCheckNode returns success|failure|pending|other for one rollup node.
func classifyCheckNode(n ghCheckNode) string {
	// Prefer conclusion (CheckRun completed), then state (StatusContext / rollup state),
	// then status (CheckRun lifecycle).
	conclusion := strings.ToUpper(strings.TrimSpace(n.Conclusion))
	state := strings.ToUpper(strings.TrimSpace(n.State))
	status := strings.ToUpper(strings.TrimSpace(n.Status))

	// Explicit failure outcomes.
	for _, v := range []string{conclusion, state} {
		switch v {
		case "FAILURE", "ERROR", "TIMED_OUT", "CANCELLED", "CANCELED", "ACTION_REQUIRED", "STARTUP_FAILURE":
			return "failure"
		}
	}

	// In-progress / expected / pending lifecycle.
	for _, v := range []string{status, state, conclusion} {
		switch v {
		case "PENDING", "EXPECTED", "IN_PROGRESS", "QUEUED", "WAITING", "REQUESTED", "WAITING_TO_MERGE":
			return "pending"
		}
	}

	// Success-class (including neutral/skipped as non-blocking green).
	for _, v := range []string{conclusion, state} {
		switch v {
		case "SUCCESS", "NEUTRAL", "SKIPPED":
			return "success"
		}
	}
	if status == "COMPLETED" && (conclusion == "" || conclusion == "SUCCESS" || conclusion == "NEUTRAL" || conclusion == "SKIPPED") {
		if conclusion == "" && state != "" {
			// COMPLETED without conclusion but with state already handled above.
			return "other"
		}
		return "success"
	}

	if conclusion == "" && state == "" && status == "" {
		return "other"
	}
	return "other"
}

// runPRComment implements wrk --pr --comment C (no --title): attach an additive
// comment to the existing open PR for the linked worktree head. Never creates a
// PR, never ensure-pushes, and never warns about title (title was not passed).
func runPRComment(workDir, comment string, colorFlag bool) error {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("wrk: --comment must not be empty")
	}

	checkoutRoot, headBranch, err := prSharedGates(workDir)
	if err != nil {
		return err
	}

	existing, err := listOpenPRForHead(checkoutRoot, headBranch)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("wrk: no open pull request for branch %q", headBranch)
	}

	if err := commentPR(checkoutRoot, existing.Number, comment); err != nil {
		return err
	}
	fmt.Println(prSuccessToken("comment added", stdoutColorEnabled(colorFlag)))
	fmt.Println(existing.URL)
	return nil
}

// runPRPushExisting implements wrk --pr --push [--comment C] without --title:
// require an open PR for the linked worktree head, full tip push (runPushMain
// semantics), optional additive comment, then print the PR URL.
//
// Open-PR list runs BEFORE any push so a missing PR leaves origin tip unchanged.
// Never creates a PR and never emits a title-ignored warning.
// Stdout stages (blank line between push and PR stage):
//
//	pushed <branch> → origin/<branch>
//
//	[comment added]
//	https://...
func runPRPushExisting(workDir, comment string, dryRun, force bool, colorFlag bool) error {
	comment = strings.TrimSpace(comment)

	checkoutRoot, headBranch, err := prSharedGates(workDir)
	if err != nil {
		return err
	}

	existing, err := listOpenPRForHead(checkoutRoot, headBranch)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("wrk: no open pull request for branch %q", headBranch)
	}

	// Full tip push (same as bare --push / compose push stage).
	if err := runPushMain(checkoutRoot, dryRun, force, nil); err != nil {
		return err
	}

	// Blank line between push stage and PR stage (joinStdoutBlocks style).
	fmt.Println()

	if comment != "" {
		if err := commentPR(checkoutRoot, existing.Number, comment); err != nil {
			return err
		}
		fmt.Println(prSuccessToken("comment added", stdoutColorEnabled(colorFlag)))
	}
	fmt.Println(existing.URL)
	return nil
}

// prSharedGates enforces the common --pr preconditions: linked worktree,
// github.com origin, named head branch, and gh on PATH. Returns checkout root
// and head branch name.
func prSharedGates(workDir string) (checkoutRoot, headBranch string, err error) {
	checkoutRoot, err = requireLinkedWorktree(workDir, "--pr")
	if err != nil {
		return "", "", err
	}

	// origin must be github.com (fetch URL; push may use pushurl to a bare).
	originURL, err := gitOutputDir(checkoutRoot, "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("wrk: --pr requires an origin remote (github.com): %w", err)
	}
	originURL = strings.TrimSpace(originURL)
	if !isGitHubRemoteURL(originURL) {
		return "", "", fmt.Errorf("wrk: --pr requires a github.com origin remote (got %q)", originURL)
	}

	headBranch, err = worktree.ReadBranch(checkoutRoot)
	if err != nil {
		return "", "", fmt.Errorf("wrk: resolve current branch for --pr: %w", err)
	}
	headBranch = strings.TrimSpace(headBranch)
	if headBranch == "" || headBranch == "HEAD" {
		return "", "", fmt.Errorf("wrk: --pr cannot run on a detached HEAD")
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return "", "", fmt.Errorf("Error: wrk: --pr requires the GitHub CLI (gh); install from https://cli.github.com/")
	}
	return checkoutRoot, headBranch, nil
}

// runPR implements wrk --pr --title T --comment C from a linked worktree.
// It ensures the head branch exists on origin (push only when missing), then
// creates or attaches a GitHub PR via gh:
//   - new PR: title + comment as initial body (no separate issue comment)
//   - existing PR: title ignored; comment as additive gh pr comment
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

	if existing != nil {
		// Title ignored when PR already exists; comment becomes an issue comment.
		warn := fmt.Sprintf("warning: title ignored (PR already exists); existing title: %s", existing.Title)
		fmt.Fprintln(os.Stderr, FormatStderrWarning(warn))
		prURL = existing.URL
		if err := commentPR(checkoutRoot, existing.Number, comment); err != nil {
			return err
		}
		fmt.Println(prSuccessToken("comment added", colorEnabled))
	} else {
		// New PR: --comment is the initial body (not a separate issue comment).
		url, _, err := createPR(checkoutRoot, title, comment, baseBranch, headBranch)
		if err != nil {
			return err
		}
		prURL = url
		fmt.Println(prSuccessToken("PR created", colorEnabled))
		fmt.Println(prSuccessToken("title set", colorEnabled) + ": " + title)
		fmt.Println(prSuccessToken("body set", colorEnabled))
	}

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
// when gh only prints the link). body is the initial PR description (--comment).
func createPR(repo, title, body, baseBranch, headBranch string) (url string, number int, err error) {
	// Non-interactive gh requires --title and --body (or --fill).
	out, runErr := runGh(repo, "pr", "create",
		"--title", title,
		"--body", body,
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
