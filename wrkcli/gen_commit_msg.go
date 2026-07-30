package wrkcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/commit_msg"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// genCommitMsgDisallowedFlags are wrk mode / create-flow flags that cannot
// appear with bare --gen-commit-msg (library-owned flags like --dry-run are allowed).
// Pipeline compose partners (--done, --merge-back, --sync, --tag-next, --push,
// --propagate-tags, --reinstall-local, --exec) are peeled before this list is
// consulted, so they never hit the bare exclusive path.
var genCommitMsgDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--status", "--repos", "--projects",
	"--projects-dep-graph",
	"--scan-git-repos", "--no-cache", "--include-worktrees",
	"--fetch", "--github", "--add", "--rm", "--where", "--cd", "--main",
	"--bring", "--no-dep", "--reinstall-local", "--tag-next",
	"--propagate-tags", "--sync",
	"-t", "--task", "--set-task",
	"--exec",
	"--web", "--port", "--dev",
	"--version",
	"--set-config", "--bash-integration",
	"--push", "--json",
	"--confirm-from-stdin", "--confirm", "--no-in-module-replace",
	"--no-cd", "--force-cd",
	"--new-window", "--no-new-window",
	"--new-terminal", "--reuse-terminal", "--smart-terminal", "--no-new-terminal",
	"--open-in-agent", "--no-open-in-agent",
	"--no-config",
	"--color", "-v", "--verbose",
}

// genCommitMsgComposePartners are flags that compose with --gen-commit-msg under
// the activeRoot pipeline model (peeled; not mutually exclusive).
var genCommitMsgComposePartners = []string{
	"--done", "--merge-back",
	"--sync", "--tag-next", "--push", "--propagate-tags",
	"--reinstall-local", "--exec",
}

// hasGenCommitComposePartner reports whether args include a pipeline partner
// that peels --gen-commit-msg into a multi-stage compose path.
func hasGenCommitComposePartner(args []string) bool {
	for _, p := range genCommitMsgComposePartners {
		if hasArg(args, p) {
			return true
		}
	}
	return false
}

// genCommitMsgValueFlags are library flags that take a value (separate arg or =form).
// --dry-run is intentionally not peeled: it is a shared wrk primary modifier.
var genCommitMsgValueFlags = map[string]struct{}{
	"--dir":                 {},
	"--model":               {},
	"--agent-runner":        {},
	"--agent-runner-binary": {},
}

// genCommitMsgBoolFlags are library bool flags peeled when composing with a primary.
var genCommitMsgBoolFlags = map[string]struct{}{
	"--commit":    {},
	"--no-verify": {},
	"--add-all":   {},
}

// peelGenCommitMsgForCompose removes --gen-commit-msg and its library-owned flags
// from args so lessflags can parse the remaining wrk primary/post flags.
// Returned genArgs do not include the --gen-commit-msg token itself.
func peelGenCommitMsgForCompose(args []string) (has bool, genArgs []string, rest []string) {
	rest = make([]string, 0, len(args))
	skipNext := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		arg := args[i]
		name := arg
		if j := strings.IndexByte(arg, '='); j >= 0 {
			name = arg[:j]
		}

		if name == "--gen-commit-msg" {
			has = true
			continue
		}
		if _, ok := genCommitMsgBoolFlags[name]; ok {
			genArgs = append(genArgs, arg)
			continue
		}
		if _, ok := genCommitMsgValueFlags[name]; ok {
			if strings.Contains(arg, "=") {
				genArgs = append(genArgs, arg)
				continue
			}
			genArgs = append(genArgs, arg)
			if i+1 < len(args) {
				genArgs = append(genArgs, args[i+1])
				skipNext = true
			}
			continue
		}
		rest = append(rest, arg)
	}
	return has, genArgs, rest
}

func genArgsHasFlag(genArgs []string, flag string) bool {
	for _, a := range genArgs {
		name := a
		if j := strings.IndexByte(a, '='); j >= 0 {
			name = a[:j]
		}
		if name == flag {
			return true
		}
	}
	return false
}

// runGenCommitMsgPreStage runs library gen-commit-msg on the source worktree
// before --done / --merge-back. Requires --commit; rejects composed --dir
// (wrk workDir wins). Forwards peeled library flags and --dry-run when set.
func runGenCommitMsgPreStage(workDir string, enabled bool, genArgs []string, dryRun bool, primaryFlag string) error {
	if !enabled {
		return nil
	}
	if !genArgsHasFlag(genArgs, "--commit") {
		return fmt.Errorf("wrk: --commit is required with --gen-commit-msg when used with %s", primaryFlag)
	}
	if genArgsHasFlag(genArgs, "--dir") {
		return fmt.Errorf("wrk: --dir is not valid with --gen-commit-msg when used with %s", primaryFlag)
	}

	return runGenCommitMsgStage(workDir, genArgs, dryRun)
}

// runGenCommitMsgStage runs library gen-commit-msg against workDir (activeRoot).
// Used as stage 1 of multi-stage compose (with or without done/merge-back).
// Does not require --commit (caller / library may still fail later).
func runGenCommitMsgStage(workDir string, genArgs []string, dryRun bool) error {
	if genArgsHasFlag(genArgs, "--dir") {
		return fmt.Errorf("wrk: --dir is not valid with --gen-commit-msg when used in multi-stage compose")
	}
	// Exclusive-branch: refuse --commit (including dry-run) when HEAD is shared.
	if genArgsHasFlag(genArgs, "--commit") {
		if err := refuseCommitIfBranchShared(workDir); err != nil {
			return err
		}
	}
	forwarded := make([]string, 0, len(genArgs)+3)
	forwarded = append(forwarded, genArgs...)
	// Pin to activeRoot workDir.
	forwarded = append(forwarded, "--dir", workDir)
	if dryRun {
		forwarded = append(forwarded, "--dry-run")
	}
	err := commit_msg.RunGenCommitMsg(forwarded)
	if err == nil {
		return nil
	}
	// Dry-run compose with empty index (common for dashboard recipe defaults):
	// library prints would: lines then errors "no staged". Continue so primary
	// dry-run plan can still run; real (non-dry) runs still fail.
	if dryRun && strings.Contains(err.Error(), "no staged") {
		fmt.Fprintf(os.Stderr, "would: skip gen-commit-msg (no staged changes)\n")
		return nil
	}
	return err
}

// withStashedStagedForDryPlan temporarily stashes only the index (staged
// changes) so MergeBack --rm dry-run can plan as if --gen-commit-msg --commit
// had already created a clean tree. Restores the staged set afterward (zero
// permanent mutation). Unstaged dirt is left in place so MergeBack still fails
// when a real commit would leave the worktree dirty.
func withStashedStagedForDryPlan(workDir string, fn func() error) error {
	if err := worktree.IsClean(workDir); err == nil {
		return fn()
	}
	// Stash staged only (git ≥2.35 --staged). Quiet message avoids noisy stdout.
	if err := gitRunDir(workDir, "stash", "push", "--staged", "-m", "wrk-gen-commit-msg-dry-run"); err != nil {
		// No staged changes to stash (or older git): run as-is; MergeBack may fail dirty.
		return fn()
	}
	fnErr := fn()
	// Always try to restore staged set for zero-mutation dry-run contract.
	popOut, popErr := gitCombinedRunDir(workDir, nil, "stash", "pop")
	if fnErr != nil {
		return fnErr
	}
	if popErr != nil {
		return fmt.Errorf("wrk: restore staged after gen-commit dry plan: %w\n%s", popErr, string(popOut))
	}
	return nil
}

// genCommitMsgHelpText mirrors agent-pro commit_msg help. Printed from wrk
// when -h/--help is requested so wrkcli.Capture (L2) does not hit the library
// less-gen path that calls os.Exit(0).
const genCommitMsgHelpText = `Usage: gen-commit-msg [options]

Generate a commit message for the currently staged changes using AI.
Logs are printed to stderr; the resulting commit message is printed to stdout.

Options:
  --dir DIR    Git directory to use (defaults to current directory)
  --model MODEL
              Model to use for generation
  --agent-runner RUNNER
              Agent runner to use (opencode|commandcode, default: opencode)
  --agent-runner-binary PATH
              Override the agent runner executable path
  --add-all    Stage all changes (git add -A) before generate; dry-run prints would: git add -A
  --commit     Run git commit with the generated message after printing it
  --no-verify  Skip git commit hooks (requires --commit)
  --dry-run    Pure plan: inspect staged set, print mock message; no agent, no unstage, no commit
  -h, --help   Show this help message
`

// runGenCommitMsg handles bare wrk --gen-commit-msg [...] (no primary).
// Remaining flags are forwarded to agent-pro commit_msg.RunGenCommitMsg.
func runGenCommitMsg(args []string, ctx *invocationContext) error {
	for _, arg := range args {
		name := arg
		if i := strings.IndexByte(arg, '='); i >= 0 {
			name = arg[:i]
		}
		for _, d := range genCommitMsgDisallowedFlags {
			if name == d {
				return fmt.Errorf("wrk: --gen-commit-msg is mutually exclusive with other modes")
			}
		}
	}

	forwarded := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--gen-commit-msg" {
			continue
		}
		// Top-level wrk yes/confirm flags are no-ops for bare gen-commit-msg
		// (library has no Y/n prompts). Strip so accidental -y from compose
		// recipes does not surface as "unrecognized flag".
		name := arg
		if i := strings.IndexByte(arg, '='); i >= 0 {
			name = arg[:i]
		}
		switch name {
		case "-y", "--yes", "--confirm", "--confirm-from-stdin":
			continue
		}
		forwarded = append(forwarded, arg)
	}

	ctx.skipEvent = true
	ctx.command = "gen-commit-msg"

	// Library RunGenCommitMsg uses less-gen Help without HelpNoExit, so -h/--help
	// would call os.Exit(0) and panic under testing.T / wrkcli.Capture. Handle
	// help here and return nil (exit 0) instead.
	for _, arg := range forwarded {
		if arg == "-h" || arg == "--help" {
			fmt.Print(genCommitMsgHelpText)
			if !strings.HasSuffix(genCommitMsgHelpText, "\n") {
				fmt.Println()
			}
			return nil
		}
	}

	// Library defaults to os.Getwd() when --dir is absent. Under Capture, pin
	// --dir to processCwd() (virtual Capture Dir) so InProcess leaves hit the
	// fixture repo rather than the suite package directory.
	hasDir := false
	dirVal := ""
	for i, arg := range forwarded {
		if arg == "--dir" {
			hasDir = true
			if i+1 < len(forwarded) {
				dirVal = forwarded[i+1]
			}
			break
		}
		if strings.HasPrefix(arg, "--dir=") {
			hasDir = true
			dirVal = strings.TrimPrefix(arg, "--dir=")
			break
		}
	}
	if !hasDir {
		wd, err := processCwd()
		if err != nil {
			return err
		}
		forwarded = append(forwarded, "--dir", wd)
		dirVal = wd
	}

	// Exclusive-branch: refuse --commit (including dry-run) when HEAD is shared.
	if genArgsHasFlag(forwarded, "--commit") {
		checkDir := dirVal
		if checkDir == "" {
			wd, err := processCwd()
			if err != nil {
				return err
			}
			checkDir = wd
		}
		if err := refuseCommitIfBranchShared(checkDir); err != nil {
			return err
		}
	}

	return commit_msg.RunGenCommitMsg(forwarded)
}
