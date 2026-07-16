package wrkcli

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/agent/commit_msg"
)

// genCommitMsgDisallowedFlags are wrk mode / create-flow flags that cannot
// appear with --gen-commit-msg (library-owned flags like --dry-run are allowed).
var genCommitMsgDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--status", "--repos", "--projects",
	"--projects-dep-graph",
	"--scan-git-repos", "--no-cache",
	"--fetch", "--add", "--rm", "--where", "--cd", "--main",
	"--dep", "--bring", "--all-deps", "--reinstall-local", "--tag-next",
	"--propagate-tags", "--sync",
	"-t", "--task", "--set-task",
	"--exec",
	"--web", "--port", "--dev",
	"--version",
	"--set-config", "--bash-integration",
	"--push", "--json",
	"--confirm-from-stdin", "--no-in-module-replace",
	"--no-cd", "--force-cd",
	"--new-window", "--no-new-window",
	"--new-terminal", "--reuse-terminal", "--smart-terminal", "--no-new-terminal",
	"--open-in-agent", "--no-open-in-agent",
	"--no-config",
	"--color", "-v", "--verbose",
}

// runGenCommitMsg handles wrk --gen-commit-msg [...].
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
		forwarded = append(forwarded, arg)
	}

	ctx.skipEvent = true
	ctx.command = "gen-commit-msg"
	return commit_msg.RunGenCommitMsg(forwarded)
}
