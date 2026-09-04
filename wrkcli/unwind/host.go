package unwind

import (
	"fmt"
	"sync"
)

// Host is the CLI-owned ops bundle injected by wrkcli so this package does
// not import wrkcli (cycle: wrkcli → unwind).
type Host struct {
	GitRun            func(repoPath string, args ...string) error
	GitOutput         func(repoPath string, args ...string) (string, error)
	ShortHEAD         func(repo string) (string, error)
	GoModTidy         func(dir string) error
	PushMain          func(mainRepo string, dryRun, force bool, tags []string) error
	Sync              func(workDir string, dryRun, color, noColor bool) error
	GenCommit         func(workDir string, genArgs []string, dryRun, allowEmptySkip bool) error
	ReinstallLocal    func(mainPath string, color, noColor bool) (int, error)
	MapMergeBackError func(err error, op string) error
	IsNoPushRemote    func(error) bool
	IsNoStagedCommit  func(error) bool
}

var (
	hostMu     sync.RWMutex
	activeHost Host
)

// UseHost installs host for the duration of the returned restore func.
func UseHost(h Host) func() {
	return useHost(h)
}

func useHost(h Host) func() {
	hostMu.Lock()
	prev := activeHost
	activeHost = h
	hostMu.Unlock()
	return func() {
		hostMu.Lock()
		activeHost = prev
		hostMu.Unlock()
	}
}

func currentHost() Host {
	hostMu.RLock()
	defer hostMu.RUnlock()
	return activeHost
}

func hostErr(name string) error {
	return fmt.Errorf("unwind: %s host not configured", name)
}

func runPushMain(mainRepo string, dryRun, force bool, tags []string) error {
	h := currentHost()
	if h.PushMain == nil {
		return hostErr("PushMain")
	}
	return h.PushMain(mainRepo, dryRun, force, tags)
}

func runSyncWithColor(workDir string, dryRun, color, noColor bool) (struct{}, error) {
	h := currentHost()
	if h.Sync == nil {
		return struct{}{}, hostErr("Sync")
	}
	return struct{}{}, h.Sync(workDir, dryRun, color, noColor)
}

func runGenCommitMsgStage(workDir string, genArgs []string, dryRun, allowEmptySkip bool) error {
	h := currentHost()
	if h.GenCommit == nil {
		return hostErr("GenCommit")
	}
	return h.GenCommit(workDir, genArgs, dryRun, allowEmptySkip)
}

func mapMergeBackSharedError(err error, op string) error {
	h := currentHost()
	if h.MapMergeBackError == nil {
		return err
	}
	return h.MapMergeBackError(err, op)
}

func isNoPushRemoteErr(err error) bool {
	h := currentHost()
	if h.IsNoPushRemote == nil {
		return false
	}
	return h.IsNoPushRemote(err)
}

func isNoStagedCommitErr(err error) bool {
	h := currentHost()
	if h.IsNoStagedCommit == nil {
		return false
	}
	return h.IsNoStagedCommit(err)
}
