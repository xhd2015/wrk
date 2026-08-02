package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
	"github.com/xhd2015/wrk/workops"
)

// resolveMainRepoForWorkDir returns the main repository root for workDir.
// Errors match runMain messaging when workDir is not a git checkout.
// Delegates to workops.WhereMain.
func resolveMainRepoForWorkDir(workDir string) (string, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	mainRepo, err := workops.WhereMain(cwd)
	if err != nil {
		// Preserve historical messaging for non-git / not-a-repo paths.
		if strings.Contains(err.Error(), "is not a git repository") {
			return "", fmt.Errorf("%s is not a git repository", cwd)
		}
		return "", err
	}
	return mainRepo, nil
}

// runMain opens a nested interactive shell at the main repository root for the
// current checkout. Always nested (ignores WRK_FOLLOWUP_FILE); minimal UX.
// workDir is the process cwd (no path positional).
func runMain(workDir string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	mainRepo, err := resolveMainRepoForWorkDir(cwd)
	if err != nil {
		return err
	}

	// Already at main repo root: notice on stderr, no shell, exit 0.
	if sameDirPath(cwd, mainRepo) {
		fmt.Fprintf(os.Stderr, "wrk: already at main repository root: %s\n", mainRepo)
		return nil
	}

	// Always nested shell — never write follow-up cd; no install hint / path print.
	err = interactive.LoginInteractive(mainRepo, filepath.Base(mainRepo), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}
