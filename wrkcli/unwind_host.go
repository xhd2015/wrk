package wrkcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/commit_msg"
	"github.com/xhd2015/agent-pro/agent/git_runner"
	"github.com/xhd2015/wrk/wrkcli/unwind"
)

func mapUnwindError(err error) error {
	if err == nil {
		return nil
	}
	var ece unwind.ExitCodeError
	if errors.As(err, &ece) {
		return ExitCodeError{Code: ece.Code}
	}
	return err
}

func newUnwindHost() unwind.Host {
	return unwind.Host{
		GitRun:    gitRunDir,
		GitRunIO:  gitRunDirTo,
		GitOutput: gitOutputDir,
		ShortHEAD: shortHEAD,
		GoModTidy: goModTidy,
		PushMain: func(mainRepo string, dryRun, force bool, tags []string, io unwind.HostIO) error {
			return runPushMainWrite(mainRepo, dryRun, force, tags, true, io.Out())
		},
		Sync: func(workDir string, dryRun, color, noColor bool, io unwind.HostIO) error {
			_, err := runSyncWithColorTo(workDir, dryRun, color, noColor, io.Out(), io.Err())
			return err
		},
		GenCommit: func(workDir string, genArgs []string, dryRun, allowEmptySkip bool, io unwind.HostIO) error {
			_ = io
			return runGenCommitMsgStage(workDir, genArgs, dryRun, allowEmptySkip)
		},
		GenerateCommitMsg: generateCommitMsgCore,
		GitCommit:         gitCommitWithMessage,
		ReinstallLocal: func(mainPath string, color, noColor bool, io unwind.HostIO) (int, error) {
			st, err := runReinstallLocalExTo(mainPath, false, true, color, noColor, nil, io.Out(), io.Err())
			if err != nil {
				return 0, err
			}
			return st.Reinstalled, nil
		},
		MapMergeBackError: mapMergeBackSharedError,
		IsNoPushRemote:    isNoPushRemoteErr,
		IsNoStagedCommit:  isNoStagedCommitErr,
	}
}

func hostGenArgsFlagValue(genArgs []string, flag string) string {
	for i, a := range genArgs {
		name := a
		val := ""
		if j := strings.IndexByte(a, '='); j >= 0 {
			name = a[:j]
			val = a[j+1:]
		}
		if name != flag {
			continue
		}
		if val != "" {
			return val
		}
		if i+1 < len(genArgs) && !strings.HasPrefix(genArgs[i+1], "-") {
			return genArgs[i+1]
		}
		return ""
	}
	return ""
}

// generateCommitMsgCore calls agent-pro Generate only (no --add-all / --commit).
func generateCommitMsgCore(workDir string, genArgs []string, io unwind.HostIO) (string, error) {
	model := hostGenArgsFlagValue(genArgs, "--model")
	agentRunner := hostGenArgsFlagValue(genArgs, "--agent-runner")
	agentBinary := hostGenArgsFlagValue(genArgs, "--agent-runner-binary")
	if agentRunner == "" {
		agentRunner = "opencode"
	}
	logger := &genCommitWriterLogger{w: io.Err()}
	msg, err := commit_msg.Generate(workDir, commit_msg.GenerateOptions{
		Model:             model,
		AgentRunner:       agentRunner,
		AgentRunnerBinary: agentBinary,
		Logger:            logger,
	})
	if err != nil {
		return "", err
	}
	fmt.Fprintf(io.Err(), "--- Generated Commit Message ---\n")
	fmt.Fprintln(io.Out(), msg)
	return msg, nil
}

func gitCommitWithMessage(workDir, message string, noVerify bool, io unwind.HostIO) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("wrk: empty commit message")
	}
	output, err := git_runner.CommitWithRetry(workDir, message, 5, noVerify)
	if len(output) > 0 {
		fmt.Fprint(io.Err(), string(output))
	}
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out != "" {
			return fmt.Errorf("git commit failed: %w\n%s", err, out)
		}
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

type genCommitWriterLogger struct{ w io.Writer }

func (l *genCommitWriterLogger) Log(msg string) {
	w := l.w
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintln(w, msg)
}
func (l *genCommitWriterLogger) Error(msg string) {
	w := l.w
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintln(w, msg)
}
