package wrkcli

import (
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

type invocationContext struct {
	origWd    string
	rawArgs   []string
	workDir   string
	mainRepo  string
	command   string
	eventArgs []string
	wrkHome   string
	skipEvent bool
	withGo    withgo.ResolveOptions
}

func newInvocationContext(origWd string, args []string) *invocationContext {
	return &invocationContext{
		origWd:  origWd,
		rawArgs: args,
		command: "create",
	}
}

func (ctx *invocationContext) finish(exitCode int) {
	if ctx.skipEvent || ctx.wrkHome == "" {
		return
	}
	workDir := ctx.workDir
	if workDir == "" {
		workDir = ctx.origWd
	}
	ev := storage.Event{
		TS:       time.Now().UTC().Format(time.RFC3339),
		Command:  ctx.command,
		WorkDir:  storage.NormalizePath(workDir),
		MainRepo: ctx.mainRepo,
		Args:     ctx.eventArgs,
		ExitCode: exitCode,
	}
	_ = storage.AppendEvent(ctx.wrkHome, ev)
}

func (ctx *invocationContext) autoRecord() error {
	main, err := storage.AutoRecord(ctx.wrkHome, ctx.workDir)
	if err != nil {
		return err
	}
	if main != "" {
		ctx.mainRepo = main
	}
	return nil
}

func resolveCommand(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, done, list, status, repos, mergeBack bool, bring bool, reinstallLocal, tagNext, propagateTags, syncFlag, pushFlag, prFlag, cd, mainFlag, unwind bool) string {
	switch {
	case setTaskFlagSet:
		return "set-task"
	case unwind:
		// Primary mode: compose with ship/land flags (done/tag-next/push) without
		// losing command identity.
		return "unwind"
	case projects:
		return "projects"
	case projectsDepGraph:
		return "projects-dep-graph"
	case addFlagSet:
		return "add"
	case removeFlagSet:
		return "rm"
	case whereFlagSet:
		return "where"
	case prFlag:
		// --pr primary; may compose with --status / --push / --gen-commit-msg
		// (event stays "pr"). When --pr --status, command is "pr" not "status".
		return "pr"
	case status:
		// status wins over main when both set (wrk --main --status)
		return "status"
	case done:
		// Prefer done / merge-back over reinstall-local / tag-next / sync so
		// composition keeps primary command identity (event "done", not tail).
		return "done"
	case mergeBack:
		return "merge-back"
	case reinstallLocal:
		// reinstall-local wins over main when both set (wrk --main --reinstall-local)
		return "reinstall-local"
	case cd:
		// cd wins over main when both set (wrk --main --cd)
		return "cd"
	case mainFlag:
		return "main"
	case repos:
		return "repos"
	case bring:
		return "bring"
	case list:
		return "list"
	case tagNext:
		return "tag-next"
	case propagateTags:
		return "propagate-tags"
	case syncFlag:
		return "sync"
	case pushFlag:
		// Bare --push primary (option R). Composition with --tag-next / --done
		// keeps the primary command name above.
		return "push"
	default:
		return "create"
	}
}

var flagValueArgs = map[string]struct{}{
	"--bring":    {},
	"-t":         {},
	"--task":     {},
	"--set-task": {},
	"--add":      {},
	"--rm":       {},
	"--port":     {},
	"--title":    {},
	"--comment":  {},
}

// varargsSlurpFlags take zero or more following non-flag tokens (less-flags Varargs).
var varargsSlurpFlags = map[string]struct{}{
	"--bring":           {},
	"--reinstall-local": {},
}

// slurpNonFlagTokens returns consecutive tokens from start that do not start
// with '-'. Used for Varargs-style --bring / --reinstall-local.
func slurpNonFlagTokens(args []string, start int) (values []string, next int) {
	next = start
	for next < len(args) && !strings.HasPrefix(args[next], "-") {
		values = append(values, args[next])
		next++
	}
	return values, next
}

func extractEventArgs(original, positionals []string) []string {
	pos := 0
	var out []string
	for i := 0; i < len(original); i++ {
		arg := original[i]
		if pos < len(positionals) && arg == positionals[pos] {
			pos++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			out = append(out, arg)
			if _, ok := varargsSlurpFlags[arg]; ok {
				vals, next := slurpNonFlagTokens(original, i+1)
				out = append(out, vals...)
				i = next - 1
				continue
			}
			if _, ok := flagValueArgs[arg]; ok {
				if i+1 < len(original) {
					out = append(out, original[i+1])
					i++
				}
			}
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}
