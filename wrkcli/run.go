package wrkcli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/wrk/workops"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/term"
)

// RunOpts configures an in-process wrk invocation. CLI main uses Run (zero opts).
type RunOpts struct {
	// Args are CLI args after the program name (same as Run).
	Args []string

	// ScanTestPauseAfterFirstPrint, when > 0, pauses after the first path
	// printed by --scan-git-repos so mid-scan interrupt tests can cancel
	// while the walk is still open. Production leaves this zero.
	//
	// Supporting tests:
	//   - cmd/wrk/tests/scan-git-repos/interrupt/sigint-after-first-path
	//   - (other leaves under scan-git-repos/interrupt/ if added)
	ScanTestPauseAfterFirstPrint time.Duration
}

// Run executes wrk logic with args. The first positional argument,
// if present, is the source directory for all modes.
func Run(args []string) error {
	return RunWithOpts(RunOpts{Args: args})
}

// RunWithOpts is like Run but accepts optional in-process test/runtime opts
// (passed down the call stack; no env or package globals).
func RunWithOpts(opts RunOpts) error {
	origWd, err := processCwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	ctx := newInvocationContext(origWd, opts.Args)
	var runErr error
	defer func() {
		exitCode := 0
		if runErr != nil {
			var ece ExitCodeError
			if errors.As(runErr, &ece) {
				exitCode = ece.Code
			} else {
				exitCode = 1
			}
		}
		ctx.finish(exitCode)
	}()
	runErr = run(origWd, opts.Args, ctx, opts)
	return runErr
}

// rejectWhereEqualsForm fails on --where=value. lessflags Bool accepts equals form
// and only treats ""/"true" as true, so --where=spl would silently clear the flag
// and fall through to dashboard. Product contract: Bool+positional only.
func rejectWhereEqualsForm(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--where=") {
			return fmt.Errorf("wrk: --where does not take equals form (=value)")
		}
	}
	return nil
}

// rejectPrEqualsForm fails on --pr=value. --pr is Bool; PR URL is a separate
// positional when composed with --where (or bare --pr has no value). Equals form
// would otherwise silently clear the flag (non-"true" values → false).
func rejectPrEqualsForm(args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--pr=") {
			return fmt.Errorf("wrk: --pr does not take equals form (=value); pass a full GitHub pull request URL as a positional argument")
		}
	}
	return nil
}

func run(origWd string, args []string, ctx *invocationContext, opts RunOpts) error {
	if len(args) > 0 && args[0] == "skill" {
		wrkHome, err := resolveWrkHome()
		if err != nil {
			return err
		}
		ctx.wrkHome = wrkHome
		ctx.workDir = origWd
		ctx.command = "skill"
		ctx.eventArgs = args[1:]
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		if err := ctx.autoRecord(); err != nil {
			return err
		}
		return runSkill(origWd, args[1:], wrkHome)
	}
	if hasArg(args, "--bash-integration") {
		ctx.skipEvent = true
		return runBashIntegration(args)
	}
	if hasArg(args, "--set-config") {
		if hasArg(args, "--no-config") {
			return fmt.Errorf("wrk: --no-config is mutually exclusive with --set-config")
		}
		return runSetConfig(origWd, args, ctx)
	}
	if hasArg(args, "--version") {
		if len(args) == 1 && args[0] == "--version" {
			fmt.Println(Version())
			ctx.skipEvent = true
			return nil
		}
		return fmt.Errorf("wrk: --version is mutually exclusive with other modes")
	}
	// Bare --gen-commit-msg (no pipeline partner): exclusive early path.
	// With --done / --merge-back / other pipeline stages: peel library flags and
	// run as stage 1 of activeRoot compose later.
	// Message sources are XOR: AI (--gen-commit-msg) vs manual (-m/--message).
	if hasArg(args, "--gen-commit-msg") && hasMessageFlag(args) {
		return fmt.Errorf("wrk: --message is mutually exclusive with --gen-commit-msg")
	}
	// --main is never valid with --gen-commit-msg (named reject before bare gen path).
	if hasArg(args, "--gen-commit-msg") && hasArg(args, "--main") {
		return fmt.Errorf("wrk: --main is not valid with --gen-commit-msg")
	}
	var genCommitMsg bool
	var genCommitArgs []string
	parseArgs := args
	if hasArg(args, "--gen-commit-msg") {
		if !hasGenCommitComposePartner(args) {
			return runGenCommitMsg(args, ctx)
		}
		genCommitMsg, genCommitArgs, parseArgs = peelGenCommitMsgForCompose(args)
	}

	if err := rejectWhereEqualsForm(parseArgs); err != nil {
		return err
	}
	if err := rejectPrEqualsForm(parseArgs); err != nil {
		return err
	}

	var done bool
	var mergeBack bool
	var list bool
	var status bool
	var repos bool
	var projects bool
	var projectsDepGraph bool
	var colorFlag bool
	var noColorFlag bool
	var fetchFlag bool
	var githubFlag bool
	var verbose bool
	var addPath *string
	var removePath *string
	var confirmFromStdin bool
	var forceConfirm bool
	var assumeYes bool
	var noInModuleReplace bool
	var bringPaths []string
	var noDep bool
	var reinstallLocal bool
	var tagNext bool
	var propagateTags bool
	var syncFlag bool
	var pushFlag bool
	var forcePush bool // -f/--force: modifier of --push only (force-with-lease)
	var prFlag bool
	var prTitle *string
	var prComment *string
	var jsonFlag bool
	var dryRun bool
	var taskDesc *string
	var setTaskDesc *string
	var where bool
	var noCd bool
	var forceCd bool
	var cd bool
	var mainFlag bool
	var unwind bool
	var showGraph bool
	var execArgs []string
	// Manual commit message path: --commit -m/--message (wrk-owned; not AI gen).
	// When --gen-commit-msg peels --commit/--no-verify/--add-all, those stay in genArgs.
	var commitFlag bool
	var commitMessage *string
	var noVerify bool
	var addAll bool
	// Create UX one-shot flags.
	var newFlag bool // --new: explicit create entry (former bare create)
	var newWindow bool
	var noNewWindow bool
	var newTerminal bool
	var reuseTerminal bool
	var smartTerminal bool
	var noNewTerminal bool
	var openInAgent bool
	var noOpenInAgent bool
	var noConfig bool
	var webFlag bool
	var webDev bool
	// *int target: nil = --port absent; non-nil = present (0 allowed → auto later).
	var portFlag *int
	var scanGitRepos bool
	var noCache bool
	var includeWorktrees bool
	// *string targets: nil = flag absent; non-nil empty = present but empty.
	// Cut("--exec") must be registered so tokens after --exec are never re-parsed as flags.
	// Register --new before --new-window / --new-terminal so exact long names stay unambiguous.
	remaining, err := lessflags.Bool("--done", &done).
		Bool("--merge-back", &mergeBack).
		Bool("-l,--list", &list).
		Bool("--status", &status).
		Bool("--repos", &repos).
		Bool("--projects", &projects).
		Bool("--projects-dep-graph", &projectsDepGraph).
		Bool("--fetch", &fetchFlag).
		Bool("--github", &githubFlag).
		Bool("-v,--verbose", &verbose).
		Bool("--color", &colorFlag).
		Bool("--no-color", &noColorFlag).
		Bool("--web", &webFlag).
		Bool("--dev", &webDev).
		Int("--port", &portFlag).
		Bool("--scan-git-repos", &scanGitRepos).
		Bool("--no-cache", &noCache).
		Bool("--include-worktrees", &includeWorktrees).
		String("--add", &addPath).
		String("--rm", &removePath).
		Bool("--confirm-from-stdin", &confirmFromStdin).
		Bool("--confirm", &forceConfirm).
		Bool("-y,--yes", &assumeYes).
		Bool("--no-in-module-replace", &noInModuleReplace).
		Bool("--no-cd", &noCd).
		Bool("--force-cd", &forceCd).
		Bool("--cd", &cd).
		Bool("--main", &mainFlag).
		Bool("--unwind", &unwind).
		Bool("--show-graph", &showGraph).
		Bool("--reinstall-local", &reinstallLocal).
		Bool("--tag-next", &tagNext).
		Bool("--propagate-tags", &propagateTags).
		Bool("--sync", &syncFlag).
		Bool("--push", &pushFlag).
		Bool("-f,--force", &forcePush).
		Bool("--pr", &prFlag).
		String("--title", &prTitle).
		String("--comment", &prComment).
		Bool("--json", &jsonFlag).
		Bool("--dry-run", &dryRun).
		Bool("--commit", &commitFlag).
		String("-m,--message", &commitMessage).
		Bool("--no-verify", &noVerify).
		Bool("--add-all", &addAll).
		Bool("--new", &newFlag).
		Bool("--new-window", &newWindow).
		Bool("--no-new-window", &noNewWindow).
		Bool("--new-terminal", &newTerminal).
		Bool("--reuse-terminal", &reuseTerminal).
		Bool("--smart-terminal", &smartTerminal).
		Bool("--no-new-terminal", &noNewTerminal).
		Bool("--open-in-agent", &openInAgent).
		Bool("--no-open-in-agent", &noOpenInAgent).
		Bool("--no-config", &noConfig).
		StringSlice("--bring", &bringPaths).
		Bool("--no-dep", &noDep).
		String("-t,--task", &taskDesc).
		String("--set-task", &setTaskDesc).
		Bool("--where", &where).
		Cut("--exec", &execArgs).
		Help("-h,--help", usage()).
		HelpNoExit().
		Parse(parseArgs)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			// Help text already printed by Parse; exit 0.
			ctx.skipEvent = true
			return nil
		}
		// lessflags says "unrecognized flag: …". Keep that wording for the
		// general case (stderr-newline locks the exact string), but map the
		// hard-removed --dep / --all-deps errors to "unknown flag" so callers
		// see an explicit unknown/invalid signal for deleted modes.
		msg := err.Error()
		if msg == "unrecognized flag: --dep" || msg == "unrecognized flag: --all-deps" {
			return fmt.Errorf("unknown flag: %s", strings.TrimPrefix(msg, "unrecognized flag: "))
		}
		return err
	}

	// --force-cd and --no-cd are mutually exclusive (hard error before any work).
	if forceCd && noCd {
		return fmt.Errorf("wrk: --force-cd and --no-cd are mutually exclusive")
	}
	// --color and --no-color are mutually exclusive (stdout three-mode policy).
	if colorFlag && noColorFlag {
		return fmt.Errorf("wrk: --color and --no-color are mutually exclusive")
	}

	// Manual commit message validation (order matches sealed commit-msg/validation intents).
	// Gen path peels --commit/--no-verify/--add-all; those flags then stay off top-level.
	// XOR with gen is checked early (before peel) when both appear on the raw argv.
	if commitMessage != nil && !commitFlag {
		return fmt.Errorf("wrk: -m/--message requires --commit")
	}
	if commitFlag && !genCommitMsg && commitMessage == nil {
		return fmt.Errorf("wrk: --commit requires -m/--message or --gen-commit-msg")
	}
	manualMessage := ""
	if commitMessage != nil {
		manualMessage = *commitMessage
		if strings.TrimSpace(manualMessage) == "" {
			return fmt.Errorf("wrk: commit message must not be empty")
		}
	}
	if noVerify && !commitFlag && !genCommitMsg {
		return fmt.Errorf("wrk: --no-verify requires --commit")
	}
	// --add-all stages before manual/gen commit, or cascade pin commits under --unwind.
	if addAll && !commitFlag && !genCommitMsg && !unwind {
		return fmt.Errorf("wrk: --add-all requires --commit")
	}
	// True when wrk should run the manual commit stage (message source present + --commit).
	manualCommit := commitFlag && commitMessage != nil && !genCommitMsg

	taskFlagSet := taskDesc != nil
	setTaskFlagSet := setTaskDesc != nil
	bringMode := len(bringPaths) > 0
	addFlagSet := addPath != nil
	removeFlagSet := removePath != nil
	whereFlagSet := where
	portFlagSet := portFlag != nil

	if webFlag {
		ctx.command = "web"
	} else if scanGitRepos {
		ctx.command = "scan-git-repos"
	} else {
		ctx.command = resolveCommand(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, done, list, status, repos, mergeBack, bringMode, reinstallLocal, tagNext, propagateTags, syncFlag, pushFlag, prFlag, cd, mainFlag, unwind)
		// P1: pure bare no-args is dashboard, not create. --new / positionals / -t
		// and create modifiers keep command "create".
		if ctx.command == "create" && isDashboardBareEntry(newFlag, taskFlagSet, remaining, createUXFlags{
			newWindow: newWindow, noNewWindow: noNewWindow,
			newTerminal: newTerminal, reuseTerminal: reuseTerminal, smartTerminal: smartTerminal,
			noNewTerminal: noNewTerminal, openInAgent: openInAgent, noOpenInAgent: noOpenInAgent,
		}, noConfig, noCd, forceCd, len(execArgs) > 0) {
			ctx.command = "dashboard"
		}
	}
	ctx.eventArgs = extractEventArgs(args, remaining)

	setInvocationVerbose(verbose)
	// Keep force color for main's FormatStderrError after Run returns (do not
	// clear in defer — main prints err after Run exits).
	SetForceStderrColor(colorFlag)
	worktree.GitVerboseLogger = logGitCommand
	defer func() {
		setInvocationVerbose(false)
		worktree.GitVerboseLogger = nil
	}()

	// --no-cache is only valid with --scan-git-repos.
	if noCache && !scanGitRepos {
		return fmt.Errorf("wrk: --no-cache is only valid with --scan-git-repos")
	}
	// --include-worktrees is only valid with --scan-git-repos.
	if includeWorktrees && !scanGitRepos {
		return fmt.Errorf("wrk: --include-worktrees is only valid with --scan-git-repos")
	}
	// --no-dep is only valid with --bring.
	if noDep && !bringMode {
		return fmt.Errorf("wrk: --no-dep is only valid with --bring")
	}

	if fetchFlag && !projects && !status && !webFlag {
		return fmt.Errorf("wrk: --fetch is only valid with --projects or --status")
	}
	if githubFlag && !projects {
		return fmt.Errorf("wrk: --github is only valid with --projects")
	}

	// remaining holds 0, 1, or 2 positionals for most modes:
	//   remaining[0] = sourceDir (valid for ALL modes — cwd when absent)
	//   remaining[1] = spawnTarget (create-only, was targetDir)
	// --scan-git-repos treats all remaining args as scan roots (any count).
	if !scanGitRepos && len(remaining) > 2 {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	var sourceDir string
	var spawnTarget string
	if len(remaining) >= 1 {
		sourceDir = remaining[0]
	}
	if len(remaining) == 2 {
		spawnTarget = remaining[1]
	}
	// --bring does not accept leftover positionals (no multi-value sugar; no workDir override).
	if bringMode && len(remaining) > 0 {
		return fmt.Errorf("wrk: unexpected arguments")
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	ctx.wrkHome = wrkHome

	// --port / --dev are only valid with --web (reject before other mode work).
	if portFlagSet && !webFlag {
		ctx.workDir = origWd
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		return fmt.Errorf("wrk: --port is only valid with --web")
	}
	if webDev && !webFlag {
		ctx.workDir = origWd
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		return fmt.Errorf("wrk: --dev is only valid with --web")
	}

	// --web is a standalone long-running mode: local HTTP UI + wrkserver API.
	if webFlag {
		otherMode := done || mergeBack || list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || reinstallLocal || tagNext || propagateTags || syncFlag ||
			dryRun || pushFlag || prFlag || jsonFlag || taskFlagSet || setTaskFlagSet || fetchFlag || noCd || forceCd ||
			cd || mainFlag || unwind || confirmFromStdin || forceConfirm || noInModuleReplace || scanGitRepos ||
			newFlag || newWindow || noNewWindow || newTerminal || reuseTerminal || smartTerminal ||
			noNewTerminal || openInAgent || noOpenInAgent || len(execArgs) > 0
		ctx.workDir = origWd
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		if otherMode {
			return fmt.Errorf("wrk: --web is mutually exclusive with other modes")
		}
		if len(remaining) > 0 {
			return fmt.Errorf("wrk: unexpected arguments")
		}
		port := 0
		if portFlagSet {
			port = *portFlag
		}
		return runWeb(WebServeOptions{WrkHome: wrkHome, Port: port, Dev: webDev})
	}

	// --scan-git-repos discovers main repos under roots (print-only; never records).
	if scanGitRepos {
		otherMode := done || mergeBack || list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || bringMode ||
			reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || unwind || pushFlag || prFlag || jsonFlag || taskFlagSet ||
			setTaskFlagSet || fetchFlag || noCd || forceCd || cd || mainFlag ||
			confirmFromStdin || forceConfirm || noInModuleReplace || webFlag ||
			newFlag || newWindow || noNewWindow || newTerminal || reuseTerminal || smartTerminal ||
			noNewTerminal || openInAgent || noOpenInAgent || noConfig || len(execArgs) > 0
		ctx.workDir = origWd
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		if otherMode {
			return fmt.Errorf("wrk: --scan-git-repos is mutually exclusive with other modes")
		}
		if err := ctx.autoRecord(); err != nil {
			return err
		}
		return runScanGitRepos(wrkHome, remaining, noCache, includeWorktrees, verbose, opts.ScanTestPauseAfterFirstPrint)
	}

	// --cd and --where are always mutually exclusive (including under --main).
	// Prefer this over arity when both Bool flags leave two positionals.
	if cd && whereFlagSet {
		ctx.workDir = origWd
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		return fmt.Errorf("wrk: --cd is mutually exclusive with other modes")
	}

	// --cd / --where share arity: without --main exactly one path/basename positional;
	// with --main exactly zero positionals (main is resolved from cwd).
	if cd {
		if mainFlag {
			if len(remaining) > 0 {
				ctx.workDir = origWd
				if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
					return err
				}
				return fmt.Errorf("wrk: unexpected arguments")
			}
		} else if len(remaining) == 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --cd requires a path argument")
		} else if len(remaining) > 1 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}
	if whereFlagSet {
		if prFlag {
			// --where --pr: exactly one full GitHub PR URL positional.
			if len(remaining) == 0 {
				ctx.workDir = origWd
				if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
					return err
				}
				return errWherePRNeedsFullURL
			}
			if len(remaining) > 1 {
				ctx.workDir = origWd
				if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
					return err
				}
				return fmt.Errorf("wrk: unexpected arguments")
			}
		} else if mainFlag {
			if len(remaining) > 0 {
				ctx.workDir = origWd
				if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
					return err
				}
				return fmt.Errorf("wrk: unexpected arguments")
			}
		} else if len(remaining) == 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --where requires a path argument")
		} else if len(remaining) > 1 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// --main takes no path positional when used alone. It may compose with --status
	// (status of main repo, no shell); positionals then follow --status rules.
	// It may also compose with --where / --cd (print main / runCd main; zero positionals),
	// with --reinstall-local (and --dry-run as its modifier), and with activeRoot
	// pipeline stages (sync/tag-next/push/propagate/reinstall/exec).
	// Mutual exclusion with other modes is checked later; if another mode flag is
	// also set, prefer that error over unexpected arguments.
	if mainFlag {
		// Pipeline partners, reinstall-local, --where, and --cd are compose partners (not otherMode).
		mainPipelinePartner := reinstallLocal || tagNext || propagateTags || syncFlag || pushFlag || len(execArgs) > 0
		otherMode := done || list || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet ||
			bringMode || jsonFlag || mergeBack || taskFlagSet ||
			setTaskFlagSet || noCd || spawnTarget != ""
		if !mainPipelinePartner && !status && !whereFlagSet && !cd {
			otherMode = otherMode || dryRun
		}
		// --fetch is only valid with --status (or --projects); allow with --main --status.
		if !status {
			otherMode = otherMode || fetchFlag
		}
		if otherMode {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --main is mutually exclusive with other modes")
		}
		if !status && !mainPipelinePartner && !whereFlagSet && !cd && len(remaining) > 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// --sync takes no positionals when used alone. It may compose with --done /
	// --merge-back and with other pipeline stages (activeRoot model; no primary required).
	// Prefer mode-clash errors over unexpected args when combined with non-pipeline modes.
	// --json multi-stage is rejected later with a --json-named error (not here).
	if syncFlag {
		// Pipeline partners (done/merge-back/tag-next/push/propagate/reinstall/gen-commit/exec)
		// are intentionally excluded so multi-stage composition is allowed.
		// --main is a scope modifier for the activeRoot pipeline (not otherMode).
		otherMode := list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || bringMode ||
			taskFlagSet || setTaskFlagSet ||
			cd || fetchFlag || spawnTarget != ""
		// Bare --sync --json (no tag-next) still exclusive; multi-stage +json named later.
		if jsonFlag && !tagNext {
			otherMode = true
		}
		// noCd/forceCd are done/create modifiers: exclusive with bare/sync+merge-back,
		// allowed with --done --sync.
		if (noCd || forceCd) && !done {
			otherMode = true
		}
		if otherMode {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --sync is mutually exclusive with other modes")
		}
		if !done && !mergeBack && len(remaining) > 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// Resolve sourceDir to absolute; default to process cwd when absent.
	// Passed to every sub-command as workDir instead of using os.Getwd/Chdir.
	createMode := isCreateMode(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, repos, status, bringMode, reinstallLocal, tagNext, propagateTags, syncFlag, pushFlag, prFlag, list, done, mergeBack, cd, mainFlag, unwind)
	uxFlags := createUXFlags{
		newWindow:     newWindow,
		noNewWindow:   noNewWindow,
		newTerminal:   newTerminal,
		reuseTerminal: reuseTerminal,
		smartTerminal: smartTerminal,
		noNewTerminal: noNewTerminal,
		openInAgent:   openInAgent,
		noOpenInAgent: noOpenInAgent,
	}
	if err := uxFlags.validate(); err != nil {
		return err
	}
	if uxFlags.any() && !createMode {
		return fmt.Errorf("wrk: create UX flags are only valid with create mode")
	}
	dirHint := &DirHintOptions{
		RawArgs:     args,
		Positionals: remaining,
	}
	// Basename fallback: create/status/list/repos/--cd (without --main).
	// --where basename is a lookup key, not a source dir (workDir stays cwd).
	// --main / --main --where / --main --cd use cwd only.
	// One-arg create: if source resolve fails and the arg is task-like, offer
	// treat-as-task (promote creates from process cwd).
	var promotedTask string
	resolveSrc := sourceDir
	if whereFlagSet && !mainFlag {
		// remaining[0] is the basename operand; do not resolve as workDir.
		resolveSrc = ""
	}
	workDir, err := resolveSourceWorkDir(origWd, resolveSrc, createMode || status || list || repos || (cd && !mainFlag), wrkHome, dirHint)
	if err != nil {
		if createMode && !taskFlagSet && sourceDir != "" && spawnTarget == "" && isTaskLike(sourceDir) {
			promote, perr := confirmTaskLikePromote("source", sourceDir, assumeYes)
			if perr != nil {
				return perr
			}
			if promote {
				promotedTask = sourceDir
				workDir = origWd
				err = nil
			} else {
				return err
			}
		} else {
			return err
		}
	}
	ctx.workDir = workDir
	if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
		return err
	}
	if err := ctx.autoRecord(); err != nil {
		return err
	}

	if addFlagSet && strings.TrimSpace(*addPath) == "" {
		return fmt.Errorf("wrk: --add requires a path argument")
	}
	if removeFlagSet && strings.TrimSpace(*removePath) == "" {
		return fmt.Errorf("wrk: --rm requires a path argument")
	}

	hasExec := len(execArgs) > 0
	if hasExec {
		// --exec is the last pipeline stage of any activeRoot compose (with or without
		// --done/--merge-back). Still invalid with non-pipeline modes.
		if list {
			return fmt.Errorf("wrk: --exec is not valid with --list")
		}
		if status {
			return fmt.Errorf("wrk: --exec is not valid with --status")
		}
		if repos {
			return fmt.Errorf("wrk: --exec is not valid with --repos")
		}
		if projects {
			return fmt.Errorf("wrk: --exec is not valid with --projects")
		}
		if projectsDepGraph {
			return fmt.Errorf("wrk: --exec is not valid with --projects-dep-graph")
		}
		if addFlagSet {
			return fmt.Errorf("wrk: --exec is not valid with --add")
		}
		if removeFlagSet {
			return fmt.Errorf("wrk: --exec is not valid with --rm")
		}
		if whereFlagSet {
			return fmt.Errorf("wrk: --exec is not valid with --where")
		}
		// --exec may compose with --main as the last activeRoot pipeline stage
		// (activeRoot rewritten to main). Bare --main --exec is also allowed.
	}

	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && strings.TrimSpace(*setTaskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && (taskFlagSet || done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || spawnTarget != "" || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --set-task is mutually exclusive with other flags")
	}
	if setTaskFlagSet {
		// Default auto-yes for rename prompt; --confirm restores Y/n; -y still auto-yes.
		return runSetTask(workDir, *setTaskDesc, planAssumeYes(assumeYes, forceConfirm), noCd, forceCd, execArgs)
	}

	if taskFlagSet && strings.TrimSpace(*taskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --task is only valid with create mode.
	if taskFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || mergeBack || prFlag || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --task is mutually exclusive with --done, --merge-back, --list, --status, --repos, --projects, --add, --rm, --where, and --bring")
	}

	// --new is the create entry; exclusive with non-create modes.
	if newFlag {
		otherMode := done || mergeBack || list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || setTaskFlagSet || bringMode ||
			reinstallLocal || tagNext || propagateTags || syncFlag || pushFlag || prFlag || jsonFlag ||
			cd || mainFlag || dryRun
		if otherMode {
			return fmt.Errorf("wrk: --new is mutually exclusive with other modes")
		}
	}

	// --title / --comment are only valid with --pr.
	if !prFlag && (prTitle != nil || prComment != nil) {
		return fmt.Errorf("wrk: --title and --comment are only valid with --pr")
	}
	// --pr is a primary that may compose with peel partners --push and
	// --gen-commit-msg (fixed order: gen-commit → push → pr), with
	// --status for read-only PR status, and with --where for location lookup
	// (full GitHub PR URL → local worktree path). Still exclusive with other
	// modes (--done, --merge-back, --list, --main, etc.).
	// Modes:
	//   bare --pr (no --title/--comment): show open PR URL (or empty)
	//   --pr --status: print open PR metadata + checks/reviews rollup
	//   --where --pr <url>: print local worktree path(s) for PR head branch
	//   --pr --comment C (no --title): comment-only on existing open PR
	//   --pr --push (no --title/--comment): push-existing (open PR required → full push → URL)
	//   --pr --push --comment C (no --title): push then comment on open PR
	//   --pr --title T --comment C: create/attach
	//   incomplete create (title without comment): error
	//   --gen-commit-msg with --pr still requires --title (create compose)
	//   --pr --status + --title/--comment/--push: invalid combination
	if prFlag {
		// status is allowed with --pr (PR status mode); validated below.
		// --where is a compose partner for location lookup (not otherMode).
		otherMode := done || mergeBack || list || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || bringMode || reinstallLocal || tagNext ||
			propagateTags || syncFlag || dryRun || jsonFlag || taskFlagSet || setTaskFlagSet ||
			spawnTarget != "" || cd || mainFlag || unwind || fetchFlag || hasExec || newFlag ||
			confirmFromStdin || forceConfirm || noInModuleReplace || noCd || forceCd
		if otherMode {
			return fmt.Errorf("wrk: --pr is mutually exclusive with other modes")
		}
		if whereFlagSet {
			// --where --pr location lookup: no create/attach, comment, push, status.
			if status {
				return fmt.Errorf("wrk: --where is mutually exclusive with other modes")
			}
			if prTitle != nil || prComment != nil {
				return fmt.Errorf("wrk: --where --pr cannot be combined with --title or --comment")
			}
			if pushFlag {
				return fmt.Errorf("wrk: --where --pr cannot be combined with --push")
			}
			if genCommitMsg {
				return fmt.Errorf("wrk: --where --pr cannot be combined with --gen-commit-msg")
			}
		} else if status {
			// PR status is read-only: no create/attach, comment, or push.
			if prTitle != nil || prComment != nil {
				return fmt.Errorf("wrk: --pr --status cannot be combined with --title or --comment")
			}
			if pushFlag {
				return fmt.Errorf("wrk: --pr --status cannot be combined with --push")
			}
			if genCommitMsg {
				return fmt.Errorf("wrk: --pr --status cannot be combined with --gen-commit-msg")
			}
		} else if prTitle == nil && prComment == nil {
			// Bare show, or push-existing when --push (no title/comment).
			// gen-commit + pr still needs create/attach title.
			if genCommitMsg {
				return fmt.Errorf("wrk: --title is required with --pr")
			}
		} else if prTitle == nil {
			// Comment-only, or push+comment when --push (no title).
			// Non-empty body required. gen-commit compose still needs title.
			if strings.TrimSpace(*prComment) == "" {
				return fmt.Errorf("wrk: --comment must not be empty")
			}
			if genCommitMsg {
				return fmt.Errorf("wrk: --title is required with --pr")
			}
		} else {
			// Title set: create/attach requires --comment (and both non-empty).
			if prComment == nil {
				return fmt.Errorf("wrk: --comment is required with --pr")
			}
			if strings.TrimSpace(*prTitle) == "" {
				return fmt.Errorf("wrk: --title must not be empty")
			}
			if strings.TrimSpace(*prComment) == "" {
				return fmt.Errorf("wrk: --comment must not be empty")
			}
		}
	}

	if list && done {
		return fmt.Errorf("wrk: --list and --done are mutually exclusive")
	}
	if list && mergeBack {
		return fmt.Errorf("wrk: --list and --merge-back are mutually exclusive")
	}
	if done && mergeBack {
		return fmt.Errorf("wrk: --done and --merge-back are mutually exclusive")
	}
	if repos && (done || list || status || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || spawnTarget != "" || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --repos is mutually exclusive with other modes")
	}
	if projects && (done || list || status || repos || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --projects is mutually exclusive with other modes")
	}
	if projectsDepGraph && (done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag || fetchFlag || unwind) {
		return fmt.Errorf("wrk: --projects-dep-graph is mutually exclusive with other modes")
	}
	if addFlagSet && (done || list || status || repos || projects || projectsDepGraph || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --add is mutually exclusive with other modes")
	}
	if removeFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: --rm is mutually exclusive with other modes")
	}
	// --where composes with --main (print main path) and --pr (PR URL → worktree path).
	// Still exclusive with --cd and other modes. prFlag is carved out here; invalid
	// --where --pr + title/comment/push/status is checked in the prFlag block above.
	if whereFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || cd || unwind) {
		return fmt.Errorf("wrk: --where is mutually exclusive with other modes")
	}
	// --cd composes with --main (runCd main). Still exclusive with --where and other modes.
	if cd && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || prFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || noCd || unwind) {
		return fmt.Errorf("wrk: --cd is mutually exclusive with other modes")
	}
	// --main composes with --status (and --fetch when status is set), with
	// --where / --cd, with --reinstall-local, and with activeRoot pipeline stages
	// (sync/tag-next/push/propagate-tags/reinstall-local/exec, plus --dry-run as modifier).
	// Exclusive with done/merge-back, gen-commit-msg (checked earlier), and non-pipeline modes.
	if mainFlag {
		mainPipelinePartner := reinstallLocal || tagNext || propagateTags || syncFlag || pushFlag || hasExec
		otherMode := done || list || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || bringMode || jsonFlag || mergeBack || prFlag || taskFlagSet || setTaskFlagSet || spawnTarget != "" || noCd || unwind || (!status && fetchFlag)
		if !mainPipelinePartner && !status && !whereFlagSet && !cd {
			otherMode = otherMode || dryRun
		}
		if otherMode {
			return fmt.Errorf("wrk: --main is mutually exclusive with other modes")
		}
	}
	// --status composes with --pr (PR status) and --main; exclusive with push/list/etc.
	// prFlag is carved out here; invalid --pr --status + title/comment/push is checked above.
	// --unwind remains exclusive with bare --status.
	if status && (done || list || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || spawnTarget != "" || cd || unwind) {
		return fmt.Errorf("wrk: --status is mutually exclusive with other modes")
	}
	if confirmFromStdin && !done && !mergeBack {
		return fmt.Errorf("wrk: --confirm-from-stdin is only valid with --done or --merge-back")
	}
	if forceConfirm && !done && !mergeBack && !setTaskFlagSet {
		return fmt.Errorf("wrk: --confirm is only valid with --done, --merge-back, or --set-task")
	}
	if noInModuleReplace && !done {
		return fmt.Errorf("wrk: --no-in-module-replace is only valid with --done")
	}
	if bringMode && (done || list || mergeBack || tagNext || propagateTags || syncFlag || cd || mainFlag || reinstallLocal || unwind) {
		return fmt.Errorf("wrk: --bring is mutually exclusive with --done, --merge-back and --list")
	}
	// --reinstall-local may compose with pipeline stages (activeRoot model) and with
	// --main / --done / --merge-back. Still exclusive with list/status/repos and similar.
	if reinstallLocal {
		otherMode := list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || bringMode ||
			cd || taskFlagSet || setTaskFlagSet ||
			spawnTarget != "" || jsonFlag || fetchFlag
		if done || mergeBack {
			// Primary compose: post stages and done modifiers are allowed.
		} else if !tagNext && !propagateTags && !syncFlag && !pushFlag && !genCommitMsg && !manualCommit && !hasExec {
			// Bare / --main reinstall only: exclusive with primary-only modifiers.
			otherMode = otherMode || confirmFromStdin || forceConfirm || noInModuleReplace || noCd || forceCd
		} else {
			// Multi-stage without primary: still reject done-only modifiers.
			otherMode = otherMode || confirmFromStdin || forceConfirm || noInModuleReplace || noCd || forceCd
		}
		if otherMode {
			return fmt.Errorf("wrk: --reinstall-local is mutually exclusive with other modes")
		}
	}
	// --tag-next may compose with other pipeline stages (activeRoot must be main at
	// that stage). Still exclusive with list/status/repos and other non-pipeline modes.
	// With --unwind, --tag-next is a pin modifier (not exclusive).
	if tagNext && !unwind {
		// --main is a scope modifier (activeRoot := main), not exclusive with --tag-next.
		otherMode := bringMode || list || cd ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != ""
		if otherMode {
			return fmt.Errorf("wrk: --tag-next is mutually exclusive with other modes")
		}
	}
	// --propagate-tags may compose with pipeline stages (activeRoot model).
	// --push alone with bare --propagate-tags (no tag-next/done/merge-back) is invalid.
	// --json is rejected separately so the error names both flags.
	if propagateTags {
		otherMode := bringMode || list || cd ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != "" || unwind
		if !done && !mergeBack && !tagNext && pushFlag {
			// --push alone with bare --propagate-tags is invalid; only with --tag-next compose.
			otherMode = true
		}
		if otherMode {
			return fmt.Errorf("wrk: --propagate-tags is mutually exclusive with other modes")
		}
	}
	// --sync may compose with pipeline stages (activeRoot model); exclusive with
	// non-pipeline modes. Multi-stage + --json is rejected by the --json check below.
	if syncFlag {
		otherMode := bringMode || list || cd ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != ""
		if jsonFlag && !tagNext {
			otherMode = true
		}
		if otherMode {
			return fmt.Errorf("wrk: --sync is mutually exclusive with other modes")
		}
	}
	// --push is a bare primary (option R: push current checkout branch), or a
	// composition stage with other pipeline flags. Bare --push is exclusive with
	// non-pipeline modes. --json is rejected separately so the error names --json.
	// --main is a scope modifier and may compose with --push (pipeline activeRoot).
	// With --unwind, --push is a pin modifier (not exclusive).
	// --pr is a peel partner (gen-commit → push → pr); not exclusive with --push.
	if pushFlag && !tagNext && !done && !mergeBack && !unwind && !prFlag {
		otherMode := bringMode || list || cd ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != ""
		// propagate without tag-next/done still invalid (handled above for propagate).
		if otherMode {
			return fmt.Errorf("wrk: --push is mutually exclusive with other modes")
		}
	}
	// --json is only valid with bare --tag-next (optionally --push), or with
	// --unwind --show-graph; never with multi-stage compose or --done / --merge-back /
	// --propagate-tags.
	if jsonFlag && done {
		return fmt.Errorf("wrk: --json is not valid with --done")
	}
	if jsonFlag && mergeBack {
		return fmt.Errorf("wrk: --json is not valid with --merge-back")
	}
	if jsonFlag && propagateTags {
		return fmt.Errorf("wrk: --json is not valid with --propagate-tags")
	}
	jsonWithShowGraph := unwind && showGraph
	if jsonFlag && !tagNext && !jsonWithShowGraph {
		return fmt.Errorf("wrk: --json is only valid with --tag-next or --unwind --show-graph")
	}
	if jsonFlag && tagNext && (syncFlag || reinstallLocal || genCommitMsg || manualCommit || hasExec) {
		return fmt.Errorf("wrk: --json is not valid with multi-stage compose (only with bare --tag-next)")
	}
	// --dry-run is valid with bare --sync / --tag-next / --propagate-tags /
	// --reinstall-local / --push, with --done / --merge-back composition (full multi-stage plan is later phases),
	// with --unwind (stack peel plan), with --gen-commit-msg (handled early via runGenCommitMsg),
	// and with manual --commit -m/--message.
	if dryRun && !done && !mergeBack && !tagNext && !propagateTags && !syncFlag && !reinstallLocal && !pushFlag && !unwind && !manualCommit {
		return fmt.Errorf("wrk: --dry-run is only valid with --done, --merge-back, --tag-next, --propagate-tags, --sync, --reinstall-local, --push, --unwind, --gen-commit-msg, or --commit -m/--message")
	}
	// -f/--force is a push modifier only (D5): never a bare primary.
	if forcePush && !pushFlag {
		return fmt.Errorf("wrk: -f/--force is only valid with --push")
	}

	// Manual --commit -m is a pipeline stage (same partners as gen-commit-msg).
	// Exclusive with non-pipeline modes (status/list/create/etc.).
	if manualCommit {
		otherMode := list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || bringMode ||
			taskFlagSet || setTaskFlagSet || cd || mainFlag ||
			newFlag || webFlag || scanGitRepos || spawnTarget != "" || jsonFlag || fetchFlag
		if otherMode {
			return fmt.Errorf("wrk: --commit is mutually exclusive with other modes")
		}
	}

	// spawnTarget only applies to the create path. Reject for any other mode.
	if spawnTarget != "" && (bringMode || reinstallLocal || tagNext || propagateTags || syncFlag || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || done || mergeBack || prFlag || cd || mainFlag || unwind) {
		return fmt.Errorf("wrk: unexpected arguments")
	}

	// --show-graph is only valid with --unwind (read-only stack graph).
	if showGraph && !unwind {
		return fmt.Errorf("wrk: --show-graph is only valid with --unwind")
	}

	// --unwind is a primary mode: stack DAG plan / cycle preflight / free-first peel.
	// Composes with ship/land flags and repository-local post stages.
	// --show-graph is a read-only submode: exclusive with dry-run and apply partners.
	if unwind {
		otherMode := list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || bringMode ||
			propagateTags ||
			taskFlagSet || setTaskFlagSet || cd || mainFlag ||
			newFlag || hasExec || spawnTarget != ""
		// --json is allowed only with --show-graph (G6); bare unwind rejects --json.
		if jsonFlag && !showGraph {
			otherMode = true
		}
		if otherMode {
			return fmt.Errorf("wrk: --unwind is mutually exclusive with other modes")
		}
		if confirmFromStdin || forceConfirm || noInModuleReplace {
			return fmt.Errorf("wrk: --unwind is mutually exclusive with other modes")
		}
		if showGraph {
			if dryRun {
				return fmt.Errorf("wrk: --show-graph cannot be used with --dry-run")
			}
			if done {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --done")
			}
			if mergeBack {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --merge-back")
			}
			if tagNext {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --tag-next")
			}
			if pushFlag {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --push")
			}
			if forcePush {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with -f/--force")
			}
			if syncFlag {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --sync")
			}
			if reinstallLocal {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --reinstall-local")
			}
			if genCommitMsg {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --gen-commit-msg")
			}
			if manualCommit {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --commit")
			}
			if propagateTags {
				return fmt.Errorf("wrk: --show-graph is mutually exclusive with --propagate-tags")
			}
		}
		return runUnwind(workDir, UnwindFlags{
			DryRun:         dryRun,
			TagNext:        tagNext,
			Push:           pushFlag,
			Force:          forcePush,
			Done:           done,
			MergeBack:      mergeBack,
			ReinstallLocal: reinstallLocal,
			Color:          colorFlag,
			NoColor:        noColorFlag,
			Sync:           syncFlag,
			GenCommitMsg:   genCommitMsg,
			GenCommitArgs:  genCommitArgs,
			AddAll:         addAll,
			ShowGraph:      showGraph,
			JSON:           jsonFlag,
		})
	}

	if projects {
		colorEnabled := term.IsTerminal(int(os.Stdout.Fd())) || colorFlag
		return runProjects(wrkHome, colorEnabled, fetchFlag, githubFlag)
	}
	if projectsDepGraph {
		return runProjectsDepGraph(wrkHome)
	}
	if addFlagSet {
		return runAdd(wrkHome, *addPath)
	}
	if removeFlagSet {
		return runRemove(wrkHome, *removePath)
	}
	// --where: basename lookup, --main print main abs path, or --pr full URL → worktree path(s).
	if whereFlagSet {
		if mainFlag {
			mainRepo, err := resolveMainRepoForWorkDir(workDir)
			if err != nil {
				return err
			}
			fmt.Println(mainRepo)
			return nil
		}
		if prFlag {
			return runWherePR(wrkHome, workDir, remaining[0])
		}
		return runWhere(wrkHome, remaining[0])
	}
	if status {
		if prFlag {
			// PR status mode (--pr --status): open PR metadata + check rollup.
			return runPRStatus(workDir)
		}
		colorEnabled := term.IsTerminal(int(os.Stdout.Fd())) || colorFlag
		statusRoot := workDir
		if mainFlag {
			// Status the main repository of this checkout (no nested shell).
			// Dir labels still use original workDir as display base.
			mainRepo, err := resolveMainRepoForWorkDir(workDir)
			if err != nil {
				return err
			}
			statusRoot = mainRepo
		}
		return runStatus(statusRoot, workDir, colorEnabled, fetchFlag)
	}
	// Bare / --main --reinstall-local before bare --main so compose does not open a nested shell.
	// Multi-stage reinstall is handled by activeRoot pipeline below.
	if reinstallLocal && !done && !mergeBack && !genCommitMsg && !manualCommit && !syncFlag && !tagNext && !pushFlag && !propagateTags && !hasExec {
		return runReinstallLocal(workDir, dryRun, mainFlag, colorFlag)
	}
	// --cd: resolve target (or with --main, main repo of cwd) then jump.
	// Handled before bare --main so --main --cd never opens a nested shell.
	if cd {
		target := workDir
		if mainFlag {
			mainRepo, err := resolveMainRepoForWorkDir(workDir)
			if err != nil {
				return err
			}
			target = mainRepo
		}
		return runCd(target, execArgs)
	}
	// --main with pipeline partners: rewrite activeRoot to main (no nested shell).
	// Bare --main alone still opens a nested shell via runMain.
	if mainFlag {
		mainPipelinePartner := tagNext || propagateTags || syncFlag || pushFlag || reinstallLocal || hasExec
		if !mainPipelinePartner {
			return runMain(workDir)
		}
		mainRepo, err := resolveMainRepoForWorkDir(workDir)
		if err != nil {
			return err
		}
		// Already on main: non-fatal notice; continue pipeline under main.
		cwdAbs, absErr := filepath.Abs(workDir)
		if absErr == nil {
			top, topErr := worktree.ShowToplevel(cwdAbs)
			if topErr == nil && sameDirPath(top, mainRepo) {
				fmt.Fprintf(os.Stderr, "wrk: --main is not necessary (already at main repository root); continuing\n")
			}
		}
		workDir = mainRepo
		// Fall through to activeRoot pipeline / bare stage handlers with main as workDir.
	}
	if repos {
		return runRepos(workDir)
	}
	if bringMode {
		return runBring(workDir, bringPaths, wrkHome, args, execArgs, noDep)
	}
	if list {
		return runList(workDir)
	}
	// Prefer done / merge-back over bare tag-next / propagate / sync so composition
	// runs the primary path (post-pipeline: sync → tag-next → push → propagate-tags → reinstall-local).
	// Optional pre-stage: --gen-commit-msg --commit … or manual --commit -m on the source worktree.
	// After successful done/merge-back, activeRoot switches to main for later stages.
	if done {
		if err := runGenCommitMsgPreStage(workDir, genCommitMsg, genCommitArgs, dryRun, "--done"); err != nil {
			return err
		}
		if err := runManualCommitPreStage(workDir, manualCommit, manualMessage, noVerify, addAll, dryRun); err != nil {
			return err
		}
		runPrimary := func() error {
			// Own keeps default auto-yes; cascade not-included requires -y or explicit confirm (D3).
			return runDone(workDir, wrkHome, confirmFromStdin, assumeYes, forceConfirm, noInModuleReplace, noCd, forceCd, execArgs, syncFlag, tagNext, pushFlag, forcePush, propagateTags, reinstallLocal, dryRun, colorFlag)
		}
		// Dry-run gen/manual commit pre would commit staged dirt; MergeBack --rm still
		// requires a clean tree today. Stash staged only for the dry plan, then restore.
		if dryRun && (genCommitMsg || manualCommit) {
			return withStashedStagedForDryPlan(workDir, runPrimary)
		}
		return runPrimary()
	}
	if mergeBack {
		if err := runGenCommitMsgPreStage(workDir, genCommitMsg, genCommitArgs, dryRun, "--merge-back"); err != nil {
			return err
		}
		if err := runManualCommitPreStage(workDir, manualCommit, manualMessage, noVerify, addAll, dryRun); err != nil {
			return err
		}
		// merge-back keeps the worktree (Remove=false); dirty is allowed by MergeBack.
		// Default auto-yes; --confirm restores prompts; -y still auto-yes.
		return runMergeBack(workDir, wrkHome, confirmFromStdin, planAssumeYes(assumeYes, forceConfirm), syncFlag, tagNext, pushFlag, forcePush, propagateTags, reinstallLocal, dryRun, colorFlag)
	}
	// Multi-stage without done/merge-back: fixed order on activeRoot (= cwd toplevel).
	// Stages: gen-commit|manual-commit → sync → tag-next → push → pr → propagate-tags → reinstall-local → exec.
	{
		stageN := 0
		if genCommitMsg || manualCommit {
			stageN++
		}
		if syncFlag {
			stageN++
		}
		if tagNext {
			stageN++
		}
		if pushFlag {
			stageN++
		}
		if prFlag {
			stageN++
		}
		if propagateTags {
			stageN++
		}
		if reinstallLocal {
			stageN++
		}
		if hasExec {
			stageN++
		}
		// tag-next + push [+json] stays on the dedicated bare path (json-clean stdout).
		bareTagPushJSON := tagNext && pushFlag && jsonFlag && stageN == 2
		if stageN > 1 && !bareTagPushJSON {
			title, comment := "", ""
			if prFlag {
				// Title/comment may be nil for push-existing / push+comment.
				if prTitle != nil {
					title = *prTitle
				}
				if prComment != nil {
					comment = *prComment
				}
			}
			return runActiveRootPipeline(workDir, wrkHome, genCommitMsg, genCommitArgs, manualCommit, manualMessage, noVerify, addAll, syncFlag, tagNext, pushFlag, forcePush, prFlag, title, comment, propagateTags, reinstallLocal, dryRun, colorFlag, execArgs)
		}
	}
	// Bare compose: --tag-next --propagate-tags [--push] [--dry-run].
	// Fixed stage order tag-next → push? → propagate-tags.
	if tagNext && propagateTags {
		if err := requireMainActiveRoot(workDir, "--tag-next"); err != nil {
			return err
		}
		return runTagNextPropagateCompose(workDir, wrkHome, dryRun, pushFlag, forcePush)
	}
	if tagNext {
		if err := requireMainActiveRoot(workDir, "--tag-next"); err != nil {
			return err
		}
		// Create tags locally only; push (if any) is via runPushMain with tag list
		// so branch + tags are published (not tagscope tags-only push).
		tagRes, err := runTagNextAtResult(workDir, "HEAD", dryRun, false, jsonFlag)
		if err != nil {
			return err
		}
		if pushFlag {
			if !jsonFlag {
				fmt.Println() // blank line between tag-next block and push confirm
			}
			// With --json: still push branch+tags, but keep stdout JSON-clean.
			if err := runPushMainWithOutput(tagRes.MainRepo, dryRun, forcePush, tagRes.Tags, !jsonFlag); err != nil {
				return err
			}
		}
		return nil
	}
	if propagateTags {
		return runPropagateTags(workDir, wrkHome, dryRun)
	}
	if syncFlag {
		return runSync(workDir, dryRun)
	}
	// Bare --push: option R — push current checkout branch (linked worktree →
	// that worktree's branch; main → main's branch). -f/--force → force-with-lease.
	if pushFlag {
		return runBarePush(workDir, dryRun, forcePush)
	}
	// Bare --pr: show / comment-only / create-attach depending on flags.
	if prFlag {
		if prTitle == nil && prComment == nil {
			return runPRShow(workDir)
		}
		if prTitle == nil {
			return runPRComment(workDir, *prComment, colorFlag)
		}
		return runPR(workDir, *prTitle, *prComment, colorFlag)
	}
	// Bare manual commit: --commit -m/--message (no pipeline partners).
	if manualCommit {
		ctx.command = "commit"
		return runManualCommitStage(workDir, manualMessage, noVerify, addAll, dryRun)
	}
	task := ""
	if taskDesc != nil {
		task = *taskDesc
	}
	if promotedTask != "" {
		task = promotedTask
	}

	// Bare no-args (no --new, no positionals, no create-selecting flags) → dashboard View.
	// Create still runs with --new, <dir>, -t/--task, create UX flags, or create modifiers.
	if isDashboardBareEntry(newFlag, taskFlagSet || promotedTask != "", remaining, uxFlags, noConfig, noCd, forceCd, len(execArgs) > 0) {
		ctx.command = "dashboard"
		return runDashboard(workDir, ctx)
	}

	// Two-arg create without -t: second positional may be a forgotten task description.
	// When -t is already set, second remains target-dir (no treat-as-task).
	if createMode && !taskFlagSet && promotedTask == "" && spawnTarget != "" {
		if isTaskLike(spawnTarget) && !isExistingDirArg(spawnTarget, origWd) {
			promote, perr := confirmTaskLikePromote("target", spawnTarget, assumeYes)
			if perr != nil {
				return perr
			}
			if promote {
				task = spawnTarget
				spawnTarget = ""
			}
		}
	}

	// With <target-dir> or --no-config, skip config create.* UX; CLI flags still apply.
	uxPlan, err := resolveCreateUX(wrkHome, uxFlags, spawnTarget == "" && !noConfig)
	if err != nil {
		return err
	}
	return runCreate(workDir, origWd, spawnTarget, task, noCd, forceCd, execArgs, uxPlan)
}

// isDashboardBareEntry is true when the invocation is pure bare no-args dashboard
// (not create). Create-selecting: --new, positionals, --task/-t, create UX flags,
// --no-config, --no-cd/--force-cd, or --exec.
func isDashboardBareEntry(newFlag, taskFlagSet bool, remaining []string, ux createUXFlags, noConfig, noCd, forceCd, hasExec bool) bool {
	if newFlag || taskFlagSet || len(remaining) > 0 || ux.any() || noConfig || noCd || forceCd || hasExec {
		return false
	}
	return true
}

// usage returns the wrk help text printed by lessflags when -h/--help is given.
func usage() string {
	return `wrk — git worktree helper

Usage:
  wrk                              open dashboard snapshot (no create)
  wrk --new [dir] [target-dir] [flags]
  wrk [dir] [target-dir] [flags]   create when dir/task/UX args select create

With --new (or create-selecting args such as <dir> / -t), creates a git worktree
from the current directory (or <dir>) and prints its path. With <target-dir>, the
worktree is spawned there instead of the default location (~/.wrk/worktrees/).
Bare wrk with no args prints the dashboard stage snapshot (does not create a worktree).

Positional arguments:
  <dir>          optional source checkout to create the worktree from
                 (default: current directory)
  <target-dir>   optional spawn location for the worktree:
                   - missing, parent exists   -> spawn exactly at <target-dir>
                   - existing directory        -> spawn a default-named sub-dir
                   - missing parent            -> error

Flags:
  --new                           create a worktree (explicit create entry)
  --done [--gen-commit-msg --commit … | --commit -m MSG] [--sync] [--tag-next] [--push] [--propagate-tags] [--reinstall-local] [--dry-run] [--confirm] [--confirm-from-stdin]
                                  merge worktree branch back and remove it (default auto-yes)
                                  (optional pre: gen or manual --commit -m on worktree; optional post-success: --sync, --tag-next, --push, --propagate-tags, --reinstall-local from main)
  --merge-back [--gen-commit-msg --commit … | --commit -m MSG] [--sync] [--tag-next] [--push] [--propagate-tags] [--reinstall-local] [--dry-run] [--confirm] [--confirm-from-stdin]
                                  merge worktree branch back WITHOUT removing it (default auto-yes)
                                  (optional pre: gen or manual --commit -m on worktree; optional post-success: --sync, --tag-next, --push, --propagate-tags, --reinstall-local from main)
  --done --no-in-module-replace   block --done on ANY local replace (strict)
  --list                          list worktrees (git worktree list)
  --status                        show status for git repos under this checkout
  --repos                         list git repos under this checkout
  --unwind [--gen-commit-msg --commit …] [--done|--merge-back] [--sync] [--tag-next] [--push] [--reinstall-local] [--dry-run]
                                  plan free-first peel order over the checkout stack DAG
                                  (linked peels: optional generated commit → land → sync → tag/push → pin)
  --unwind --show-graph [--json] [--color|--no-color]
                                  read-only: print repo + module stack graph and peel order
                                  (mutually exclusive with --dry-run and apply/land/pin partners)
  --projects                      list recorded main repository paths
  --projects-dep-graph            module-level dep graph across registered projects
  --scan-git-repos [ROOT...]      list valid git repos under roots (print-only; never mutates projects.json; default: ~)
  --no-cache                      with --scan-git-repos: disable scan cache read/write
  --include-worktrees             with --scan-git-repos: also list linked worktrees
  --github                        with --projects: only show projects whose origin is github.com
  --fetch                         with --projects or --status: fetch upstream before Remote: compare
  -v, --verbose                   log major git commands and go mod tidy to stderr
  --add <dir>                     manually record a main repository path
  --rm <dir>                      remove a recorded main repository path
  --where <basename>              look up saved project path(s) by basename (positional; also: wrk <basename> --where)
                                  (with --main: print main repo abs path of this checkout; no basename)
  --cd <path|basename>            jump into directory (in-place follow-up or interactive shell)
                                  (with --main: jump to main repo of this checkout; no path)
  --main                          open nested shell at main repository root for this checkout
                                  (with --status: run status against the main repo instead;
                                   with --where: print main path; with --cd: runCd to main;
                                   with --reinstall-local: reinstall from main repo modules;
                                   with pipeline stages: run activeRoot as main, no nested shell)
  --bring <path>                  spawn a dependency worktree under ./external (repeatable: --bring p1 --bring p2); soft-skip go.mod replace when not a module dep
  --no-dep                        with --bring: worktree only; skip replace and tidy
  --reinstall-local [--dry-run]   reinstall local module binaries already in GOBIN/GOPATH/bin
                                  (with --main: scan main repository modules for this checkout;
                                   also: after successful --done / --merge-back, scan main tip)
  --tag-next [--dry-run] [--push] [--json]  plan/apply per-scope release tags
                                  (also: after successful --done / --merge-back; --json only bare)
                                  (also: with --propagate-tags: tag then bump consumers)
  --propagate-tags [--dry-run]    plan consumer go.mod bumps from source release tags
                                  (also: after --tag-next / --done / --merge-back;
                                  compose dry-run uses planned next tags when with --tag-next)
  --sync [--dry-run]              FF-only bi-directional sync main ↔ linked worktrees
                                  (also: after successful --done / --merge-back)
  --dry-run                       with --done/--merge-back/--tag-next/--propagate-tags/--sync/--push/--reinstall-local/--unwind/--gen-commit-msg/--commit -m: plan only
  --push                          push current checkout branch to upstream/origin;
                                  with --done/--merge-back: push main branch (and tags when with --tag-next);
                                  with --tag-next: also push newly created tags (branch + tags);
                                  with --pr: full-push branch tip then run PR path
  -f, --force                     with --push: force-with-lease branch push (tags stay non-force);
                                  only valid with --push
  --pr                            show open GitHub PR URL for current linked-worktree branch (or empty);
                                  with --status: print open PR metadata + checks/reviews rollup (or empty);
                                  with --comment only: add comment to existing open PR (error if none);
                                  with --push (no --title): full-push tip when open PR exists (error if none);
                                  with --title and --comment: create or attach a PR (requires gh);
                                  ensures remote head branch on create/attach (push only if missing);
                                  new PR: title + comment as initial body; existing: title ignored, comment added;
                                  compose: [--gen-commit-msg --commit … | --commit -m MSG] [--push] --pr --title T --comment C
                                  (order: commit → push → pr; with --push full-pushes tip before PR path)
  --title <title>                 with --pr create/attach: PR title (required, non-empty; only used on create)
  --comment <body>                with --pr: comment-only body, or create/attach initial body / additive comment
  --json                          with bare --tag-next or --unwind --show-graph: machine-readable stdout
                                  (not valid with --propagate-tags)
  --task <desc>                   append task slug to worktree/branch names
  --set-task <desc>               rename worktree/branch to match new task
  -y, --yes                       auto-confirm Y/n prompts (compat; default already auto-yes for
                                  --done/--merge-back/--set-task including cascade)
  --confirm                       force interactive Y/n for --done/--merge-back/--set-task
                                  (opt out of default auto-yes; use with --confirm-from-stdin on non-TTY)
  --no-cd                         do not write shell follow-up cd lines (for bash auto-cd wrapper)
  --force-cd                      always land in dest after create/--done/--set-task (bypass gates)
  --new-window                    create Mission Control Desktop (implies --new-terminal)
  --new-terminal                  open iTerm2 in a new window at the worktree
  --reuse-terminal                reuse current iTerm2 session when possible
  --smart-terminal                smart iTerm2 window/tab reuse
  --open-in-agent                 launch agent-run after create (iTerm follow-up or current process)
  --no-new-window                 disable window UX for this run
  --no-new-terminal               disable terminal UX for this run
  --no-open-in-agent              disable agent UX for this run
  --no-config                     do not read $WRK_HOME/config.json for this run
  --exec <cmd> [args...]          after success, run command in the mode target directory
  --commit -m, --message MSG      commit staged changes with MSG (manual; no AI); requires --commit;
                                  exclusive with --gen-commit-msg; optional --no-verify, --add-all, --dry-run;
                                  also: pre-stage before --done / --merge-back / --pr / pipeline partners
  --no-verify                     with --commit: skip git commit hooks
  --add-all                       with --commit: stage all (git add -A) before commit;
                                  with --unwind: allow cascade pin when go.mod/go.sum dirty
  --gen-commit-msg [--dir DIR] [--model MODEL] [--agent-runner RUNNER]
                  [--agent-runner-binary PATH] [--commit] [--no-verify] [--dry-run]
                                  generate a commit message for staged changes (AI);
                                  exclusive with -m/--message; also: pre-stage before --done /
                                  --merge-back / --pr (requires --commit with primary; --dir not valid when composed)

  --web                           start local web UI (React SPA + API on 127.0.0.1)
  --port PORT                     listen port for --web only (default: free port from 8080)
  --dev                           with --web: proxy UI to Vite (wrk-react/) for HMR
  --version                       print version and exit
  --help, -h                      show this help and exit

Config management:
  wrk --set-config --create [flags]  merge create UX defaults into $WRK_HOME/config.json
  wrk --set-config --show            pretty-print effective config.json

Skill commands:
  wrk skill --list|-l             list available skills (wrk)
  wrk skill --show [--header]     print wrk SKILL.md (full or YAML header only)
  wrk skill --install [flags]     install wrk SKILL.md to agent directories

Environment:
  WRK_HOME              worktree storage root (default: ~/.wrk)
  WRK_DATE              override the run date (YYYY-MM-DD) used in worktree/branch names
`
}

func runProjects(wrkHome string, colorEnabled bool, fetchEnabled bool, githubOnly bool) error {
	endPerf := beginProjectsPerfRun()
	defer endPerf()

	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return err
	}
	if githubOnly {
		paths = filterGitHubProjectPaths(paths)
	}
	if len(paths) == 0 {
		return nil
	}

	results := make([]projectStatusData, len(paths))
	done := make([]bool, len(paths))

	nextPrint := 0
	printedAny := false
	var mu sync.Mutex
	flush := func() {
		for nextPrint < len(paths) && done[nextPrint] {
			if printedAny {
				fmt.Println()
			}
			printProjectStatusFromData(results[nextPrint], colorEnabled)
			printedAny = true
			nextPrint++
		}
	}

	workers := minInt(projectsProjectWorkers(), len(paths))
	jobs := make(chan int, len(paths))
	for i := range paths {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				p := paths[i]
				endProject := beginProjectPerf(p)
				data, _ := gatherProjectStatus(p, colorEnabled, fetchEnabled)
				endProject()

				mu.Lock()
				results[i] = data
				done[i] = true
				flush()
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return nil
}

// envTruthy reports whether s is a truthy env value: 1, true, or yes
// (case-insensitive). Empty and other values are false.
func envTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

// runScanGitRepos discovers git repositories under roots and always prints each
// valid path once (discovery order via OnRepo). It never reads or writes
// projects.json — print-only forever. By default only RepoTypeMain is emitted;
// includeWorktrees also lists linked worktrees.
//
// wrkHome is unused for registry mutation (kept for call-site symmetry with
// other modes). Two-base mapping (P5): each CLI root under $HOME maps to
// universe "home" (shared product cache under $HOME/.cache/git-repo-scan/home/);
// roots outside home map to universe "root". Scan passes the user-provided
// Roots so the library only discovers under those paths while still
// loading/merging the home universe index. Emit is filtered to paths under the
// CLI roots. Empty CacheRoot uses the scan_repo product default
// (HOME/.cache/git-repo-scan). verbose (from -v/--verbose) and truthy
// WRK_SCAN_DEBUG enable scan_repo Debug plus greppable cache_base + filter
// lines on stderr.
//
// pauseAfterFirstPrint, when > 0, blocks after the first emitted path (select on
// timer or ctx cancel) so interrupt tests can land SIGINT/cancel mid-scan.
// Production passes 0. Supporting tests:
//   - cmd/wrk/tests/scan-git-repos/interrupt/sigint-after-first-path
func runScanGitRepos(wrkHome string, roots []string, noCache bool, includeWorktrees bool, verbose bool, pauseAfterFirstPrint time.Duration) error {
	_ = wrkHome // scan never mutates projects.json under WRK_HOME

	// Resolve home for default root and two-base mapping. Default root requires
	// home to exist as a directory; mapping uses home only when available.
	home, homeErr := userHomeDir()
	if len(roots) == 0 {
		if homeErr != nil || home == "" {
			return fmt.Errorf("wrk: --scan-git-repos requires a home directory to use as default root")
		}
		if st, err := os.Stat(home); err != nil || !st.IsDir() {
			return fmt.Errorf("wrk: --scan-git-repos requires a home directory to use as default root (~ is missing or not a directory)")
		}
		// Reject homes that only exist due to package-init side effects
		// (e.g. agent-pro tool_resolve runs `npm bin -g` which creates $HOME/.npm
		// when npm is on PATH). A real user home always has other content;
		// a never-created FakeHome for tests must not become a false default root.
		if scanHomeUnusableAsDefaultRoot(home) {
			return fmt.Errorf("wrk: --scan-git-repos requires a home directory to use as default root (~ is missing or not a directory)")
		}
		roots = []string{home}
	}

	// Normalize CLI roots for emit filter + debug mapping.
	filterRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		abs, absErr := filepath.Abs(r)
		if absErr != nil {
			abs = r
		}
		filterRoots = append(filterRoots, filepath.Clean(abs))
	}

	debug := verbose || envTruthy(os.Getenv("WRK_SCAN_DEBUG"))

	// P5: greppable two-base mapping debug (cache_base + emit filter per root).
	if debug {
		homeClean := ""
		if homeErr == nil && home != "" {
			homeClean = filepath.Clean(home)
		}
		for _, fr := range filterRoots {
			cacheBase := scanCacheBaseForRoot(homeClean, fr)
			fmt.Fprintf(os.Stderr, "scan: cache_base=%s filter=%s\n", cacheBase, fr)
		}
	}

	// Cancelable scan so Ctrl-C / SIGTERM stops the walk, keeps scan disk
	// cache progress when applicable, and exits 130 with a warning. Does not
	// touch projects.json.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Opt-in async warm polish: WRK_SCAN_REFRESH_ASYNC=1|true|yes.
	// Serve/OnRepo results freeze at warm serve; durable cache may continue
	// until Join (min WarmRefreshBudget, then while process still joining).
	asyncRefresh := envTruthy(os.Getenv("WRK_SCAN_REFRESH_ASYNC"))
	mode := scan_repo.WarmRefreshSync
	if asyncRefresh && !noCache {
		mode = scan_repo.WarmRefreshAsync
	}

	// CacheRoot left empty → product default under $HOME when cache enabled.
	printed := make(map[string]bool)
	// emitPath prints path once and optionally pauses after the first emit so
	// interrupt tests (ScanTestPauseAfterFirstPrint) can cancel mid-scan.
	// Owned by runScanGitRepos — not a test-injected OnRepo.
	emitPath := func(path string) error {
		if printed[path] {
			return nil
		}
		fmt.Println(path)
		first := len(printed) == 0
		printed[path] = true
		if first && pauseAfterFirstPrint > 0 {
			select {
			case <-time.After(pauseAfterFirstPrint):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	onRepo := func(repo scan_repo.Repo) error {
		isMain := repo.RepoType == scan_repo.RepoTypeMain
		isWorktree := repo.RepoType == scan_repo.RepoTypeWorktree
		if !isMain && !(includeWorktrees && isWorktree) {
			return nil
		}
		if repo.Error != "" {
			return nil
		}
		path := storage.NormalizePath(repo.Path)
		// Emit filter: only print paths under CLI-provided roots.
		if !pathUnderAnyRoot(path, filterRoots) {
			return nil
		}
		return emitPath(path)
	}

	// Explicit CacheRoot from resolved home so Capture FakeHome (captureUserHome)
	// does not need process Setenv("HOME") for scan_repo's default cache path.
	cacheRoot := ""
	if !noCache && homeErr == nil && home != "" {
		cacheRoot = filepath.Join(home, ".cache", "git-repo-scan")
	}
	opts := scan_repo.Options{
		Roots:           filterRoots,
		NoCache:         noCache,
		CacheRoot:       cacheRoot,
		Debug:           debug,
		Stderr:          os.Stderr,
		OnRepo:          onRepo,
		WarmRefreshMode: mode,
	}

	var result scan_repo.Result
	var scanErr error
	if mode == scan_repo.WarmRefreshAsync {
		var sess scan_repo.Session
		sess, scanErr = scan_repo.ScanSession(ctx, opts)
		result = sess.Result
		// Always Join: min-budget wait if main finishes early. Join(ctx) so
		// SIGINT cancels the budget wait (Stop + flush already-written index).
		if sess.Refresh != nil {
			if joinErr := sess.Join(ctx); joinErr != nil && scanErr == nil && !errors.Is(joinErr, context.Canceled) {
				// Polish errors are best-effort; surface only when scan ok and debug.
				if debug {
					fmt.Fprintf(os.Stderr, "warning: async refresh: %v\n", joinErr)
				}
			}
		}
	} else {
		result, scanErr = scan_repo.Scan(ctx, opts)
	}
	if scanErr != nil {
		if errors.Is(scanErr, context.Canceled) {
			fmt.Fprintln(os.Stderr, "warning: scan interrupted; cache progress may be saved (projects.json unchanged)")
			return ExitCodeError{Code: 130}
		}
		return scanErr
	}
	// Interrupt during async Join (min-budget wait) after successful serve.
	if errors.Is(ctx.Err(), context.Canceled) {
		fmt.Fprintln(os.Stderr, "warning: scan interrupted; cache progress may be saved (projects.json unchanged)")
		return ExitCodeError{Code: 130}
	}
	if debug {
		fmt.Fprintf(os.Stderr, "scan: printed=%d\n", len(printed))
		if asyncRefresh {
			fmt.Fprintf(os.Stderr, "scan: refresh_mode=async\n")
		}
	}
	for _, re := range result.RootErrors {
		fmt.Fprintf(os.Stderr, "warning: scan root %s: %s\n", re.Root, re.Error)
	}
	return nil
}

// scanCacheBaseForRoot maps a CLI root to the cache universe base name:
// under $HOME (including HOME itself) → "home"; otherwise → "root".
// Empty homeClean means home is unknown → treat as "root".
func scanCacheBaseForRoot(homeClean, absRoot string) string {
	if homeClean == "" {
		return "root"
	}
	absRoot = filepath.Clean(absRoot)
	if absRoot == homeClean || pathIsUnder(absRoot, homeClean) {
		return "home"
	}
	return "root"
}

// scanHomeUnusableAsDefaultRoot reports whether home is not a real user home for
// bare --scan-git-repos defaults: empty, or only package-manager side-effect
// directories created when a missing $HOME is first touched (e.g. npm → .npm).
func scanHomeUnusableAsDefaultRoot(home string) bool {
	entries, err := os.ReadDir(home)
	if err != nil {
		return true
	}
	if len(entries) == 0 {
		return true
	}
	for _, e := range entries {
		switch e.Name() {
		case ".npm", ".cache", ".config", ".local", ".node_repl_history":
			// Side effects from npm/node/tool_resolve init — ignore.
			continue
		default:
			return false
		}
	}
	return true
}

// pathUnderAnyRoot reports whether path is absRoot or a descendant of any root.
func pathUnderAnyRoot(path string, roots []string) bool {
	path = filepath.Clean(path)
	for _, root := range roots {
		root = filepath.Clean(root)
		if path == root {
			return true
		}
		if pathIsUnder(path, root) {
			return true
		}
	}
	return false
}

func runAdd(wrkHome, addDir string) error {
	abs, err := filepath.Abs(addDir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wrk: %s does not exist", abs)
		}
		return fmt.Errorf("stat dir: %w", err)
	}
	if !worktree.IsInsideWorkTree(abs) {
		return fmt.Errorf("%s is not a git repository", abs)
	}
	top, err := worktree.ShowToplevel(abs)
	if err != nil {
		return err
	}
	mainRepo, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return err
	}
	mainRepo = storage.NormalizePath(mainRepo)
	if err := storage.RecordProject(wrkHome, mainRepo, storage.SourceManual); err != nil {
		return err
	}
	fmt.Println(mainRepo)
	return nil
}

func runRemove(wrkHome, removeDir string) error {
	abs, err := filepath.Abs(removeDir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	mainRepoPath := storage.NormalizePath(abs)
	if _, err := os.Stat(abs); err == nil {
		if worktree.IsInsideWorkTree(abs) {
			top, err := worktree.ShowToplevel(abs)
			if err != nil {
				return err
			}
			mainRepo, err := worktree.ResolveMainRepo(top)
			if err != nil {
				return err
			}
			mainRepoPath = storage.NormalizePath(mainRepo)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat dir: %w", err)
	}
	removed, err := storage.RemoveProject(wrkHome, mainRepoPath)
	if err != nil {
		return err
	}
	if removed {
		fmt.Println(mainRepoPath)
	}
	return nil
}

func runList(workDir string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	cmd := gitCommand("-C", cwd, "worktree", "list")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return err
	}
	outStr := string(out)
	if len(outStr) > 0 && !strings.HasSuffix(outStr, "\n") {
		outStr += "\n"
	}
	fmt.Print(outStr)
	return nil
}

// requireLinkedWorktree ensures activeRoot is a linked worktree for --done/--merge-back.
// Error names the gated flag and mentions linked worktree.
func requireLinkedWorktree(workDir, flag string) (checkoutRoot string, err error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return "", fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err = worktree.ShowToplevel(cwd)
	if err != nil {
		return "", err
	}
	if !worktree.IsLinked(checkoutRoot) {
		return "", fmt.Errorf("wrk: %s requires a linked worktree (%s is not a linked worktree)", flag, checkoutRoot)
	}
	return checkoutRoot, nil
}

// requireMainActiveRoot ensures activeRoot is the main repository checkout for --tag-next.
// Error names --tag-next and mentions main.
func requireMainActiveRoot(workDir, flag string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	if worktree.IsLinked(checkoutRoot) {
		return fmt.Errorf("wrk: %s requires the main repository checkout (activeRoot is a linked worktree, not main)", flag)
	}
	return nil
}

// runActiveRootPipeline runs multi-stage compose without --done/--merge-back.
// activeRoot stays the git toplevel of workDir for the whole run.
// Stage order: gen-commit|manual-commit → sync → tag-next → push → pr → propagate-tags → reinstall-local → exec.
// --tag-next is gated to main activeRoot; other stages OK on linked worktrees.
// When withPush and withPR and title set: full branch push first, then runPR
// (ensure-push is a no-op once remote tip already matches after the push stage).
// When withPush and withPR and title empty: push-existing / push+comment —
// list open PR first, then full tip push, then optional comment + URL.
func runActiveRootPipeline(workDir, wrkHome string, genCommitMsg bool, genCommitArgs []string, manualCommit bool, manualMessage string, noVerify, addAll, withSync, withTagNext, withPush, forcePush, withPR bool, prTitle, prComment string, withPropagateTags, withReinstallLocal, dryRun bool, colorFlag bool, execArgs []string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}
	activeRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	// Gate tag-next early so we do not partially apply prior stages incorrectly
	// when the only illegal stage is tag-next from a linked worktree.
	// Still run gen-commit/sync before tag when legal; for illegal tag-next from WT,
	// fail at the tag stage after earlier stages... Tests for bare tag-next and
	// multi-stage tag-next-from-WT expect no tag created; partial sync is OK if
	// non-zero exit. Prefer fail-fast before any stage when tag-next is requested
	// from a linked worktree so push/tag cannot apply under wrong activeRoot.
	if withTagNext {
		if err := requireMainActiveRoot(activeRoot, "--tag-next"); err != nil {
			return err
		}
	}

	printed := false
	blankBefore := func() {
		if printed {
			fmt.Println()
		}
		printed = true
	}

	if genCommitMsg {
		if err := runGenCommitMsgStage(activeRoot, genCommitArgs, dryRun); err != nil {
			return err
		}
		printed = true
	}
	if manualCommit {
		if err := runManualCommitStage(activeRoot, manualMessage, noVerify, addAll, dryRun); err != nil {
			return err
		}
		printed = true
	}
	if withSync {
		blankBefore()
		if err := runSync(activeRoot, dryRun); err != nil {
			return err
		}
		printed = true
	}

	var createdTags []string
	var tagPlan tagscope.ChangePlan
	if withTagNext {
		blankBefore()
		tagRes, err := runTagNextAtResult(activeRoot, "HEAD", dryRun, false, false)
		if err != nil {
			return err
		}
		createdTags = tagRes.Tags
		tagPlan = tagRes.Plan
		printed = true
	}

	// Push-existing / push+comment: --push --pr without --title.
	// List open PR BEFORE push so a missing open PR leaves origin tip unchanged.
	pushExisting := withPush && withPR && strings.TrimSpace(prTitle) == ""
	if pushExisting {
		blankBefore()
		if err := runPRPushExisting(activeRoot, prComment, dryRun, forcePush, colorFlag); err != nil {
			return err
		}
		printed = true
	} else {
		if withPush {
			blankBefore()
			var tags []string
			if withTagNext {
				tags = createdTags
			}
			if err := runPushMain(activeRoot, dryRun, forcePush, tags); err != nil {
				return err
			}
			printed = true
		}
		// --pr after push (or after gen-commit when --push omitted). Ensure-push
		// inside runPR is idempotent when a prior full push already published the tip.
		if withPR {
			blankBefore()
			if err := runPR(activeRoot, prTitle, prComment, colorFlag); err != nil {
				return err
			}
			printed = true
		}
	}
	if withPropagateTags {
		blankBefore()
		var releaseOverride []SourceRelease
		if dryRun && withTagNext {
			releases, err := ResolveSourceReleases(activeRoot)
			if err != nil {
				return err
			}
			releaseOverride = applyPlannedTagsToReleases(releases.Releases, tagPlan)
			if len(releaseOverride) == 0 {
				return fmt.Errorf("wrk: no usable release tags for source modules")
			}
		}
		if err := runPropagateTagsWithReleases(activeRoot, wrkHome, dryRun, releaseOverride); err != nil {
			return err
		}
		printed = true
	}
	if withReinstallLocal {
		blankBefore()
		// Scan modules under activeRoot (already the checkout root).
		if err := runReinstallLocal(activeRoot, dryRun, false, colorFlag); err != nil {
			return err
		}
		printed = true
	}
	// --exec is last; skip under dry-run (plan-only pipeline).
	if len(execArgs) > 0 && !dryRun {
		_ = printed
		if err := runExecInDir(activeRoot, execArgs); err != nil {
			return err
		}
	}
	return nil
}

func runDone(workDir, wrkHome string, confirmFromStdin, yesFlag, forceConfirm, noInModuleReplace, noCd, forceCd bool, execArgs []string, withSync, withTagNext, withPush, forcePush, withPropagateTags, withReinstallLocal, dryRun bool, colorFlag bool) error {
	// Shell process cwd (inherited from interactive shell), not merely workDir.
	// Used after remove to decide whether auto-cd is needed.
	shellCwd, _ := processCwd()

	// Own merge-back: default auto-yes unless --confirm ( -y still auto-yes).
	// Cascade not-included: default auto-yes does NOT apply; only -y/--yes (D3).
	ownAssumeYes := planAssumeYes(yesFlag, forceConfirm)
	cascadeAssumeYes := yesFlag

	checkoutRoot, err := requireLinkedWorktree(workDir, "--done")
	if err != nil {
		return err
	}
	cwd := checkoutRoot

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	// Nested main under consumer is a hard error (D1). Dry-run still runs
	// preflight (D7). Phase banners only when cascade targets ≥ 1.
	// When WRK_SCAN_REFRESH_ASYNC is set, defer Join so polish can steal time
	// during cascade/merge-back (Result stays serve-frozen for cascade targets).
	cascadeTargets, cascadeRefresh, err := listCascadeLinkedWorktrees(consumerTop, checkoutRoot)
	if err != nil {
		return err
	}
	if cascadeRefresh != nil {
		defer func() {
			_ = cascadeRefresh.Join(context.Background())
		}()
	}
	// D2/D7: all-or-nothing dirty preflight on cascade targets + own before any
	// mutation or successful would: cascade plan.
	if len(cascadeTargets) > 0 {
		if err := preflightCascadeDirty(cascadeTargets, checkoutRoot); err != nil {
			return err
		}
	}
	hasCascade := len(cascadeTargets) > 0
	if hasCascade {
		fmt.Println("==> cascade")
		for _, path := range cascadeTargets {
			if err := mergeBackExternalWorktree(path, wrkHome, confirmFromStdin, cascadeAssumeYes, dryRun); err != nil {
				return err
			}
		}
	}

	// Guard: classify every local filesystem replace under the checkout (main
	// or sub-module). wrk --bring writes replace => ./external/... and
	// --done's cascade removes those external worktrees, so a remaining local
	// replace would dangle — those (extra-repo) block. An intra-repo replace
	// (target exists and shares the consumer's toplevel, e.g. ../../ or ./sub
	// pointing back into the same repo) is stable, so under the default lenient
	// guard it only warns and --done proceeds; --no-in-module-replace makes
	// every local replace block. Scanning every module (not just the nearest
	// go.mod) also catches sub-module replaces a single upward lookup would
	// miss. A checkout with no go.mod yields zero modules → guard is a no-op →
	// MergeBack proceeds (it is pure git).
	if err := blockIfLocalReplace(consumerTop, noInModuleReplace); err != nil {
		return err
	}

	if hasCascade {
		fmt.Println("==> own")
	}
	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     true,
		DryRun:     dryRun,
		TmpDir:     filepath.Join(wrkHome, "worktrees"),
		StashLabel: "wrk-merge-back",
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, ownAssumeYes)
		},
	})
	if err != nil {
		return mapMergeBackSharedError(err, "--done")
	}
	// printDryRun already wrote planned commands (no trailing newline).
	if result.Action == "dry-run" {
		fmt.Println()
	} else {
		fmt.Println(result.Message)
	}
	if result.Action == "aborted" {
		return nil
	}
	// Post-pipeline: sync → tag-next → push → propagate-tags → reinstall-local → exec → land.
	// Dry-run: still print post stages in dry mode; skip exec/land.
	// Real success: apply post stages then exec/land.
	if dryRun {
		if err := runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, forcePush, withPropagateTags, true); err != nil {
			return err
		}
		return runComposeReinstallLocal(result, withReinstallLocal, true, colorFlag)
	}
	if err := runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, forcePush, withPropagateTags, false); err != nil {
		return err
	}
	if err := runComposeReinstallLocal(result, withReinstallLocal, false, colorFlag); err != nil {
		return err
	}
	if err := runExecInDir(result.TargetPath, execArgs); err != nil {
		return err
	}
	if forceCd {
		// --force-cd bypasses cwd-missing and foreign-repo ancestor gates.
		if err := forceLandInDir(result.TargetPath); err != nil {
			return err
		}
	} else if err := writeFollowupCDAfterDoneRemove(noCd, shellCwd, result.TargetPath); err != nil {
		return err
	}
	return nil
}

func runMergeBack(workDir, wrkHome string, confirmFromStdin, assumeYes, withSync, withTagNext, withPush, forcePush, withPropagateTags, withReinstallLocal, dryRun bool, colorFlag bool) error {
	checkoutRoot, err := requireLinkedWorktree(workDir, "--merge-back")
	if err != nil {
		return err
	}

	// Land core via workops (Remove=false). Sync/tag-next/push compose stays
	// in CLI so stage printing and dry-run tip pretends remain correct.
	// workops Sync is not used here (would mute compose stdout).
	mb, err := workops.MergeBackFull(context.Background(), workops.MergeBackOptions{
		WorktreeDir: checkoutRoot,
		Sync:        false,
		DryRun:      dryRun,
		WrkHome:     wrkHome,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		return mapMergeBackSharedError(err, "--merge-back")
	}
	result := worktreeMergeBackResultFromOps(mb)
	// printDryRun already wrote planned commands (no trailing newline).
	if result.Action == "dry-run" {
		fmt.Println()
	} else {
		fmt.Println(result.Message)
	}
	if result.Action == "aborted" {
		return nil
	}
	// Post-pipeline same order as runDone (no exec/land). Worktree kept.
	if err := runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, forcePush, withPropagateTags, dryRun); err != nil {
		return err
	}
	return runComposeReinstallLocal(result, withReinstallLocal, dryRun, colorFlag)
}

// worktreeMergeBackResultFromOps adapts workops.MergeBackResult for existing
// compose helpers typed on *worktree.MergeBackResult.
func worktreeMergeBackResultFromOps(mb *workops.MergeBackResult) *worktree.MergeBackResult {
	if mb == nil {
		return &worktree.MergeBackResult{}
	}
	return &worktree.MergeBackResult{
		SourcePath: mb.SourcePath,
		TargetPath: mb.TargetPath,
		Branch:     mb.Branch,
		Relation:   mb.Relation,
		Action:     mb.Action,
		Message:    mb.Message,
	}
}

// runComposeReinstallLocal runs the optional post-merge reinstall tail from main
// (result.TargetPath). useMain=true so the scan is main-repo modules after merge,
// not a removed worktree. Blank line before the stage when other stages may have
// printed. Empty / skip-only plans exit 0 (do not fail the ship).
func runComposeReinstallLocal(result *worktree.MergeBackResult, withReinstallLocal, dryRun bool, colorFlag bool) error {
	if !withReinstallLocal {
		return nil
	}
	mainPath := result.TargetPath
	if mainPath == "" {
		return fmt.Errorf("wrk: merge-back result missing target path")
	}
	fmt.Println() // blank line before reinstall stage
	// Scan main tip after merge (useMain equivalent from main path).
	err := runReinstallLocal(mainPath, dryRun, true, colorFlag)
	if err == nil {
		return nil
	}
	// Empty / non-module main: do not fail the ship (dashboard / pure-git fixtures).
	if strings.Contains(err.Error(), "no go.mod modules found") ||
		strings.Contains(err.Error(), "no go.mod found") {
		if dryRun {
			fmt.Fprintf(os.Stderr, "would: skip reinstall-local (%s)\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "skip reinstall-local: %s\n", err.Error())
		}
		return nil
	}
	return err
}

// runComposePostStages runs optional post-merge stages in fixed order:
//
//	sync → tag-next (local create or plan) → push (branch + tags) → propagate-tags.
//
// When dryRun is true, stages plan only against the would-be main tip after the
// planned primary merge (source worktree HEAD for ahead/FF cases). When dry-run
// includes --tag-next and --propagate-tags, planned next tags are threaded into
// the propagate plan (same as bare --tag-next --propagate-tags --dry-run).
// Blank line between major stdout stages.
func runComposePostStages(result *worktree.MergeBackResult, sourcePath, wrkHome string, withSync, withTagNext, withPush, forcePush, withPropagateTags, dryRun bool) error {
	if !withSync && !withTagNext && !withPush && !withPropagateTags {
		return nil
	}
	mainPath := result.TargetPath
	if mainPath == "" {
		return fmt.Errorf("wrk: merge-back result missing target path")
	}

	// Would-be tip for dry-run tag-next / sync: after planned FF merge of an
	// ahead source, main moves to source HEAD. Already-included / noop leave
	// main tip unchanged.
	headRef := "HEAD"
	var pretendMainAt string
	if dryRun {
		tip, err := resolveWouldBeMainTip(sourcePath, mainPath, result.Relation)
		if err != nil {
			return err
		}
		headRef = tip
		pretendMainAt = tip
	}

	var createdTags []string
	var tagPlan tagscope.ChangePlan
	if withSync {
		fmt.Println() // blank line between primary message and sync block
		if err := runSyncOpts(mainPath, syncOpts{
			DryRun:        dryRun,
			PretendMainAt: pretendMainAt,
		}); err != nil {
			return err
		}
	}
	if withTagNext {
		fmt.Println() // blank line before tag-next block
		// Create tags locally only; push (if any) is via runPushMain with tag list.
		// Dry-run plans against would-be tip; real apply uses main HEAD post-merge.
		// Keep full result so dry-run can thread planned next tags into propagate.
		tagRes, err := runTagNextAtResult(mainPath, headRef, dryRun, false, false)
		if err != nil {
			return err
		}
		createdTags = tagRes.Tags
		tagPlan = tagRes.Plan
	}
	if withPush {
		fmt.Println() // blank line before push confirmation
		var tags []string
		if withTagNext {
			tags = createdTags
		}
		if err := runPushMain(mainPath, dryRun, forcePush, tags); err != nil {
			return err
		}
	}
	if withPropagateTags {
		fmt.Println() // blank line before propagate-tags block
		// Always run from mainPath: after --done the source worktree is gone.
		var releaseOverride []SourceRelease
		if dryRun && withTagNext {
			// Core dry-run contract: plan consumer bumps against planned next
			// versions even though tags do not exist yet (mirror P6 compose).
			releases, err := ResolveSourceReleases(mainPath)
			if err != nil {
				return err
			}
			releaseOverride = applyPlannedTagsToReleases(releases.Releases, tagPlan)
			if len(releaseOverride) == 0 {
				return fmt.Errorf("wrk: no usable release tags for source modules")
			}
		}
		// Apply (or dry-run without tag-next): resolve existing source tags.
		// Apply after tag-next sees newly created tags on main.
		if err := runPropagateTagsWithReleases(mainPath, wrkHome, dryRun, releaseOverride); err != nil {
			return err
		}
	}
	return nil
}

// resolveWouldBeMainTip returns the commit main would land on after the planned
// primary merge: source HEAD for ahead/diverged, else current main HEAD.
func resolveWouldBeMainTip(sourcePath, mainPath, relation string) (string, error) {
	switch relation {
	case "ahead", "diverged":
		out, err := gitOutputDir(sourcePath, "rev-parse", "HEAD")
		if err != nil {
			return "", fmt.Errorf("resolve would-be tip from source: %w", err)
		}
		return strings.TrimSpace(out), nil
	default:
		out, err := gitOutputDir(mainPath, "rev-parse", "HEAD")
		if err != nil {
			return "", fmt.Errorf("resolve would-be tip from main: %w", err)
		}
		return strings.TrimSpace(out), nil
	}
}

// planAssumeYes returns whether merge-back / set-task plan prompts should be
// skipped. Default is auto-yes; --confirm forces prompts; -y/--yes still auto-yes.
func planAssumeYes(assumeYes, forceConfirm bool) bool {
	return assumeYes || !forceConfirm
}

// listCascadeLinkedWorktrees returns linked worktrees under consumerTop that
// are cascade merge-back targets (excludes the consumer checkout itself).
// Nested main repos under the consumer tree are a hard error (D1) — not
// warn+skip — so --done aborts before cascade/own mutations.
//
// Discovery uses scan_repo.ScanSession with ListWorktrees: true and
// Roots: [consumerTop] (scan_repo owns base-path filtering). Targets are
// collected from top-level RepoTypeWorktree rows and from each repo's inner
// Worktrees field.
//
// When WRK_SCAN_REFRESH_ASYNC is truthy, warm polish may run in the background;
// the returned RefreshHandle must be Joined by the caller (after cascade work
// so polish can use that wall time). Result used for targets is serve-frozen.
func listCascadeLinkedWorktrees(consumerTop, checkoutRoot string) ([]string, *scan_repo.RefreshHandle, error) {
	mode := scan_repo.WarmRefreshSync
	if envTruthy(os.Getenv("WRK_SCAN_REFRESH_ASYNC")) {
		mode = scan_repo.WarmRefreshAsync
	}
	sess, err := scan_repo.ScanSession(context.Background(), scan_repo.Options{
		Roots:           []string{consumerTop},
		ListWorktrees:   true,
		WarmRefreshMode: mode,
	})
	if err != nil {
		return nil, nil, err
	}
	result := sess.Result

	cleanCheckout := filepath.Clean(checkoutRoot)
	cleanConsumer := filepath.Clean(consumerTop)
	seen := make(map[string]struct{})
	var targets []string
	addLinked := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if clean == cleanCheckout {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		if !worktree.IsLinked(path) {
			return
		}
		seen[clean] = struct{}{}
		targets = append(targets, path)
	}

	for _, repo := range result.Repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			if filepath.Clean(repo.Path) != cleanConsumer {
				if sess.Refresh != nil {
					sess.Stop()
					_ = sess.Join(context.Background())
				}
				return nil, nil, fmt.Errorf("Error: nested main repo under consumer blocks cascade: %s", repo.Path)
			}
			// consumerTop main (if present): not a cascade target itself;
			// collect linked worktrees from the inner Worktrees field.
			for _, wt := range repo.Worktrees {
				if wt.IsMain {
					continue
				}
				addLinked(wt.Path)
			}
			continue
		}
		if repo.RepoType == scan_repo.RepoTypeWorktree {
			addLinked(repo.Path)
		}
		// Inner Worktrees on any remaining row (defensive; normally filled on mains).
		for _, wt := range repo.Worktrees {
			if wt.IsMain {
				continue
			}
			addLinked(wt.Path)
		}
	}
	return targets, sess.Refresh, nil
}

// preflightCascadeDirty fails hard if any cascade target or the own checkout
// has uncommitted changes. Runs before phase headers/mutations so cascade cannot
// remove externals then fail on own dirty (D2), and so dry-run cannot print a
// false would: success plan (D7).
func preflightCascadeDirty(cascadeTargets []string, ownPath string) error {
	for _, path := range cascadeTargets {
		if err := worktree.IsClean(path); err != nil {
			return fmt.Errorf("Error: %w", err)
		}
	}
	if err := worktree.IsClean(ownPath); err != nil {
		return fmt.Errorf("Error: %w", err)
	}
	return nil
}

// mergeBackExternalWorktree merge-backs (or removes) an external dependency
// worktree during the --done cascade.
//
// External dep worktrees are worktrees of the DEP repo (registered under
// <depMain>/.git/worktrees/, per createExternalWorktree's git -C depMain worktree
// add), so MergeBack resolves the owning main repo from the worktree's .git
// gitdir (the dep main) and merges the dep branch back into the dep repo — the
// branch shares the dep's history, so the merge-base check resolves. This
// ensures dep work committed on the external worktree is merged back into the
// dep repo before the worktree is removed. Relation to dep main: already-included
// → remove only (D8, no confirm); ahead/diverged → Confirm. Cascade assumeYes is
// only true with -y (D3: default auto-yes does not apply to cascade not-included).
// --confirm-from-stdin for non-TTY when prompting (D4).
//
// When dryRun is true (and preflight already passed), prints a compact plan line
// and does not mutate (D6). Real success prints result.Message on stdout (D5).
func mergeBackExternalWorktree(externalPath, wrkHome string, confirmFromStdin, assumeYes, dryRun bool) error {
	if dryRun {
		fmt.Printf("would: cascade merge-back %s\n", externalPath)
		return nil
	}
	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: externalPath,
		TargetPath: "",
		Remove:     true,
		TmpDir:     filepath.Join(wrkHome, "worktrees"),
		StashLabel: "wrk-merge-back",
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		// Structured framing: main prints err.Error() to stderr. Prefix Error:
		// and include cascade worktree path so failures are not bare git detail
		// (e.g. "rebase conflict:") alone.
		return fmt.Errorf("Error: cascade merge-back %s: %w", externalPath, err)
	}
	// D5: print MergeBack Message on stdout (same as own path).
	if result != nil && result.Message != "" {
		fmt.Println(result.Message)
	}
	if result != nil && result.Action == "aborted" {
		// Stop cascade + own after user decline; non-zero so callers do not
		// treat partial --done as success.
		return fmt.Errorf("merge-back aborted")
	}
	return nil
}

// runBring materializes external dependency worktree(s) under ./external and
// best-effort applies go.mod replace (soft SKIP notices on stderr, exit 0).
// bringArgs comes from repeatable --bring (one or more paths).
// When noDep is true: still materialize/reuse external worktree, but skip all
// replace/tidy and module analyse/match (no SKIP messages).
//
// Multi-bring (len>1): left→right; print each abs path before the next; soft
// SKIP continues; hard errors fail-fast (earlier external worktrees kept).
// go mod tidy runs once per consumer module after all replaces so mid-loop tidy
// cannot drop still-pending requires. Single-bring tidies immediately after
// replaces (unchanged).
//
// Preflight resolves each --bring arg once (basename confirm at most once per
// arg). Apply reuses the cached absolute path — no second resolveDirArg prompt.
//
// Non-git consumer cwd: uses abs(cwd) as the external parent, skips
// ensureGitignoreExternal, and soft-skips replace/tidy with
// "SKIP local dep replacement: <abs-cwd> is not a git repository" (unless noDep).
func runBring(workDir string, bringArgs []string, wrkHome string, rawArgs []string, execArgs []string, noDep bool) error {
	if len(bringArgs) == 0 {
		return fmt.Errorf("wrk: --bring requires a path")
	}
	if len(bringArgs) > 1 && len(execArgs) > 0 {
		return fmt.Errorf("wrk: --exec is only valid with a single --bring path")
	}

	// Resolve each arg once up front (interactive select at most once per arg).
	// Duplicate raw/resolved paths fail before any external worktree is created.
	// Per-arg resolve errors are deferred to the left→right apply loop so prior
	// successes still materialize (multi fail-fast).
	resolved, resolveErrs, err := preflightResolveBringArgs(bringArgs, wrkHome, rawArgs)
	if err != nil {
		return err
	}

	multi := len(bringArgs) > 1
	// Multi-only plan on stderr after full successful preflight, before create.
	// Single-arg skips this (noise). Duplicate/hard preflight err returns above.
	if multi {
		allResolved := true
		for _, e := range resolveErrs {
			if e != nil {
				allResolved = false
				break
			}
		}
		if allResolved {
			fmt.Fprintln(os.Stderr, "will bring:")
			for i, a := range bringArgs {
				fmt.Fprintf(os.Stderr, "  %s → %s\n", a, resolved[i])
			}
		}
	}

	tidyAtEnd := make(map[string]struct{})
	var lastExternal string

	for i := range bringArgs {
		if resolveErrs[i] != nil {
			return resolveErrs[i]
		}
		depPath := resolved[i]
		if depPath == "" {
			// Args after a preflight resolve failure are left empty; fail-fast
			// returns at the first resolveErrs[i] above before reaching them.
			return fmt.Errorf("wrk: internal: unresolved --bring path: %s", bringArgs[i])
		}
		externalPath, replacedDirs, err := bringOneFromResolved(workDir, depPath, noDep, !multi)
		if err != nil {
			return err
		}
		for _, d := range replacedDirs {
			tidyAtEnd[d] = struct{}{}
		}
		// Print each success before attempting the next (fail-fast still shows first path).
		absPath, err := filepath.Abs(externalPath)
		if err != nil {
			return fmt.Errorf("resolve external worktree path: %w", err)
		}
		fmt.Println(absPath)
		lastExternal = absPath
	}

	if multi && !noDep && len(tidyAtEnd) > 0 {
		dirs := make([]string, 0, len(tidyAtEnd))
		for d := range tidyAtEnd {
			dirs = append(dirs, d)
		}
		sort.Strings(dirs)
		for _, d := range dirs {
			if err := goModTidyForBring(d); err != nil {
				return err
			}
		}
	}

	// --exec only valid with single bring (rejected above for multi).
	return runExecInDir(lastExternal, execArgs)
}

// preflightResolveBringArgs resolves each --bring arg once and rejects
// duplicates. On success, resolved[i] is the cleaned absolute path.
// When resolve fails for an arg, resolveErrs[i] is set and later args are not
// resolved (caller applies left→right and returns that error after prior
// successes). A non-nil err is only for duplicate detection failures.
func preflightResolveBringArgs(bringArgs []string, wrkHome string, rawArgs []string) (resolved []string, resolveErrs []error, err error) {
	seenArg := make(map[string]struct{}, len(bringArgs))
	for _, a := range bringArgs {
		if _, ok := seenArg[a]; ok {
			return nil, nil, fmt.Errorf("wrk: duplicate --bring path: %s", a)
		}
		seenArg[a] = struct{}{}
	}

	resolved = make([]string, len(bringArgs))
	resolveErrs = make([]error, len(bringArgs))
	seenResolved := make(map[string]struct{}, len(bringArgs))
	hint := &DirHintOptions{RawArgs: rawArgs, DepMode: true}

	for i, a := range bringArgs {
		p, resErr := resolveDirArg(a, true, wrkHome, hint)
		if resErr != nil {
			// Defer hard resolve errors to the ordered bring loop.
			resolveErrs[i] = resErr
			return resolved, resolveErrs, nil
		}
		abs, absErr := filepath.Abs(p)
		if absErr != nil {
			resolveErrs[i] = absErr
			return resolved, resolveErrs, nil
		}
		abs = filepath.Clean(abs)
		if _, ok := seenResolved[abs]; ok {
			return nil, nil, fmt.Errorf("wrk: duplicate --bring path: %s", abs)
		}
		seenResolved[abs] = struct{}{}
		resolved[i] = abs
	}
	return resolved, resolveErrs, nil
}

// bringOneFromResolved materializes one dependency worktree from an already-
// resolved absolute dep path and optionally applies replaces.
// When tidyNow is true, runs go mod tidy after each successful replace (single-bring).
// When tidyNow is false (multi mid-loop), returns consumer dirs that need tidy later.
// Does not call resolveDirArg — preflight already resolved and confirmed basenames.
func bringOneFromResolved(workDir, depPath string, noDep, tidyNow bool) (externalPath string, replacedDirs []string, err error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return "", nil, fmt.Errorf("resolve cwd: %w", err)
	}

	insideWorkTree := worktree.IsInsideWorkTree(cwd)

	var consumerTop string
	if insideWorkTree {
		consumerTop, err = worktree.ShowToplevel(cwd)
		if err != nil {
			return "", nil, err
		}
	} else {
		// --bring from non-git cwd: parent for external/ is abs(cwd).
		consumerTop = cwd
	}
	if _, err := worktree.ResolveMainRepo(depPath); err != nil {
		return "", nil, err
	}

	// Create external worktree (+ /external gitignore only when consumer is git).
	externalPath, err = createExternalWorktreeForRepo(consumerTop, depPath)
	if err != nil {
		return "", nil, err
	}

	// --no-dep: worktree only; skip replace + tidy (and analyse/match).
	if noDep {
		return externalPath, nil, nil
	}

	// Non-git consumer: soft-skip replace/tidy without module analyse.
	if !insideWorkTree {
		fmt.Fprintf(os.Stderr, "SKIP local dep replacement: %s is not a git repository\n", cwd)
		return externalPath, nil, nil
	}

	// Soft-skip replace with notices; worktree already exists.
	depModules, err := scan.Scan(depPath, scan.Options{})
	if err != nil {
		return "", nil, fmt.Errorf("scan dep modules: %w", err)
	}
	if len(depModules) == 0 {
		fmt.Fprintf(os.Stderr, "SKIP local dep replacement: %s is not a go module\n", depPath)
		return externalPath, nil, nil
	}

	consumerModules, err := scan.Scan(consumerTop, scan.Options{})
	if err != nil {
		return "", nil, fmt.Errorf("scan consumer modules: %w", err)
	}
	if len(consumerModules) == 0 {
		fmt.Fprintf(os.Stderr, "SKIP local dep replacement: consumer has no Go modules\n")
		return externalPath, nil, nil
	}

	matchingConsumerDirs, depModDir := matchDepToConsumerModules(consumerTop, consumerModules, depModules)
	if len(matchingConsumerDirs) == 0 {
		fmt.Fprintf(os.Stderr, "SKIP local dep replacement: %s is not a dependency of any consumer module\n", depPath)
		return externalPath, nil, nil
	}

	// The replace must target the directory holding the dep module's go.mod:
	// the repo root when depModDir is ".", or the sub-module subdir otherwise.
	replaceDir := externalPath
	if depModDir != "." {
		replaceDir = filepath.Join(externalPath, depModDir)
	}
	for _, m := range matchingConsumerDirs {
		if _, _, err := replace.ReplaceIn(m.dir, replaceDir); err != nil {
			return "", nil, err
		}
		replacedDirs = append(replacedDirs, m.dir)
		if tidyNow {
			if err := goModTidyForBring(m.dir); err != nil {
				return "", nil, err
			}
		}
	}

	return externalPath, replacedDirs, nil
}

type consumerMatch struct{ dir string }

// matchDepToConsumerModules finds consumer module dirs that require/replace any
// dep module path. depModDir is the relative dir of the first matching dep module.
func matchDepToConsumerModules(consumerTop string, consumerModules []scan.Module, depModules []scan.Module) ([]consumerMatch, string) {
	var matchingConsumerDirs []consumerMatch
	var depModDir string
	for _, cm := range consumerModules {
		wanted := make(map[string]struct{})
		for _, req := range cm.Requires {
			wanted[req.Path] = struct{}{}
		}
		for _, repl := range cm.Replaces {
			wanted[repl.OldPath] = struct{}{}
		}
		for _, dm := range depModules {
			if dm.Path == "" {
				continue
			}
			if _, ok := wanted[dm.Path]; ok {
				matchingConsumerDirs = append(matchingConsumerDirs, consumerMatch{
					dir: filepath.Join(consumerTop, cm.Dir),
				})
				if depModDir == "" {
					depModDir = dm.Dir
				}
			}
		}
	}
	return matchingConsumerDirs, depModDir
}

// resolveDepModule scans the dep repo at depPath for Go modules and returns the
// directory (relative to depPath, "." for a root go.mod) and module path of the
// module the consumer requires. This handles dependency repos whose module
// lives in a subdirectory rather than at the repo root.
//
// It returns:
//   - "not a go module" when the dep repo contains no go.mod at all,
//   - "<depPath> is not a dependency of the consumer module" when none of the
//     discovered modules matches a consumer require/replace path.
func resolveDepModule(consumerModDir, depPath string) (modDir string, modPath string, err error) {
	consumerMod, err := resolve.GetModuleInfo(consumerModDir)
	if err != nil {
		return "", "", fmt.Errorf("read consumer go.mod: %w", err)
	}
	// Module paths the consumer depends on, either via require or replace.
	wanted := make(map[string]struct{})
	for _, req := range consumerMod.Require {
		wanted[req.Path] = struct{}{}
	}
	for _, repl := range consumerMod.Replace {
		wanted[repl.Old.Path] = struct{}{}
	}

	modules, err := scan.Scan(depPath, scan.Options{})
	if err != nil {
		return "", "", fmt.Errorf("scan dep modules: %w", err)
	}
	if len(modules) == 0 {
		return "", "", fmt.Errorf("not a go module: %s", depPath)
	}
	for _, m := range modules {
		if m.Path == "" {
			continue
		}
		if _, ok := wanted[m.Path]; ok {
			return m.Dir, m.Path, nil
		}
	}
	return "", "", fmt.Errorf("%s is not a dependency of the consumer module", depPath)
}

// pathIsUnder reports whether path is strictly inside parent (not parent itself).
func pathIsUnder(path, parent string) bool {
	path = filepath.Clean(path)
	parent = filepath.Clean(parent)
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	if rel == "." || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// readStdinLineForPrompt reads one line from stdin for an interactive Y/n prompt.
// Under `script`(1) with piped StdinInput (doctest UseScriptTTY harness on macOS),
// the first ReadString can observe a spurious empty EOF before the answer bytes
// arrive on the PTY; retry briefly. Persistent empty EOF is treated as empty
// input (default answer).
func readStdinLineForPrompt() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err == nil {
		return line, nil
	}
	if len(line) > 0 {
		// Partial line without trailing newline still usable (e.g. "n" + EOF).
		return line, nil
	}
	if !errors.Is(err, io.EOF) {
		return "", err
	}
	// Spurious empty EOF: retry for up to ~1s.
	for attempt := 0; attempt < 50; attempt++ {
		time.Sleep(20 * time.Millisecond)
		line, err = reader.ReadString('\n')
		if err == nil {
			return line, nil
		}
		if len(line) > 0 {
			return line, nil
		}
		if !errors.Is(err, io.EOF) {
			return "", err
		}
	}
	// No data after retries: empty line → default Y/skip.
	return "\n", nil
}

// findLiveLinkedWorktrees returns absolute paths of live linked worktrees of
// mainRepo, sorted lexicographically. Dead (missing on disk) entries are omitted.
func findLiveLinkedWorktrees(mainRepo string) ([]string, error) {
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		return nil, err
	}
	var paths []string
	seen := make(map[string]bool)
	for _, e := range linked {
		if e.Path == "" {
			continue
		}
		abs, err := filepath.Abs(e.Path)
		if err != nil {
			continue
		}
		abs = filepath.Clean(abs)
		if worktree.IsDead(abs) {
			continue
		}
		if !worktree.IsLinked(abs) {
			continue
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		paths = append(paths, abs)
	}
	sort.Strings(paths)
	return paths, nil
}

// planNamedSpawnPath is a read-only planner for named create intended spawn path.
//   - absTarget exists and is a file → error
//   - absTarget exists and is a linked worktree → intended = absTarget (caller treats as occupied)
//   - absTarget exists as a plain directory (container) → first free
//     {basename}-{token}-{date}[-N] path under it (path occupancy only; branch checked at create)
//   - absTarget missing, parent exists → intended = absTarget
//   - parent missing → error
func planNamedSpawnPath(absTarget, basename, pathToken, date, slug string) (string, error) {
	info, err := os.Stat(absTarget)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("wrk: %s is not a directory", absTarget)
		}
		// Existing linked worktree: treat as exact spawn path already taken (do not nest under it).
		if worktree.IsLinked(absTarget) {
			return absTarget, nil
		}
		// Container dir: first free named subdir under absTarget.
		for suffix := 0; suffix < 100; suffix++ {
			wtPath, _ := candidateNames(absTarget, basename, pathToken, date, slug, suffix)
			if _, serr := os.Stat(wtPath); serr == nil {
				continue
			} else if serr != nil && !os.IsNotExist(serr) {
				return "", fmt.Errorf("stat candidate path: %w", serr)
			}
			return wtPath, nil
		}
		return "", fmt.Errorf("could not find available worktree name after 99 attempts")
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat target dir: %w", err)
	}
	parentDir := filepath.Dir(absTarget)
	if _, perr := os.Stat(parentDir); perr != nil {
		if os.IsNotExist(perr) {
			return "", fmt.Errorf("wrk: %s does not exist", parentDir)
		}
		return "", fmt.Errorf("stat target parent: %w", perr)
	}
	return absTarget, nil
}

// findReusableSiblingWorktrees returns live linked worktrees of mainRepo that are
// direct siblings under spawnParent, porcelain-clean, and at the same HEAD as
// sourceCheckout. Paths are absolute, cleaned, and sorted lexicographically
// (primary = first). Dirty or clean-but-differs siblings are omitted; other-parent
// worktrees are omitted.
func findReusableSiblingWorktrees(mainRepo, sourceCheckout, spawnParent string) ([]string, error) {
	spawnParent = filepath.Clean(spawnParent)
	if abs, err := filepath.Abs(spawnParent); err == nil {
		spawnParent = filepath.Clean(abs)
	}
	all, err := findLiveLinkedWorktrees(mainRepo)
	if err != nil {
		return nil, err
	}
	srcHEAD, err := gitOutputDir(sourceCheckout, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("rev-parse HEAD at source: %w", err)
	}
	srcHEAD = strings.TrimSpace(srcHEAD)
	var reusable []string
	for _, p := range all {
		if filepath.Clean(filepath.Dir(p)) != spawnParent {
			continue
		}
		// Clean: porcelain empty (untracked counts as dirty).
		if err := worktree.IsClean(p); err != nil {
			continue
		}
		head, err := gitOutputDir(p, "rev-parse", "HEAD")
		if err != nil {
			continue
		}
		if strings.TrimSpace(head) != srcHEAD {
			continue
		}
		reusable = append(reusable, p)
	}
	// all is already sorted; reusable preserves that order.
	return reusable, nil
}

// findExistingExternalForDep returns live linked worktrees of depMain whose
// paths lie under {consumerTop}/external/, sorted lexicographically (lex-smallest
// first). Identity is the cleaned absolute path of each worktree.
func findExistingExternalForDep(consumerTop, depMain string) ([]string, error) {
	consumerTop, err := filepath.Abs(consumerTop)
	if err != nil {
		return nil, err
	}
	externalRoot := filepath.Join(consumerTop, "external")
	all, err := findLiveLinkedWorktrees(depMain)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, p := range all {
		if pathIsUnder(p, externalRoot) {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

// warnReuseExternal prints Policy A reuse warnings on stderr for the existing
// external worktree paths (paths must be non-empty and sorted).
func warnReuseExternal(basename string, paths []string) {
	if len(paths) == 0 {
		return
	}
	primary := paths[0]
	if len(paths) == 1 {
		fmt.Fprintf(os.Stderr, "wrk: warning: %s already exists under external/; reusing %s\n", basename, primary)
		return
	}
	fmt.Fprintf(os.Stderr, "wrk: warning: %s already has %d worktrees under external/; reusing %s\n", basename, len(paths), primary)
	for _, p := range paths[1:] {
		fmt.Fprintf(os.Stderr, "wrk: warning: also present: %s\n", p)
	}
}

// planExternalWorktreePath is the read-only planner for an external dep
// worktree: it resolves the dep's main repo, basename, branch base, path token,
// date and consumer main repo, then runs the suffix loop
// (externalCandidateNames + externalCandidateBlocked) to return the first
// non-blocked candidate external worktree path. It performs NO writes: no
// MkdirAll(external/), no ensureGitignoreExternal, no createExternalWorktree.
// It may call read-only git helpers (ShowToplevel, ResolveMainRepo, ReadBranch,
// resolveNamingInputs) which only run git rev-parse / git symbolic-ref.
//
// Policy A: if any live linked worktree of depMain already exists under
// {consumerTop}/external/, returns the lex-smallest path and emits reuse
// warnings on stderr (shared by --bring and dry-run planners).
func planExternalWorktreePath(consumerTop, depPath string) (externalPath string, err error) {
	depSource, err := worktree.ShowToplevel(depPath)
	if err != nil {
		return "", err
	}
	depMain, err := worktree.ResolveMainRepo(depSource)
	if err != nil {
		return "", err
	}

	if existing, err := findExistingExternalForDep(consumerTop, depMain); err != nil {
		return "", err
	} else if len(existing) > 0 {
		warnReuseExternal(filepath.Base(depMain), existing)
		return existing[0], nil
	}

	baseBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return "", err
	}
	basename := filepath.Base(depMain)
	_, pathToken, err := resolveNamingInputs(depPath, baseBranch)
	if err != nil {
		return "", err
	}
	date := resolveWrkDate()

	for suffix := 0; suffix < 100; suffix++ {
		candidatePath, branch := externalCandidateNames(consumerTop, basename, pathToken, date, suffix)
		// Branch-collision check runs against depMain: the external worktree's
		// branch lives in the dep repo (see createExternalWorktree).
		if externalCandidateBlocked(depMain, candidatePath, branch) {
			continue
		}
		return candidatePath, nil
	}
	return "", fmt.Errorf("could not find available external worktree name after 99 attempts")
}

// createExternalWorktreeForRepo materializes the external worktree for the dep
// repo resolved from depPath under {consumerTop}/external/ and returns its path.
// It plans the path via planExternalWorktreePath (so dry-run and real runs
// agree on naming), then creates the external dir, ensures .gitignore when the
// consumer parent is a git repo, and adds the worktree. It does NOT add a
// replace directive or run tidy.
//
// Policy A: when planExternalWorktreePath reuses an existing external path, this
// still ensures /external gitignore (git parents only) but does not create a new
// worktree/branch. Non-git consumerTop (e.g. --bring from a plain dir) skips
// ensureGitignoreExternal entirely.
func createExternalWorktreeForRepo(consumerTop, depPath string) (externalPath string, err error) {
	externalPath, err = planExternalWorktreePath(consumerTop, depPath)
	if err != nil {
		return "", err
	}

	externalDir := filepath.Join(consumerTop, "external")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return "", fmt.Errorf("create external dir: %w", err)
	}
	// Only write /external into .gitignore when the parent is a git repository.
	// --bring from a non-git cwd must not create a .gitignore there.
	if worktree.IsInsideWorkTree(consumerTop) {
		if err := ensureGitignoreExternal(consumerTop); err != nil {
			return "", err
		}
	}

	// Reuse path: already on disk as a live linked worktree — no git worktree add.
	if st, err := os.Stat(externalPath); err == nil && st.IsDir() && worktree.IsLinked(externalPath) {
		return externalPath, nil
	}

	depSource, err := worktree.ShowToplevel(depPath)
	if err != nil {
		return "", err
	}
	depMain, err := worktree.ResolveMainRepo(depSource)
	if err != nil {
		return "", err
	}

	baseBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return "", err
	}
	basename := filepath.Base(depMain)
	_, pathToken, err := resolveNamingInputs(depPath, baseBranch)
	if err != nil {
		return "", err
	}
	date := resolveWrkDate()

	for suffix := 0; suffix < 100; suffix++ {
		candidatePath, branch := externalCandidateNames(consumerTop, basename, pathToken, date, suffix)
		if candidatePath != externalPath {
			// planExternalWorktreePath already selected the first non-blocked
			// candidate; later suffixes are never needed here.
			continue
		}
		if err := createExternalWorktree(depMain, depPath, candidatePath, branch); err != nil {
			return "", err
		}
		break
	}
	return externalPath, nil
}

// createExternalWorktree spawns the external dep worktree as a worktree of the
// DEP repo (not the consumer). The dep already holds its own objects, so no
// remote/fetch into the consumer is needed; the worktree and its branch are
// registered under <depMain>/.git/worktrees/ — where they semantically belong.
// This also lets `wrk --done` cascade merge dep changes back into the dep repo:
// the dep branch shares the dep's history, so merge-base resolves (the previous
// consumer-owned design failed with "failed to find merge base" because the dep
// branch and the consumer's main had no common ancestor).
func createExternalWorktree(depMain, depPath, externalPath, branch string) error {
	depBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return err
	}
	if depBranch == "HEAD" {
		return fmt.Errorf("dep repository is on a detached HEAD")
	}

	// Always create a new branch from the dep's current tip. Callers must have
	// already walked to a free branch name (externalCandidateBlocked).
	// Use runGitWorktreeAdd so -v streams Preparing worktree / HEAD is now at.
	cmd := gitCommand("-C", depMain, "worktree", "add", "-b", branch, externalPath, depBranch)
	return runGitWorktreeAdd(cmd)
}

func ensureGitignoreExternal(top string) error {
	path := filepath.Join(top, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			return nil
		}
	}
	content := string(data)
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "/external\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

// blockIfLocalReplace scans every Go module under top (main + sub-modules) and
// classifies each filesystem/local replace directive. A checkout with no go.mod
// yields zero modules and is allowed: no go.mod means no replace can exist, and
// MergeBack itself is pure git.
//
// A replace is intra-repo when its target resolves to an existing directory
// that shares the consumer's git toplevel (a ../../ or ./sub reference back
// into the same repo); otherwise it is extra-repo (./external dep worktree,
// non-existent target, absolute or sibling-repo path).
//
// Under the default lenient guard, intra-repo replaces only warn (printed to
// stderr) and --done proceeds; extra-repo replaces block. When noInModuleReplace
// is set, every local replace blocks (fully strict).
func blockIfLocalReplace(top string, noInModuleReplace bool) error {
	issues, err := replace.CheckLocalReplaces(top)
	if err != nil {
		return fmt.Errorf("check local replaces under %s: %w", top, err)
	}

	for _, issue := range issues {
		hasExtra := !issue.IsIntraRepo

		if hasExtra || noInModuleReplace {
			var b strings.Builder
			b.WriteString("local filesystem replace blocks wrk --done:\n")
			fmt.Fprintf(&b, "%s\n", replace.FormatIssueLine(top, issue))
			b.WriteString("resolve replace directives manually before running wrk --done")
			return errors.New(b.String())
		}

		// Only intra-repo offenders, default lenient mode: warn and proceed.
		fmt.Fprintln(os.Stderr, replace.FormatIssueLine(top, issue))
		fmt.Fprintln(os.Stderr, "local filesystem replace (intra-repo) - tolerated, remove before pushing:")
	}
	return nil
}

func findGoModDir(cwd, top string) (string, error) {
	dir := cwd
	top = filepath.Clean(top)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Clean(dir) == top {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no go.mod found within %s", top)
}

func externalCandidateNames(consumerTop, basename, pathToken, date string, suffix int) (path, branch string) {
	// Path keeps dep basename; branch is {token}-{date}[-N] with no dep basename
	// prefix (P2). Distinct deps live in separate git repos so same branch name
	// across deps is fine; within one dep, joint path+branch -N via blocked loop.
	name := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	branch = pathToken + "-" + date
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
		branch = fmt.Sprintf("%s-%d", branch, suffix)
	}
	return filepath.Join(consumerTop, "external", name), branch
}

func externalCandidateBlocked(mainRepo, wtPath, branch string) bool {
	if _, err := os.Stat(wtPath); err == nil {
		return true
	}
	return branchExists(mainRepo, branch)
}

func runCreate(workDir string, origWd string, targetDir string, taskDesc string, noCd, forceCd bool, execArgs []string, ux createUXPlan) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	mainRepo, err := worktree.ResolveMainRepo(checkoutRoot)
	if err != nil {
		return err
	}

	baseBranch, err := worktree.ReadBranch(cwd)
	if err != nil {
		return err
	}

	date := resolveWrkDate()
	branchBase, pathToken, err := resolveNamingInputs(cwd, baseBranch)
	if err != nil {
		return err
	}
	basename := filepath.Base(mainRepo)

	// Derive task slug if --task was set (or promoted from a task-like positional).
	// CLI rejects empty/whitespace task text when the flag is present with empty value.
	// Fit slug to path/branch 255-byte budget (reserve 3 for -N); agent still gets full taskDesc.
	var slug string
	if taskDesc != "" {
		if strings.TrimSpace(taskDesc) == "" {
			return fmt.Errorf("wrk: task description must not be empty")
		}
		slug = slugify(taskDesc)
		if slug == "" {
			return fmt.Errorf("wrk: task description %q produces an empty slug", taskDesc)
		}
	}
	fitted, fitErr := fitTaskSlugForNames(basename, pathToken, date, slug)
	if fitErr != nil {
		return fitErr
	}
	slug = fitted

	if targetDir != "" {
		return runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date, slug, noCd, forceCd, execArgs, taskDesc, ux)
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}

	// Window (space) before worktree so a space failure leaves no orphan path.
	if err := ensureCreateWindow(&ux); err != nil {
		return err
	}

	result, err := CreateDefaultWorktree(cwd, wrkHome, slug)
	if err != nil {
		return err
	}
	fmt.Println(result.Path)
	// Pipeline: [window] → create (path printed) → terminal-or-agent → exec → follow-up cd.
	if err := runCreateUX(result.Path, taskDesc, ux); err != nil {
		return err
	}
	if err := runExecInDir(result.Path, execArgs); err != nil {
		return err
	}
	// --force-cd always lands parent; otherwise skip home-gated auto-cd when
	// create already opens agent and/or terminal (parent need not cd).
	if forceCd {
		return forceLandInDir(result.Path)
	}
	if ux.agent || ux.terminalMode != "" {
		return nil
	}
	return writeFollowupCDIfCwdIsHome(noCd, origWd, result.Path)
}

// runCreateTargetDir handles wrk <dir> <target-dir>. A relative <target-dir> is
// resolved against origWd (the process/shell cwd), NOT the repo dir that Run
// chdir'd into.
//
// Policy B (named create, scoped same-parent reuse): after resolving the intended
// spawn path, if that path already exists → error (no create). Else consider only
// live linked worktrees of mainRepo that are direct siblings under
// parent(intendedSpawn), porcelain-clean, and HEAD==source checkout. TTY: prompt
// skip (default Y, wording "would reuse" / "skip creating"); non-TTY: create
// without refuse. No reusable siblings → create as today with no banner.
func runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date, slug string, noCd, forceCd bool, execArgs []string, taskDesc string, ux createUXPlan) error {
	// Resolve <target-dir> against the shell cwd (origWd), not the repo dir.
	// Abs-normalize so parent comparisons match live worktree paths (macOS
	// /var → /private/var).
	absTarget := targetDir
	if !filepath.IsAbs(absTarget) {
		absTarget = filepath.Join(origWd, absTarget)
	}
	if abs, aerr := filepath.Abs(absTarget); aerr == nil {
		absTarget = abs
	}
	absTarget = filepath.Clean(absTarget)

	// Read-only plan of intended spawn path (no create yet).
	intendedSpawn, err := planNamedSpawnPath(absTarget, basename, pathToken, date, slug)
	if err != nil {
		return err
	}
	// Exact path occupied: intended spawn already on disk (dir/file/linked WT).
	if _, err := os.Stat(intendedSpawn); err == nil {
		return fmt.Errorf("wrk: %s already exists", intendedSpawn)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat intended spawn: %w", err)
	}

	// Policy B: reusable same-parent siblings only (not global linked scan).
	spawnParent := filepath.Dir(intendedSpawn)
	if abs, aerr := filepath.Abs(spawnParent); aerr == nil {
		spawnParent = filepath.Clean(abs)
	} else {
		spawnParent = filepath.Clean(spawnParent)
	}
	reusable, err := findReusableSiblingWorktrees(mainRepo, checkoutRoot, spawnParent)
	if err != nil {
		return err
	}
	if len(reusable) > 0 && term.IsTerminal(int(os.Stdin.Fd())) {
		primary := reusable[0]
		// Color on stderr only when interactive terminal and NO_COLOR is unset.
		colorOn := term.IsTerminal(int(os.Stderr.Fd())) && os.Getenv("NO_COLOR") == ""
		warnTok := "warning:"
		if colorOn {
			warnTok = colorize("warning:", ansiOrange)
		}
		pathDisp := func(p string) string {
			if colorOn {
				return colorize(p, ansiGrey)
			}
			return p
		}
		if len(reusable) > 1 {
			fmt.Fprintf(os.Stderr, "wrk: %s %s would reuse %s\n", warnTok, basename, pathDisp(primary))
			for _, p := range reusable[1:] {
				fmt.Fprintf(os.Stderr, "wrk: %s also present: %s\n", warnTok, pathDisp(p))
			}
		}
		// Prompt on stderr; default is skip (Y/empty). No trailing newline before read.
		fmt.Fprintf(os.Stderr, "wrk: %s %s would reuse %s, skip creating another? [Y/n] ", warnTok, basename, pathDisp(primary))
		line, err := readStdinLineForPrompt()
		if err != nil {
			return fmt.Errorf("wrk: read skip confirmation: %w", err)
		}
		answer := strings.TrimSpace(strings.ToLower(line))
		switch answer {
		case "", "y", "yes":
			absPath, err := filepath.Abs(primary)
			if err != nil {
				return fmt.Errorf("resolve worktree path: %w", err)
			}
			fmt.Println(absPath)
			if err := runExecInDir(absPath, execArgs); err != nil {
				return err
			}
			if forceCd {
				return forceLandInDir(absPath)
			}
			return nil
		case "n", "no":
			// Fall through to create as today.
		default:
			return fmt.Errorf("wrk: invalid input %q (expected y/n)", strings.TrimSpace(line))
		}
	}
	// Non-TTY + reusable: create (no refuse). Empty reusable: create with no banner.

	info, err := os.Stat(absTarget)
	if err == nil {
		// Case 2 / file edge: <target-dir> exists.
		if !info.IsDir() {
			return fmt.Errorf("wrk: %s is not a directory", absTarget)
		}
		// Case 2: spawn a default-named sub-dir under <target-dir>, with the
		// usual -N collision handling on both path and branch.
		// Window once before any worktree attempt (same plan for all suffixes).
		if err := ensureCreateWindow(&ux); err != nil {
			return err
		}
		for suffix := 0; suffix < 100; suffix++ {
			wtPath, branch := candidateNames(absTarget, basename, pathToken, date, slug, suffix)
			if candidateBlocked(mainRepo, wtPath, branch) {
				continue
			}
			if err := createWorktree(checkoutRoot, wtPath, branch); err != nil {
				return err
			}
			absPath, err := filepath.Abs(wtPath)
			if err != nil {
				return fmt.Errorf("resolve worktree path: %w", err)
			}
			fmt.Println(absPath)
			if err := runCreateUX(absPath, taskDesc, ux); err != nil {
				return err
			}
			// Target-dir create skips home-gated auto-cd; --force-cd still lands.
			if err := runExecInDir(absPath, execArgs); err != nil {
				return err
			}
			if forceCd {
				return forceLandInDir(absPath)
			}
			return nil
		}
		return fmt.Errorf("could not find available worktree name after 99 attempts")
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat target dir: %w", err)
	}

	// <target-dir> does not exist. Case 1 (parent exists) vs case 3 (parent missing).
	parentDir := filepath.Dir(absTarget)
	if _, perr := os.Stat(parentDir); perr != nil {
		if os.IsNotExist(perr) {
			return fmt.Errorf("wrk: %s does not exist", parentDir)
		}
		return fmt.Errorf("stat target parent: %w", perr)
	}

	// Case 1: spawn the worktree exactly at <target-dir> (fixed path, no naming
	// suffix on the path). Branch follows {token}-{date}[-slug][-N]; if the
	// preferred branch ref already exists, walk -N on the branch only (path stays).
	if err := ensureCreateWindow(&ux); err != nil {
		return err
	}
	wtPath := absTarget
	// branchBase is sanitized (same token as pathToken) via resolveNamingInputs.
	preferredBranch := branchBase + "-" + date
	if slug != "" {
		preferredBranch = preferredBranch + "-" + slug
	}
	for suffix := 0; suffix < 100; suffix++ {
		branch := preferredBranch
		if suffix > 0 {
			branch = fmt.Sprintf("%s-%d", preferredBranch, suffix)
		}
		if branchExists(mainRepo, branch) {
			continue
		}
		if err := createWorktree(checkoutRoot, wtPath, branch); err != nil {
			return err
		}
		absPath, err := filepath.Abs(wtPath)
		if err != nil {
			return fmt.Errorf("resolve worktree path: %w", err)
		}
		fmt.Println(absPath)
		if err := runCreateUX(absPath, taskDesc, ux); err != nil {
			return err
		}
		// Target-dir create skips home-gated auto-cd; --force-cd still lands.
		if err := runExecInDir(absPath, execArgs); err != nil {
			return err
		}
		if forceCd {
			return forceLandInDir(absPath)
		}
		return nil
	}
	return fmt.Errorf("could not find available branch name after 99 attempts")
}

func resolveWrkHome() (string, error) {
	// Capture injects WRK_HOME via captureWrkHome (no process Setenv).
	if captureWrkHome != "" {
		return filepath.Abs(pathfmt.Expand(captureWrkHome))
	}
	if v := os.Getenv("WRK_HOME"); v != "" {
		return filepath.Abs(pathfmt.Expand(v))
	}
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".wrk"), nil
}

func resolveWrkDate() string {
	if captureWrkDate != "" {
		return captureWrkDate
	}
	if v := os.Getenv("WRK_DATE"); v != "" {
		return v
	}
	return time.Now().Format("2006-01-02")
}

func resolveNamingInputs(cwd, baseBranch string) (branchBase, pathToken string, err error) {
	if baseBranch == "HEAD" {
		hash, err := shortHEAD(cwd)
		if err != nil {
			return "", "", err
		}
		return hash, hash, nil
	}
	// Sanitize for both path token and branch segment so branch names never
	// contain '/' (P1: feature/foo → feature-foo).
	token := sanitizeBranchToken(baseBranch)
	return token, token, nil
}

func shortHEAD(repo string) (string, error) {
	cmd := gitCommand("-C", repo, "rev-parse", "--short=7", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --short=7 HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func candidateNames(worktreesDir, basename, pathToken, date, slug string, suffix int) (path, branch string) {
	// pathToken is the sanitized branch segment from resolveNamingInputs.
	// Invariant for wrk-managed paths: Base(path) == basename + "-" + branch.
	name := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	if slug != "" {
		name = fmt.Sprintf("%s-%s", name, slug)
	}
	branch = pathToken + "-" + date
	if slug != "" {
		branch = branch + "-" + slug
	}
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
		branch = fmt.Sprintf("%s-%d", branch, suffix)
	}
	return filepath.Join(worktreesDir, name), branch
}

func candidateBlocked(mainRepo, wtPath, branch string) bool {
	if _, err := os.Stat(wtPath); err == nil {
		return true
	}
	return branchExists(mainRepo, branch)
}

func branchExists(repo, branch string) bool {
	cmd := gitCommand("-C", repo, "rev-parse", "--verify", "refs/heads/"+branch)
	return cmd.Run() == nil
}

// createWorktree always creates a new branch via `git worktree add -b`.
// Callers must ensure the branch name is free (candidateBlocked / fixed-path walk).
func createWorktree(sourceDir, wtPath, branch string) error {
	cmd := gitCommand("-C", sourceDir, "worktree", "add", "-b", branch, wtPath)
	return runGitWorktreeAdd(cmd)
}

// hasArg returns true if args contains the given flag.
func hasArg(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// slugify converts a task description into a path-safe slug.
// Rules: lowercase, non-letter-non-digit -> "-", collapse runs of "-",
// trim leading/trailing "-", truncate to 64 runes.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	runes := []rune(s)
	if len(runes) > 64 {
		s = string(runes[:64])
	}
	s = strings.Trim(s, "-")
	return s
}

// datePattern matches "-YYYY-MM-DD" in branch names for parsing wrk naming conventions.
var datePattern = regexp.MustCompile(`-(\d{4}-\d{2}-\d{2})`)

// parseBranchNaming extracts branchBase, date, slug, and suffix from a wrk-style
// branch name like "master-2026-07-01-fix-login-1". Returns an error if the
// branch doesn't contain a recognizable date pattern.
func parseBranchNaming(branch string) (branchBase, date, slug string, suffix int, err error) {
	loc := datePattern.FindStringSubmatchIndex(branch)
	if loc == nil {
		return "", "", "", 0, fmt.Errorf("no date pattern in branch name %q", branch)
	}
	branchBase = branch[:loc[0]]
	date = branch[loc[2]:loc[3]]
	tail := branch[loc[1]:] // includes leading "-"
	if tail == "" {
		return branchBase, date, "", 0, nil
	}
	tail = tail[1:] // strip leading "-"
	parts := strings.Split(tail, "-")
	last := parts[len(parts)-1]

	if n, convErr := strconv.Atoi(last); convErr == nil && n >= 0 && n < 100 {
		if len(parts) > 1 {
			slug = strings.Join(parts[:len(parts)-1], "-")
			suffix = n
		} else {
			suffix = n
		}
	} else {
		slug = tail
	}
	return branchBase, date, slug, suffix, nil
}

// runSetTask renames a linked worktree via git worktree move to include a new
// task slug in the directory and branch names. Requires TTY confirmation (or
// WRK_SET_TASK_CONFIRM=1 env var) before executing the move.
func runSetTask(workDir string, taskDesc string, assumeYes, noCd, forceCd bool, execArgs []string) error {
	if strings.TrimSpace(taskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	newSlug := slugify(taskDesc)
	if newSlug == "" {
		return fmt.Errorf("wrk: task description %q produces an empty slug", taskDesc)
	}

	// Shell process cwd (inherited from interactive shell), not merely workDir.
	// Used after move to decide whether auto-cd is needed.
	shellCwd, _ := processCwd()

	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsLinked(cwd) {
		return fmt.Errorf("wrk: --set-task must be run from inside a linked worktree")
	}

	branch, err := worktree.ReadBranch(cwd)
	if err != nil {
		return fmt.Errorf("read branch: %w", err)
	}

	branchBase, date, _, _, err := parseBranchNaming(branch)
	if err != nil {
		return fmt.Errorf("wrk: cannot parse branch name %q — is this a wrk worktree? (%w)", branch, err)
	}

	mainRepo, err := worktree.ResolveMainRepo(cwd)
	if err != nil {
		return fmt.Errorf("resolve main repo: %w", err)
	}

	basename := filepath.Base(mainRepo)
	// Sanitize so legacy slash branches (feature/foo-DATE) migrate to feature-foo-…
	pathToken := sanitizeBranchToken(branchBase)

	// Fit slug to path/branch component budget (same rules as create).
	fittedSlug, fitErr := fitTaskSlugForNames(basename, pathToken, date, newSlug)
	if fitErr != nil {
		return fitErr
	}
	newSlug = fittedSlug

	// Compute new names. We don't know the old suffix from the dir name alone,
	// so we derive it from the current dir basename. Find the wrk-style naming
	// by looking for the date pattern in the dir basename.
	// Fixed/non-wrk directory names lack the date pattern and are rejected (P3).
	curBase := filepath.Base(cwd)
	curLoc := datePattern.FindStringSubmatchIndex(curBase)
	if curLoc == nil {
		return fmt.Errorf("wrk: cannot parse directory name %q — is this a wrk worktree?", curBase)
	}
	curDate := curBase[curLoc[2]:curLoc[3]]
	curTail := curBase[curLoc[1]:]
	curSuffix := 0
	if curTail != "" {
		curTail = curTail[1:] // strip leading "-"
		// Remove the old slug (if any) from the tail to extract the suffix.
		// After date: [-slug][-N]. The suffix is at the very end if numeric.
		parts := strings.Split(curTail, "-")
		last := parts[len(parts)-1]
		if n, convErr := strconv.Atoi(last); convErr == nil && n >= 0 && n < 100 {
			curSuffix = n
		}
	}

	if curDate != date {
		return fmt.Errorf("wrk: date mismatch between branch (%s) and directory (%s)", date, curDate)
	}

	parentDir := filepath.Dir(cwd)
	cleanCwdForName := filepath.Clean(cwd)

	// Preferred names use sanitized token; walk -N from curSuffix on path OR
	// branch collision (P3). Self (current path/branch) is not a collision.
	var newDirName, newBranch, newPath string
	found := false
	for suffix := curSuffix; suffix < 100; suffix++ {
		newDirName = fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
		if newSlug != "" {
			newDirName = fmt.Sprintf("%s-%s", newDirName, newSlug)
		}
		if suffix > 0 {
			newDirName = fmt.Sprintf("%s-%d", newDirName, suffix)
		}

		newBranch = pathToken + "-" + date
		if newSlug != "" {
			newBranch = newBranch + "-" + newSlug
		}
		if suffix > 0 {
			newBranch = fmt.Sprintf("%s-%d", newBranch, suffix)
		}

		newPath = filepath.Join(parentDir, newDirName)

		pathBlocked := false
		if filepath.Clean(newPath) != cleanCwdForName {
			if _, err := os.Stat(newPath); err == nil {
				pathBlocked = true
			}
		}
		branchBlocked := false
		if newBranch != branch && branchExists(mainRepo, newBranch) {
			branchBlocked = true
		}
		if pathBlocked || branchBlocked {
			continue
		}
		found = true
		break
	}
	if !found {
		return fmt.Errorf("wrk: could not find available path/branch name after 99 attempts")
	}

	// If nothing changed, just report.
	if newDirName == curBase && newBranch == branch {
		fmt.Println("task unchanged")
		return runExecInDir(cwd, execArgs)
	}

	pathChanges := filepath.Clean(newPath) != cleanCwdForName
	branchChanges := newBranch != branch

	// Before renaming: discover nested linked worktrees under cwd so we can
	// update their gitdir metadata after the move.
	type nestedWT struct {
		oldPath string
		relPath string // relative to cwd
	}
	var nested []nestedWT
	if pathChanges {
		repos, err := discoverStatusRepos(context.Background(), cwd)
		if err != nil {
			return fmt.Errorf("discover nested worktrees: %w", err)
		}
		for _, repo := range repos {
			if repo.RepoType != scan_repo.RepoTypeWorktree {
				continue
			}
			if !worktree.IsLinked(repo.Path) {
				continue
			}
			if filepath.Clean(repo.Path) == cleanCwdForName {
				continue
			}
			rel, err := filepath.Rel(cwd, repo.Path)
			if err != nil {
				continue
			}
			nested = append(nested, nestedWT{oldPath: repo.Path, relPath: rel})
		}
	}

	// Default auto-yes (assumeYes) skips rename prompt. --confirm clears assumeYes so we
	// prompt. WRK_SET_TASK_CONFIRM=1 is a test escape hatch that auto-confirms (no prompt).
	// TTY detection sticks to stdout (same fd used for the rename plan print).
	if !assumeYes && os.Getenv("WRK_SET_TASK_CONFIRM") != "1" {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("wrk: --set-task --confirm requires a terminal (tty)")
		}
		fmt.Printf("Rename worktree:\n  %s → %s\n  branch %s → %s\n", cwd, newPath, branch, newBranch)
		fmt.Print("Proceed? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "" && answer != "y" && answer != "Y" {
			return fmt.Errorf("wrk: --set-task aborted")
		}
	}

	if pathChanges {
		// Execute git worktree move
		cmd := gitCommand("-C", mainRepo, "worktree", "move", cwd, newPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git worktree move: %w\n%s", err, out)
		}
	}

	if branchChanges {
		// Rename the branch (also covers legacy slash → sanitized migration).
		branchCmd := gitCommand("-C", mainRepo, "branch", "-m", branch, newBranch)
		out, err := branchCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git branch rename: %w\n%s", err, out)
		}
	}

	// Update gitdir metadata for nested worktrees that moved with the parent.
	// Each nested worktree's .git file says "gitdir: <mainRepo>/.git/worktrees/<name>",
	// and <mainRepo>/.git/worktrees/<name>/gitdir contains the old absolute path
	// back to the worktree. We rewrite it to the new path.
	for _, nw := range nested {
		newWtPath := filepath.Join(newPath, nw.relPath)
		gitFile := filepath.Join(newWtPath, ".git")
		data, err := os.ReadFile(gitFile)
		if err != nil {
			continue
		}
		s := strings.TrimSpace(string(data))
		const gitdirPrefix = "gitdir: "
		if !strings.HasPrefix(s, gitdirPrefix) {
			continue
		}
		gitdirBase := strings.TrimSpace(s[len(gitdirPrefix):])
		gitdirFile := filepath.Join(gitdirBase, "gitdir")
		newGitdirContent := filepath.Join(newWtPath, ".git") + "\n"
		_ = os.WriteFile(gitdirFile, []byte(newGitdirContent), 0644)
	}

	if pathChanges {
		if err := rewriteConsumerReplacePaths(cwd, newPath); err != nil {
			return fmt.Errorf("rewrite go.mod replace paths: %w", err)
		}
	}

	fmt.Println(newPath)
	if err := runExecInDir(newPath, execArgs); err != nil {
		return err
	}
	// --force-cd bypasses cwd-missing gate; otherwise write follow-up cd only if
	// shell cwd is gone (user was inside the moved worktree). Surviving
	// sibling/main stays put without --force-cd.
	if forceCd {
		return forceLandInDir(newPath)
	}
	if pathChanges {
		return writeFollowupCDIfCwdMissing(noCd, shellCwd, newPath)
	}
	return nil
}
