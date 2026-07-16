package wrkcli

import (
	"strings"
	"time"

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

func resolveCommand(projects, projectsDepGraph, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, done, list, status, repos, mergeBack bool, depPath, bringPath string, allDeps, reinstallLocal, tagNext, propagateTags, syncFlag, cd, mainFlag bool) string {
	switch {
	case setTaskFlagSet:
		return "set-task"
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
	case status:
		// status wins over main when both set (wrk --main --status)
		return "status"
	case reinstallLocal:
		// reinstall-local wins over main when both set (wrk --main --reinstall-local)
		return "reinstall-local"
	case mainFlag:
		return "main"
	case cd:
		return "cd"
	case repos:
		return "repos"
	case depPath != "":
		return "dep"
	case bringPath != "":
		return "bring"
	case allDeps:
		return "all-deps"
	case list:
		return "list"
	case done:
		// Prefer done / merge-back over tag-next / sync so composition keeps
		// primary command identity (event command "done", not "tag-next").
		return "done"
	case mergeBack:
		return "merge-back"
	case tagNext:
		return "tag-next"
	case propagateTags:
		return "propagate-tags"
	case syncFlag:
		return "sync"
	default:
		return "create"
	}
}

var flagValueArgs = map[string]struct{}{
	"--dep":      {},
	"--bring":    {},
	"-t":         {},
	"--task":     {},
	"--set-task": {},
	"--add":      {},
	"--rm":       {},
	"--where":    {},
	"--port":     {},
}

func extractEventArgs(original, positionals []string) []string {
	pos := 0
	skipValue := false
	var out []string
	for _, arg := range original {
		if skipValue {
			skipValue = false
			out = append(out, arg)
			continue
		}
		if pos < len(positionals) && arg == positionals[pos] {
			pos++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			out = append(out, arg)
			if _, ok := flagValueArgs[arg]; ok {
				skipValue = true
			}
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}