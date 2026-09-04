package wrkcli

import (
	"errors"

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
		GitOutput: gitOutputDir,
		ShortHEAD: shortHEAD,
		GoModTidy: goModTidy,
		PushMain:  runPushMain,
		Sync: func(workDir string, dryRun, color, noColor bool) error {
			_, err := runSyncWithColor(workDir, dryRun, color, noColor)
			return err
		},
		GenCommit: runGenCommitMsgStage,
		ReinstallLocal: func(mainPath string, color, noColor bool) (int, error) {
			st, err := runReinstallLocalEx(mainPath, false, true, color, noColor, nil)
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
