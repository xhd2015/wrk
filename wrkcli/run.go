package wrkcli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
	"github.com/xhd2015/wrk/wrkcli/storage"
	lessflags "github.com/xhd2015/less-flags"
	"golang.org/x/term"
)

// Run executes wrk logic with args. The first positional argument,
// if present, is the source directory for all modes.
func Run(args []string) error {
	origWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	ctx := newInvocationContext(origWd, args)
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
	runErr = run(origWd, args, ctx)
	return runErr
}

func validateWhereFlagArg(args []string) error {
	for i, arg := range args {
		if arg != "--where" {
			continue
		}
		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return fmt.Errorf("wrk: --where requires a path argument")
		}
	}
	return nil
}

func run(origWd string, args []string, ctx *invocationContext) error {
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
	if hasArg(args, "--gen-commit-msg") {
		return runGenCommitMsg(args, ctx)
	}

	if err := validateWhereFlagArg(args); err != nil {
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
	var fetchFlag bool
	var verbose bool
	var addPath *string
	var removePath *string
	var confirmFromStdin bool
	var assumeYes bool
	var noInModuleReplace bool
	var depPath string
	var bringPath string
	var allDeps bool
	var reinstallLocal bool
	var tagNext bool
	var propagateTags bool
	var syncFlag bool
	var pushFlag bool
	var jsonFlag bool
	var dryRun bool
	var taskDesc *string
	var setTaskDesc *string
	var wherePath *string
	var noCd bool
	var forceCd bool
	var cd bool
	var mainFlag bool
	var execArgs []string
	// Create UX one-shot flags.
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
	// *string targets: nil = flag absent; non-nil empty = present but empty.
	// Cut("--exec") must be registered so tokens after --exec are never re-parsed as flags.
	remaining, err := lessflags.Bool("--done", &done).
		Bool("--merge-back", &mergeBack).
		Bool("-l,--list", &list).
		Bool("--status", &status).
		Bool("--repos", &repos).
		Bool("--projects", &projects).
		Bool("--projects-dep-graph", &projectsDepGraph).
		Bool("--fetch", &fetchFlag).
		Bool("-v,--verbose", &verbose).
		Bool("--color", &colorFlag).
		Bool("--web", &webFlag).
		Bool("--dev", &webDev).
		Int("--port", &portFlag).
		Bool("--scan-git-repos", &scanGitRepos).
		Bool("--no-cache", &noCache).
		String("--add", &addPath).
		String("--rm", &removePath).
		Bool("--confirm-from-stdin", &confirmFromStdin).
		Bool("-y,--yes", &assumeYes).
		Bool("--no-in-module-replace", &noInModuleReplace).
		Bool("--no-cd", &noCd).
		Bool("--force-cd", &forceCd).
		Bool("--cd", &cd).
		Bool("--main", &mainFlag).
		Bool("--all-deps", &allDeps).
		Bool("--reinstall-local", &reinstallLocal).
		Bool("--tag-next", &tagNext).
		Bool("--propagate-tags", &propagateTags).
		Bool("--sync", &syncFlag).
		Bool("--push", &pushFlag).
		Bool("--json", &jsonFlag).
		Bool("--dry-run", &dryRun).
		Bool("--new-window", &newWindow).
		Bool("--no-new-window", &noNewWindow).
		Bool("--new-terminal", &newTerminal).
		Bool("--reuse-terminal", &reuseTerminal).
		Bool("--smart-terminal", &smartTerminal).
		Bool("--no-new-terminal", &noNewTerminal).
		Bool("--open-in-agent", &openInAgent).
		Bool("--no-open-in-agent", &noOpenInAgent).
		Bool("--no-config", &noConfig).
		String("--dep", &depPath).
		String("--bring", &bringPath).
		String("-t,--task", &taskDesc).
		String("--set-task", &setTaskDesc).
		String("--where", &wherePath).
		Cut("--exec", &execArgs).
		Help("-h,--help", usage()).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			// Help text already printed by Parse; exit 0.
			ctx.skipEvent = true
			return nil
		}
		return err
	}

	// --force-cd and --no-cd are mutually exclusive (hard error before any work).
	if forceCd && noCd {
		return fmt.Errorf("wrk: --force-cd and --no-cd are mutually exclusive")
	}

	taskFlagSet := taskDesc != nil
	setTaskFlagSet := setTaskDesc != nil
	addFlagSet := addPath != nil
	removeFlagSet := removePath != nil
	whereFlagSet := wherePath != nil
	portFlagSet := portFlag != nil

	if webFlag {
		ctx.command = "web"
	} else if scanGitRepos {
		ctx.command = "scan-git-repos"
	} else {
		ctx.command = resolveCommand(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, done, list, status, repos, mergeBack, depPath, bringPath, allDeps, reinstallLocal, tagNext, propagateTags, syncFlag, cd, mainFlag)
	}
	ctx.eventArgs = extractEventArgs(args, remaining)

	setInvocationVerbose(verbose)
	worktree.GitVerboseLogger = logGitCommand
	defer func() {
		setInvocationVerbose(false)
		worktree.GitVerboseLogger = nil
	}()

	// --no-cache is only valid with --scan-git-repos.
	if noCache && !scanGitRepos {
		return fmt.Errorf("wrk: --no-cache is only valid with --scan-git-repos")
	}

	if fetchFlag && !projects && !status && !webFlag {
		return fmt.Errorf("wrk: --fetch is only valid with --projects or --status")
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
			addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag ||
			dryRun || pushFlag || jsonFlag || taskFlagSet || setTaskFlagSet || fetchFlag || noCd || forceCd ||
			cd || mainFlag || confirmFromStdin || noInModuleReplace || scanGitRepos ||
			newWindow || noNewWindow || newTerminal || reuseTerminal || smartTerminal ||
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

	// --scan-git-repos discovers main repos under roots and records them.
	if scanGitRepos {
		otherMode := done || mergeBack || list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" ||
			allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || taskFlagSet ||
			setTaskFlagSet || fetchFlag || noCd || forceCd || cd || mainFlag ||
			confirmFromStdin || noInModuleReplace || webFlag ||
			newWindow || noNewWindow || newTerminal || reuseTerminal || smartTerminal ||
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
		return runScanGitRepos(wrkHome, remaining, noCache)
	}

	// --cd requires exactly one path positional before defaulting workDir to cwd.
	if cd {
		if len(remaining) == 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --cd requires a path argument")
		}
		if len(remaining) > 1 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// --main takes no path positional when used alone. It may compose with --status
	// (status of main repo, no shell); positionals then follow --status rules.
	// It may also compose with --reinstall-local (and --dry-run as its modifier);
	// reinstall-local takes no path positionals.
	// Mutual exclusion with other modes is checked later; if another mode flag is
	// also set, prefer that error over unexpected arguments.
	if mainFlag {
		// reinstallLocal is a compose partner (not otherMode); dry-run only allowed with it.
		otherMode := done || list || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet ||
			whereFlagSet || depPath != "" || bringPath != "" || allDeps || tagNext || propagateTags || syncFlag || pushFlag || jsonFlag || mergeBack || taskFlagSet ||
			setTaskFlagSet || noCd || cd || spawnTarget != ""
		if !reinstallLocal {
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
		if !status && !reinstallLocal && len(remaining) > 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// --sync takes no positionals when used alone. It may compose with --done /
	// --merge-back (post-success pipeline); those modes may take a source dir.
	// With a primary, --sync may also compose with --tag-next / --push / --propagate-tags / --dry-run.
	// Prefer mode-clash errors over unexpected args when combined with other modes
	// (checked later alongside tag-next family).
	if syncFlag {
		// done and mergeBack are intentionally excluded so composition is allowed.
		otherMode := list || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" ||
			allDeps || reinstallLocal || jsonFlag || taskFlagSet || setTaskFlagSet ||
			cd || mainFlag || fetchFlag || spawnTarget != ""
		// tag-next / push / propagate-tags compose with --sync only when a primary is present.
		if (tagNext || pushFlag || propagateTags) && !done && !mergeBack {
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
	createMode := isCreateMode(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, repos, status, depPath, bringPath, allDeps, reinstallLocal, tagNext, propagateTags, syncFlag, list, done, mergeBack, cd, mainFlag)
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
	// Basename fallback: create/status/list/repos/--cd. --main uses cwd only.
	// One-arg create: if source resolve fails and the arg is task-like, offer
	// treat-as-task (promote creates from process cwd).
	var promotedTask string
	workDir, err := resolveSourceWorkDir(origWd, sourceDir, createMode || status || list || repos || cd, wrkHome, dirHint)
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
	if whereFlagSet && strings.TrimSpace(*wherePath) == "" {
		return fmt.Errorf("wrk: --where requires a path argument")
	}

	hasExec := len(execArgs) > 0
	if hasExec {
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
		if mergeBack {
			return fmt.Errorf("wrk: --exec is not valid with --merge-back")
		}
		if allDeps {
			return fmt.Errorf("wrk: --exec is not valid with --all-deps")
		}
		if reinstallLocal {
			return fmt.Errorf("wrk: --exec is not valid with --reinstall-local")
		}
		if tagNext {
			return fmt.Errorf("wrk: --exec is not valid with --tag-next")
		}
		if propagateTags {
			return fmt.Errorf("wrk: --exec is not valid with --propagate-tags")
		}
		// --exec is valid with --done --sync (runs after sync); invalid with bare --sync.
		if syncFlag && !done {
			return fmt.Errorf("wrk: --exec is not valid with --sync")
		}
		if mainFlag {
			return fmt.Errorf("wrk: --exec is not valid with --main")
		}
	}

	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && strings.TrimSpace(*setTaskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && (taskFlagSet || done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --set-task is mutually exclusive with other flags")
	}
	if setTaskFlagSet {
		return runSetTask(workDir, *setTaskDesc, assumeYes, noCd, forceCd, execArgs)
	}

	if taskFlagSet && strings.TrimSpace(*taskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --task is only valid with create mode.
	if taskFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: --task is mutually exclusive with --done, --merge-back, --list, --status, --repos, --projects, --add, --rm, --where, --dep and --all-deps")
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
	if repos && (done || list || status || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --repos is mutually exclusive with other modes")
	}
	if projects && (done || list || status || repos || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --projects is mutually exclusive with other modes")
	}
	if projectsDepGraph && (done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag || fetchFlag) {
		return fmt.Errorf("wrk: --projects-dep-graph is mutually exclusive with other modes")
	}
	if addFlagSet && (done || list || status || repos || projects || projectsDepGraph || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --add is mutually exclusive with other modes")
	}
	if removeFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --rm is mutually exclusive with other modes")
	}
	if whereFlagSet && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || cd || mainFlag) {
		return fmt.Errorf("wrk: --where is mutually exclusive with other modes")
	}
	if cd && (done || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || noCd || mainFlag) {
		return fmt.Errorf("wrk: --cd is mutually exclusive with other modes")
	}
	// --main composes with --status (and --fetch when status is set) and with
	// --reinstall-local (and --dry-run when reinstall is set); exclusive otherwise.
	if mainFlag {
		otherMode := done || list || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || tagNext || propagateTags || syncFlag || pushFlag || jsonFlag || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || noCd || cd || (!status && fetchFlag)
		if !reinstallLocal {
			otherMode = otherMode || dryRun
		}
		if otherMode {
			return fmt.Errorf("wrk: --main is mutually exclusive with other modes")
		}
	}
	if status && (done || list || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || dryRun || pushFlag || jsonFlag || spawnTarget != "" || cd) {
		return fmt.Errorf("wrk: --status is mutually exclusive with other modes")
	}
	if confirmFromStdin && !done && !mergeBack {
		return fmt.Errorf("wrk: --confirm-from-stdin is only valid with --done or --merge-back")
	}
	if noInModuleReplace && !done {
		return fmt.Errorf("wrk: --no-in-module-replace is only valid with --done")
	}
	if depPath != "" && bringPath != "" {
		return fmt.Errorf("wrk: --bring and --dep are mutually exclusive")
	}
	if depPath != "" && (done || list || mergeBack || tagNext || propagateTags || syncFlag || cd || mainFlag || reinstallLocal) {
		return fmt.Errorf("wrk: --dep is mutually exclusive with --done, --merge-back and --list")
	}
	if bringPath != "" && (done || list || mergeBack || tagNext || propagateTags || syncFlag || cd || mainFlag || reinstallLocal) {
		return fmt.Errorf("wrk: --bring is mutually exclusive with --done, --merge-back and --list")
	}
	if allDeps && (depPath != "" || bringPath != "" || done || list || mergeBack || tagNext || propagateTags || syncFlag || cd || mainFlag || reinstallLocal) {
		return fmt.Errorf("wrk: --all-deps is mutually exclusive with --dep, --bring, --done, --merge-back and --list")
	}
	// --reinstall-local is exclusive with other modes except --main (and dry-run modifier).
	if reinstallLocal {
		otherMode := done || list || mergeBack || status || repos || projects || projectsDepGraph ||
			addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || bringPath != "" ||
			allDeps || tagNext || propagateTags || syncFlag || cd || taskFlagSet || setTaskFlagSet ||
			spawnTarget != "" || pushFlag || jsonFlag || confirmFromStdin || noInModuleReplace ||
			fetchFlag || noCd || forceCd
		if otherMode {
			return fmt.Errorf("wrk: --reinstall-local is mutually exclusive with other modes")
		}
	}
	// --tag-next may compose with --done / --merge-back (and then with --sync / --push /
	// --propagate-tags / --dry-run), and with bare --propagate-tags (tag-then-propagate).
	// Without those it remains exclusive with other command modes.
	if tagNext {
		otherMode := depPath != "" || bringPath != "" || list || allDeps || reinstallLocal || cd || mainFlag ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != ""
		if !done && !mergeBack {
			// bare --tag-next: still exclusive with --sync (composition needs a primary)
			otherMode = otherMode || syncFlag
		}
		if otherMode {
			return fmt.Errorf("wrk: --tag-next is mutually exclusive with other modes")
		}
	}
	// --propagate-tags may compose with:
	//   - bare --tag-next (and then --push / --dry-run), or
	//   - primary --done / --merge-back (post-pipeline; ± sync / tag-next / push / dry-run).
	// Still exclusive with list/status/repos and other modes. Bare propagate+sync needs primary.
	// --json is rejected separately so the error names both flags.
	if propagateTags {
		otherMode := depPath != "" || bringPath != "" || list || allDeps || reinstallLocal || cd || mainFlag ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != ""
		if !done && !mergeBack {
			// bare --propagate-tags: exclusive with --sync (needs primary for that combo)
			otherMode = otherMode || syncFlag
			// --push alone with bare --propagate-tags is invalid; only with --tag-next compose.
			if pushFlag && !tagNext {
				otherMode = true
			}
		}
		if otherMode {
			return fmt.Errorf("wrk: --propagate-tags is mutually exclusive with other modes")
		}
	}
	// --sync may compose with --done / --merge-back (and then with --tag-next / --push /
	// --propagate-tags); still exclusive with other modes and with --json.
	if syncFlag {
		otherMode := depPath != "" || bringPath != "" || list || allDeps || reinstallLocal || cd || mainFlag ||
			projects || projectsDepGraph || repos || addFlagSet || removeFlagSet || whereFlagSet || status ||
			taskFlagSet || setTaskFlagSet || spawnTarget != "" || jsonFlag
		if !done && !mergeBack {
			otherMode = otherMode || tagNext || pushFlag || propagateTags
		}
		if otherMode {
			return fmt.Errorf("wrk: --sync is mutually exclusive with other modes")
		}
	}
	// --push is valid with --tag-next or a primary (--done / --merge-back).
	if pushFlag && !tagNext && !done && !mergeBack {
		return fmt.Errorf("wrk: --push is only valid with --tag-next")
	}
	// --json is only valid with bare --tag-next; never with --done / --merge-back /
	// --propagate-tags (compose or bare).
	if jsonFlag && done {
		return fmt.Errorf("wrk: --json is not valid with --done")
	}
	if jsonFlag && mergeBack {
		return fmt.Errorf("wrk: --json is not valid with --merge-back")
	}
	if jsonFlag && propagateTags {
		return fmt.Errorf("wrk: --json is not valid with --propagate-tags")
	}
	if jsonFlag && !tagNext {
		return fmt.Errorf("wrk: --json is only valid with --tag-next")
	}
	// --dry-run is valid with bare --sync / --all-deps / --tag-next / --propagate-tags /
	// --reinstall-local, with --done / --merge-back composition (full multi-stage plan is later phases),
	// and with --gen-commit-msg (handled early via runGenCommitMsg).
	if dryRun && !done && !mergeBack && !allDeps && !tagNext && !propagateTags && !syncFlag && !reinstallLocal {
		return fmt.Errorf("wrk: --dry-run is only valid with --done, --merge-back, --all-deps, --tag-next, --propagate-tags, --sync, --reinstall-local, or --gen-commit-msg")
	}

	// spawnTarget only applies to the create path. Reject for any other mode.
	if spawnTarget != "" && (depPath != "" || bringPath != "" || allDeps || reinstallLocal || tagNext || propagateTags || syncFlag || list || status || repos || projects || projectsDepGraph || addFlagSet || removeFlagSet || whereFlagSet || done || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	if whereFlagSet && len(remaining) > 0 {
		return fmt.Errorf("wrk: unexpected arguments")
	}

	if projects {
		colorEnabled := term.IsTerminal(int(os.Stdout.Fd())) || colorFlag
		return runProjects(wrkHome, colorEnabled, fetchFlag)
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
	if whereFlagSet {
		return runWhere(wrkHome, *wherePath)
	}
	if status {
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
	// --reinstall-local before bare --main so compose does not open a nested shell.
	if reinstallLocal {
		return runReinstallLocal(workDir, dryRun, mainFlag)
	}
	if mainFlag {
		return runMain(workDir)
	}
	if cd {
		return runCd(workDir, execArgs)
	}
	if repos {
		return runRepos(workDir)
	}
	if depPath != "" {
		return runDep(workDir, depPath, wrkHome, args, execArgs)
	}
	if bringPath != "" {
		return runBring(workDir, bringPath, wrkHome, args, execArgs)
	}
	if allDeps {
		return runAllDeps(workDir, dryRun)
	}
	if list {
		return runList(workDir)
	}
	// Prefer done / merge-back over bare tag-next / propagate / sync so composition
	// runs the primary path (post-pipeline: sync → tag-next → push → propagate-tags).
	if done {
		return runDone(workDir, wrkHome, confirmFromStdin, assumeYes, noInModuleReplace, noCd, forceCd, execArgs, syncFlag, tagNext, pushFlag, propagateTags, dryRun)
	}
	if mergeBack {
		return runMergeBack(workDir, wrkHome, confirmFromStdin, assumeYes, syncFlag, tagNext, pushFlag, propagateTags, dryRun)
	}
	// Bare compose: --tag-next --propagate-tags [--push] [--dry-run].
	// Fixed stage order tag-next → push? → propagate-tags (push inside tag-next).
	if tagNext && propagateTags {
		return runTagNextPropagateCompose(workDir, wrkHome, dryRun, pushFlag)
	}
	if tagNext {
		_, err := runTagNext(workDir, dryRun, pushFlag, jsonFlag)
		return err
	}
	if propagateTags {
		return runPropagateTags(workDir, wrkHome, dryRun)
	}
	if syncFlag {
		return runSync(workDir, dryRun)
	}
	task := ""
	if taskDesc != nil {
		task = *taskDesc
	}
	if promotedTask != "" {
		task = promotedTask
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

// usage returns the wrk help text printed by lessflags when -h/--help is given.
func usage() string {
	return `wrk — git worktree helper

Usage:
  wrk [dir] [target-dir] [flags]

Creates a git worktree from the current directory (or <dir>) and prints its
path. With <target-dir>, the worktree is spawned there instead of the default
location (~/.wrk/worktrees/).

Positional arguments:
  <dir>          optional source checkout to create the worktree from
                 (default: current directory)
  <target-dir>   optional spawn location for the worktree:
                   - missing, parent exists   -> spawn exactly at <target-dir>
                   - existing directory        -> spawn a default-named sub-dir
                   - missing parent            -> error

Flags:
  --done [--sync] [--tag-next] [--push] [--propagate-tags] [--dry-run] [--confirm-from-stdin]
                                  merge worktree branch back and remove it
                                  (optional post-success: --sync, --tag-next, --push, --propagate-tags from main)
  --merge-back [--sync] [--tag-next] [--push] [--propagate-tags] [--dry-run] [--confirm-from-stdin]
                                  merge worktree branch back WITHOUT removing it
                                  (optional post-success: --sync, --tag-next, --push, --propagate-tags from main)
  --done --no-in-module-replace   block --done on ANY local replace (strict)
  --list                          list worktrees (git worktree list)
  --status                        show status for git repos under this checkout
  --repos                         list git repos under this checkout
  --projects                      list recorded main repository paths
  --projects-dep-graph            module-level dep graph across registered projects
  --scan-git-repos [ROOT...]      discover main git repos under roots and record them
  --no-cache                      with --scan-git-repos: disable scan cache read/write
  --fetch                         with --projects or --status: fetch upstream before Remote: compare
  -v, --verbose                   log major git commands to stderr
  --add <dir>                     manually record a main repository path
  --rm <dir>                      remove a recorded main repository path
  --where <basename>              look up saved project path(s) by basename
  --cd <path|basename>            jump into directory (in-place follow-up or interactive shell)
  --main                          open nested shell at main repository root for this checkout
                                  (with --status: run status against the main repo instead;
                                   with --reinstall-local: reinstall from main repo modules)
  --dep <path>                    spawn a dependency worktree under ./external
  --bring <path>                  like --dep, but soft-skip go.mod replace when not a module dep
  --all-deps                      link every required dep from registered projects
  --reinstall-local [--dry-run]   reinstall local module binaries already in GOBIN/GOPATH/bin
                                  (with --main: scan main repository modules for this checkout)
  --tag-next [--dry-run] [--push] [--json]  plan/apply per-scope release tags
                                  (also: after successful --done / --merge-back; --json only bare)
                                  (also: with --propagate-tags: tag then bump consumers)
  --propagate-tags [--dry-run]    plan consumer go.mod bumps from source release tags
                                  (also: after --tag-next / --done / --merge-back;
                                  compose dry-run uses planned next tags when with --tag-next)
  --sync [--dry-run]              FF-only bi-directional sync main ↔ linked worktrees
                                  (also: after successful --done / --merge-back)
  --dry-run                       with --done/--merge-back/--all-deps/--tag-next/--propagate-tags/--sync/--reinstall-local/--gen-commit-msg: plan only
  --push                          with --tag-next: push each new tag to origin;
                                  with --done/--merge-back: push main branch (and tags when with --tag-next)
  --json                          with bare --tag-next only: machine-readable plan/result on stdout
                                  (not valid with --propagate-tags)
  --task <desc>                   append task slug to worktree/branch names
  --set-task <desc>               rename worktree/branch to match new task
  -y, --yes                       auto-confirm Y/n prompts (own worktree; cascade on TTY only)
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
  --gen-commit-msg [--dir DIR] [--model MODEL] [--agent-runner RUNNER]
                  [--agent-runner-binary PATH] [--commit] [--no-verify] [--dry-run]
                                  generate a commit message for staged changes (AI)
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

func runProjects(wrkHome string, colorEnabled bool, fetchEnabled bool) error {
	endPerf := beginProjectsPerfRun()
	defer endPerf()

	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return err
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

// runScanGitRepos discovers main git repositories under roots and records
// newly seen paths in projects.json with source "scan". Already-known paths
// are not re-printed. Empty CacheRoot uses the scan_repo library default.
func runScanGitRepos(wrkHome string, roots []string, noCache bool) error {
	if len(roots) == 0 {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return fmt.Errorf("wrk: --scan-git-repos requires at least one root (or ~/Projects)")
		}
		defaultRoot := filepath.Join(home, "Projects")
		if st, err := os.Stat(defaultRoot); err != nil || !st.IsDir() {
			return fmt.Errorf("wrk: --scan-git-repos requires at least one root (or ~/Projects)")
		}
		roots = []string{defaultRoot}
	}

	result, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:   roots,
		NoCache: noCache,
		// CacheRoot empty → product default when cache is enabled.
	})
	if err != nil {
		return err
	}
	for _, re := range result.RootErrors {
		fmt.Fprintf(os.Stderr, "warning: scan root %s: %s\n", re.Root, re.Error)
	}

	pf, err := storage.LoadProjects(wrkHome)
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(pf.Projects))
	for _, p := range pf.Projects {
		known[storage.NormalizePath(p.Path)] = true
	}

	var newly []string
	for _, repo := range result.Repos {
		if repo.RepoType != scan_repo.RepoTypeMain {
			continue
		}
		if repo.Error != "" {
			continue
		}
		path := storage.NormalizePath(repo.Path)
		if known[path] {
			continue
		}
		if err := storage.RecordProject(wrkHome, path, storage.SourceScan); err != nil {
			return err
		}
		known[path] = true
		newly = append(newly, path)
	}
	// Stable order for multi-root discoveries.
	sort.Strings(newly)
	for _, path := range newly {
		fmt.Println(path)
	}
	return nil
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

func runDone(workDir, wrkHome string, confirmFromStdin, assumeYes, noInModuleReplace, noCd, forceCd bool, execArgs []string, withSync, withTagNext, withPush, withPropagateTags, dryRun bool) error {
	// Shell process cwd (inherited from interactive shell), not merely workDir.
	// Used after remove to decide whether auto-cd is needed.
	shellCwd, _ := os.Getwd()

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

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	// Dry-run never confirms/applies cascade, so skip the non-interactive block.
	if !dryRun {
		if err := checkCascadeNonInteractive(consumerTop, checkoutRoot); err != nil {
			return err
		}
	}
	if err := cascadeLinkedWorktrees(consumerTop, checkoutRoot, confirmFromStdin, assumeYes, dryRun); err != nil {
		return err
	}

	// Guard: classify every local filesystem replace under the checkout (main
	// or sub-module). wrk --dep/--all-deps write replace => ./external/... and
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

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     true,
		DryRun:     dryRun,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		return err
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
	// Post-pipeline: sync → tag-next → push → propagate-tags → exec → land.
	// Dry-run: still print post stages in dry mode; skip exec/land.
	// Real success: apply post stages then exec/land.
	if dryRun {
		return runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, withPropagateTags, true)
	}
	if err := runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, withPropagateTags, false); err != nil {
		return err
	}
	if err := runExecInDir(result.TargetPath, execArgs); err != nil {
		return err
	}
	if forceCd {
		if err := forceLandInDir(result.TargetPath); err != nil {
			return err
		}
	} else if err := writeFollowupCDIfCwdMissing(noCd, shellCwd, result.TargetPath); err != nil {
		return err
	}
	return nil
}

func runMergeBack(workDir, wrkHome string, confirmFromStdin, assumeYes, withSync, withTagNext, withPush, withPropagateTags, dryRun bool) error {
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

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     false,
		DryRun:     dryRun,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		return err
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
	// Post-pipeline same order as runDone (no exec/land). Worktree kept.
	return runComposePostStages(result, checkoutRoot, wrkHome, withSync, withTagNext, withPush, withPropagateTags, dryRun)
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
func runComposePostStages(result *worktree.MergeBackResult, sourcePath, wrkHome string, withSync, withTagNext, withPush, withPropagateTags, dryRun bool) error {
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
		if err := runPushMain(mainPath, dryRun, tags); err != nil {
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

func checkCascadeNonInteractive(consumerTop, checkoutRoot string) error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	repos, err := discoverStatusRepos(context.Background(), consumerTop)
	if err != nil {
		return err
	}
	cleanCheckout := filepath.Clean(checkoutRoot)
	for _, repo := range repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			continue
		}
		if repo.RepoType != scan_repo.RepoTypeWorktree {
			continue
		}
		if !worktree.IsLinked(repo.Path) {
			continue
		}
		if filepath.Clean(repo.Path) == cleanCheckout {
			continue
		}
		mainRepo, err := worktree.ResolveMainRepo(repo.Path)
		if err != nil {
			return err
		}
		inclusion, err := worktree.HeadIncludedInMain(mainRepo, repo.Path)
		if err != nil {
			return err
		}
		if inclusion.Relation == "ahead" || inclusion.Relation == "diverged" {
			return fmt.Errorf("wrk --done: cannot cascade merge-back non-interactively: linked worktree %s is %s and needs confirmation", repo.Path, inclusion.Relation)
		}
	}
	return nil
}

func cascadeLinkedWorktrees(consumerTop, checkoutRoot string, confirmFromStdin, assumeYes, dryRun bool) error {
	repos, err := discoverStatusRepos(context.Background(), consumerTop)
	if err != nil {
		return err
	}

	cleanCheckout := filepath.Clean(checkoutRoot)
	for _, repo := range repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			if filepath.Clean(repo.Path) != filepath.Clean(consumerTop) {
				fmt.Fprintf(os.Stderr, "warning: skipping nested main repo %s\n", repo.Path)
			}
			continue
		}
		if repo.RepoType != scan_repo.RepoTypeWorktree {
			continue
		}
		if !worktree.IsLinked(repo.Path) {
			continue
		}
		if filepath.Clean(repo.Path) == cleanCheckout {
			continue
		}
		if err := mergeBackExternalWorktree(repo.Path, confirmFromStdin, assumeYes, dryRun); err != nil {
			return err
		}
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
// → remove only; ahead/diverged → prompt (via confirmFromStdin). A
// non-interactive ahead/diverged worktree errors (no force-removal fallback).
//
// When dryRun is true, prints a compact plan line and does not mutate.
func mergeBackExternalWorktree(externalPath string, confirmFromStdin, assumeYes, dryRun bool) error {
	if dryRun {
		fmt.Printf("would: cascade merge-back %s\n", externalPath)
		return nil
	}
	_, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: externalPath,
		TargetPath: "",
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	return err
}

func runDep(workDir string, depArg string, wrkHome string, rawArgs []string, execArgs []string) error {
	return runDepLike(workDir, depArg, wrkHome, rawArgs, execArgs, false)
}

// runBring is like runDep but always materializes the external worktree and
// best-effort applies go.mod replace (soft SKIP notices on stderr, exit 0).
func runBring(workDir string, bringArg string, wrkHome string, rawArgs []string, execArgs []string) error {
	return runDepLike(workDir, bringArg, wrkHome, rawArgs, execArgs, true)
}

// runDepLike implements --dep (strict) and --bring (best-effort).
// When bestEffort is false: hard-error if dep is not a go module or not a dependency,
// then create worktree + replace. When true: always create worktree + gitignore first,
// then soft-skip replace with stderr notices when analyse fails.
func runDepLike(workDir string, depArg string, wrkHome string, rawArgs []string, execArgs []string, bestEffort bool) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	depPath, err := resolveDirArg(depArg, true, wrkHome, &DirHintOptions{
		RawArgs: rawArgs,
		DepMode: true,
	})
	if err != nil {
		return err
	}
	if _, err := worktree.ResolveMainRepo(depPath); err != nil {
		return err
	}

	var matchingConsumerDirs []consumerMatch
	var depModDir string

	if !bestEffort {
		// Strict --dep: analyse first so hard errors leave no partial worktree.
		depModules, err := scan.Scan(depPath, scan.Options{})
		if err != nil {
			return fmt.Errorf("scan dep modules: %w", err)
		}
		if len(depModules) == 0 {
			return fmt.Errorf("not a go module: %s", depPath)
		}

		consumerModules, err := scan.Scan(consumerTop, scan.Options{})
		if err != nil {
			return fmt.Errorf("scan consumer modules: %w", err)
		}

		matchingConsumerDirs, depModDir = matchDepToConsumerModules(consumerTop, consumerModules, depModules)
		if len(matchingConsumerDirs) == 0 {
			return fmt.Errorf("%s is not a dependency of any consumer module", depPath)
		}
	}

	// Create external worktree + /external gitignore (always for both modes once
	// analyse passed for --dep; always for --bring after path/git resolve).
	externalPath, err := createExternalWorktreeForRepo(consumerTop, depPath)
	if err != nil {
		return err
	}

	if bestEffort {
		// Soft-skip replace with notices; worktree already exists.
		depModules, err := scan.Scan(depPath, scan.Options{})
		if err != nil {
			return fmt.Errorf("scan dep modules: %w", err)
		}
		if len(depModules) == 0 {
			fmt.Fprintf(os.Stderr, "SKIP local dep replacement: %s is not a go module\n", depPath)
			return finishDepLike(externalPath, execArgs)
		}

		consumerModules, err := scan.Scan(consumerTop, scan.Options{})
		if err != nil {
			return fmt.Errorf("scan consumer modules: %w", err)
		}
		if len(consumerModules) == 0 {
			fmt.Fprintf(os.Stderr, "SKIP local dep replacement: consumer has no Go modules\n")
			return finishDepLike(externalPath, execArgs)
		}

		matchingConsumerDirs, depModDir = matchDepToConsumerModules(consumerTop, consumerModules, depModules)
		if len(matchingConsumerDirs) == 0 {
			fmt.Fprintf(os.Stderr, "SKIP local dep replacement: %s is not a dependency of any consumer module\n", depPath)
			return finishDepLike(externalPath, execArgs)
		}
	}

	// The replace must target the directory holding the dep module's go.mod:
	// the repo root when depModDir is ".", or the sub-module subdir otherwise.
	replaceDir := externalPath
	if depModDir != "." {
		replaceDir = filepath.Join(externalPath, depModDir)
	}
	for _, m := range matchingConsumerDirs {
		if _, _, err := replace.ReplaceIn(m.dir, replaceDir); err != nil {
			return err
		}
		if err := commands.GoModTidy(&commands.GoModEditOptions{Dir: m.dir, Stderr: false, Stdout: false}); err != nil {
			return err
		}
	}

	return finishDepLike(externalPath, execArgs)
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

func finishDepLike(externalPath string, execArgs []string) error {
	absPath, err := filepath.Abs(externalPath)
	if err != nil {
		return fmt.Errorf("resolve external worktree path: %w", err)
	}
	fmt.Println(absPath)
	return runExecInDir(absPath, execArgs)
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
// warnings on stderr (shared by --bring/--dep/--all-deps and --dry-run).
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
// agree on naming), then creates the external dir, ensures .gitignore, and adds
// the worktree. It does NOT add a replace directive or run tidy. Used by
// runAllDeps (one worktree per repo, with per-module replaces added separately).
//
// Policy A: when planExternalWorktreePath reuses an existing external path, this
// still ensures /external gitignore but does not create a new worktree/branch.
func createExternalWorktreeForRepo(consumerTop, depPath string) (externalPath string, err error) {
	externalPath, err = planExternalWorktreePath(consumerTop, depPath)
	if err != nil {
		return "", err
	}

	externalDir := filepath.Join(consumerTop, "external")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return "", fmt.Errorf("create external dir: %w", err)
	}
	if err := ensureGitignoreExternal(consumerTop); err != nil {
		return "", err
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

func runAllDeps(workDir string, dryRun bool) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	// Scan the consumer tree for all Go modules (supports repos like dot-pkgs
	// whose module lives in a subdirectory with no root go.mod).
	consumerModules, err := scan.Scan(consumerTop, scan.Options{})
	if err != nil {
		return fmt.Errorf("scan consumer modules: %w", err)
	}

	type consumerModInfo struct {
		dir             string // abs path to the module's go.mod directory
		modulePath      string
		required        map[string]bool // dep paths this module requires
		alreadyReplaced map[string]bool // dep paths already replaced in this module
	}
	var consumerMods []consumerModInfo
	allRequired := make(map[string]bool)
	allAlreadyReplaced := make(map[string]bool)
	allConsumerModules := make(map[string]bool)

	for _, cm := range consumerModules {
		dir := filepath.Join(consumerTop, cm.Dir)
		info := consumerModInfo{
			dir:             dir,
			modulePath:      cm.Path,
			required:        make(map[string]bool),
			alreadyReplaced: make(map[string]bool),
		}
		if cm.Path != "" {
			allConsumerModules[cm.Path] = true
		}
		for _, req := range cm.Requires {
			info.required[req.Path] = true
			allRequired[req.Path] = true
		}
		for _, repl := range cm.Replaces {
			info.alreadyReplaced[repl.OldPath] = true
			allAlreadyReplaced[repl.OldPath] = true
		}
		consumerMods = append(consumerMods, info)
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}

	projectPaths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return err
	}

	type linkedDep struct {
		modulePath   string
		externalPath string // replace target (repo-root external path, or a sub-dir for nested sub-modules)
	}
	seen := make(map[string]bool)
	var linked []linkedDep
	tidied := make(map[string]bool)
	for _, projectPath := range projectPaths {
		if _, err := os.Stat(projectPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}
		if !worktree.IsMainRepo(projectPath) {
			continue
		}
		// mod/scan finds all modules in the repo (root + nested sub-modules) in
		// process, with vendor/testdata/gitignore skips. On error (e.g. unreadable
		// go.mod) skip the repo.
		modules, err := scan.Scan(projectPath, scan.Options{})
		if err != nil {
			continue
		}
		// Collect matched modules first so the repo's worktree is only created
		// when at least one module matches (and shared across all of them).
		var matched []scan.Module
		for _, m := range modules {
			if m.Path == "" || allConsumerModules[m.Path] {
				continue
			}
			if !allRequired[m.Path] || allAlreadyReplaced[m.Path] || seen[m.Path] {
				continue
			}
			matched = append(matched, m)
		}
		if len(matched) == 0 {
			continue
		}

		if dryRun {
			// Dry-run: compute the planned external path (read-only) and per-
			// module replace targets, but write nothing (no
			// createExternalWorktree, no GoModEditReplace, no tidy, no
			// gitignore).
			externalPath, err := planExternalWorktreePath(consumerTop, projectPath)
			if err != nil {
				return err
			}
			for _, m := range matched {
				target := externalPath
				if m.Dir != "." {
					target = filepath.Join(externalPath, filepath.FromSlash(m.Dir))
				}
				seen[m.Path] = true
				linked = append(linked, linkedDep{modulePath: m.Path, externalPath: target})
			}
			continue
		}
		// Real run: materialize the planned external worktree.
		externalPath, err := createExternalWorktreeForRepo(consumerTop, projectPath)
		if err != nil {
			return err
		}
		for _, m := range matched {
			// m.Dir is "." for the repo root module, or a slash-joined sub-dir
			// (e.g. "services/dep") for a nested sub-module. The replace target
			// is the sub-module's directory inside the external worktree.
			target := externalPath
			if m.Dir != "." {
				target = filepath.Join(externalPath, filepath.FromSlash(m.Dir))
			}
			// Replace in every consumer module that requires this dep.
			for _, cm := range consumerMods {
				if !cm.required[m.Path] || cm.alreadyReplaced[m.Path] {
					continue
				}
				opts := &commands.GoModEditOptions{Dir: cm.dir, Stderr: false, Stdout: false}
				if err := commands.GoModEditReplace(m.Path, target, opts); err != nil {
					return err
				}
				tidied[cm.dir] = true
			}
			seen[m.Path] = true
			linked = append(linked, linkedDep{modulePath: m.Path, externalPath: target})
		}
	}

	if !dryRun {
		for dir := range tidied {
			if err := commands.GoModTidy(&commands.GoModEditOptions{Dir: dir, Stderr: false, Stdout: false}); err != nil {
				return err
			}
		}
	}

	prefix := ""
	if dryRun {
		prefix = "would: "
	}
	for _, d := range linked {
		rel, err := filepath.Rel(consumerTop, d.externalPath)
		if err != nil {
			return fmt.Errorf("rel external path: %w", err)
		}
		fmt.Printf("%swrk %s at ./%s\n", prefix, d.modulePath, rel)
	}
	fmt.Printf("%swrked %d deps\n", prefix, len(linked))
	return nil
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
	cmd := gitCommand("-C", depMain, "worktree", "add", "-b", branch, externalPath, depBranch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}
	return nil
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
// Policy B (named bring): if source mainRepo already has any live linked
// worktree (anywhere, not only under target/external), prompt to skip (TTY,
// default Y) or hard-error (non-TTY). Skip prints the existing path on stdout
// and does not create. Answering n proceeds with create as today.
func runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date, slug string, noCd, forceCd bool, execArgs []string, taskDesc string, ux createUXPlan) error {
	// Resolve <target-dir> against the shell cwd (origWd), not the repo dir.
	absTarget := targetDir
	if !filepath.IsAbs(absTarget) {
		absTarget = filepath.Join(origWd, absTarget)
	}
	absTarget = filepath.Clean(absTarget)

	// Policy B: any live linked worktree of the source main repo.
	if existing, err := findLiveLinkedWorktrees(mainRepo); err != nil {
		return err
	} else if len(existing) > 0 {
		primary := existing[0]
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("wrk: %s already has a linked worktree at %s; refusing non-interactive create (default is skip; re-run in a TTY)", basename, primary)
		}
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
		if len(existing) > 1 {
			fmt.Fprintf(os.Stderr, "wrk: %s %s already has %d linked worktrees; reusing candidate %s\n", warnTok, basename, len(existing), pathDisp(primary))
			for _, p := range existing[1:] {
				fmt.Fprintf(os.Stderr, "wrk: %s also present: %s\n", warnTok, pathDisp(p))
			}
		}
		// Prompt on stderr; default is skip (Y/empty). No trailing newline before read.
		fmt.Fprintf(os.Stderr, "wrk: %s %s already has a linked worktree at %s, skip creating another? [Y/n] ", warnTok, basename, pathDisp(primary))
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
	if v := os.Getenv("WRK_HOME"); v != "" {
		return filepath.Abs(pathfmt.Expand(v))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".wrk"), nil
}

func resolveWrkDate() string {
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
	shellCwd, _ := os.Getwd()

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

	// TTY check (escape hatch for testing via WRK_SET_TASK_CONFIRM=1; -y bypasses)
	if !assumeYes && os.Getenv("WRK_SET_TASK_CONFIRM") != "1" {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("wrk: --set-task requires a terminal (tty)")
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
