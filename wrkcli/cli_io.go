package wrkcli

import (
	"io"
	"os"
)

// cliStdout / cliStderr are the process defaults used when a call site has no
// *invocationContext. Parallel-safe: always the process streams (no package
// capture, no mutex).
//
// In-process tests that need captured output must print via ctx.out() / ctx.errw()
// (set from RunOptions.Stdout/Stderr). Do not reintroduce process-global capture
// or os.Setenv/Chdir for leaf isolation — see DOCTEST_LINT.md §1.

func cliStdout() io.Writer { return os.Stdout }

func cliStderr() io.Writer { return os.Stderr }
