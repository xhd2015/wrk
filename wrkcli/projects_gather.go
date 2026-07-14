package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"github.com/xhd2015/gitops/git"
)

const (
	envProjectsWorktreeWorkers = "WRK_PROJECTS_WORKTREE_WORKERS"
	envProjectsProjectWorkers  = "WRK_PROJECTS_PROJECT_WORKERS"
)

type worktreeErrorDetail struct {
	path string
	msg  string
}

type projectStatusData struct {
	mainRepoPath    string
	branch          string
	short           string
	subject         string
	counts          status.WrkCounts
	remoteLine      string
	remoteRelation  git.BranchRelation
	dirtyWorktrees  int
	worktreeErrors  int
	pruneCount      int
	worktreeSummary string
	mainRepoError   string
	errorDetails    []worktreeErrorDetail
}

type fetchAsyncResult struct {
	upstream    string
	hasUpstream bool
	err         error
}

type remoteAsyncResult struct {
	remoteLine string
	relation   git.BranchRelation
	err        error
}

type wtCheckResult struct {
	path    string
	elapsed time.Duration
	isClean bool
	err     error
}

func projectsWorktreeWorkers() int {
	v := os.Getenv(envProjectsWorktreeWorkers)
	if v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return 1
		}
		return n
	}
	// Subprocess I/O bound; do not tie to GOMAXPROCS (tests may run with GOMAXPROCS=1).
	return 4
}

func projectsProjectWorkers() int {
	return envWorkerCount(envProjectsProjectWorkers, 4, 1)
}

func envWorkerCount(env string, defaultMax, minDefault int) int {
	v := os.Getenv(env)
	if v == "" {
		n := minInt(defaultMax, runtime.GOMAXPROCS(0))
		if n < minDefault {
			n = minInt(minDefault, defaultMax)
		}
		return n
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		n = minInt(defaultMax, runtime.GOMAXPROCS(0))
		if n < minDefault {
			n = minInt(minDefault, defaultMax)
		}
		return n
	}
	return n
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func startFetchAsync(mainRepoPath string, fetchEnabled bool) chan fetchAsyncResult {
	ch := make(chan fetchAsyncResult, 1)
	go func() {
		upstream, err := gitUpstreamRef(mainRepoPath)
		if err != nil {
			ch <- fetchAsyncResult{err: err}
			return
		}
		hasUpstream := upstream != ""
		if hasUpstream && fetchEnabled {
			if err := gitFetchUpstreamQuietNoOptionalLocks(mainRepoPath, upstream); err != nil {
				ch <- fetchAsyncResult{err: err}
				return
			}
		}
		ch <- fetchAsyncResult{hasUpstream: hasUpstream, upstream: upstream}
	}()
	return ch
}

func startRemoteCompareAsync(mainRepoPath string, fetchCh chan fetchAsyncResult, branchReady *sync.WaitGroup, mainStatusReady *sync.WaitGroup, branchFailed func() bool, branch func() string, counts func() status.WrkCounts, colorEnabled bool) chan remoteAsyncResult {
	ch := make(chan remoteAsyncResult, 1)
	go func() {
		fetch := <-fetchCh
		if fetch.err != nil {
			ch <- remoteAsyncResult{err: fetch.err}
			return
		}
		branchReady.Wait()
		mainStatusReady.Wait()
		if branchFailed() {
			ch <- remoteAsyncResult{}
			return
		}
		c := counts()
		isClean := c.Added == 0 && c.Changed == 0 && c.Renamed == 0 && c.Deleted == 0
		remoteColor := colorEnabled && isClean

		var (
			remoteLine string
			relation   git.BranchRelation
			err        error
		)
		if !fetch.hasUpstream {
			remoteLine = "Remote:       (no upstream)"
			relation = git.BranchRelationSame
		} else {
			var result *git.CompareBranchesResult
			result, err = projectsPerfTimedValue(mainRepoPath, "remote", func() (*git.CompareBranchesResult, error) {
				return git.CompareBranches(mainRepoPath, fetch.upstream, branch())
			})
			if err == nil {
				remoteLine = "Remote:       " + FormatRemoteBrief(result, remoteColor)
				relation = result.Relation
			}
		}
		ch <- remoteAsyncResult{remoteLine: remoteLine, relation: relation, err: err}
	}()
	return ch
}

func gatherProjectStatus(mainRepoPath string, colorEnabled bool, fetchEnabled bool) (projectStatusData, error) {
	mainRepoPath = storage.NormalizePath(mainRepoPath)

	type linkedResult struct {
		entries []worktree.Entry
		err     error
	}

	var (
		data            projectStatusData
		preludeErr      error
		mainStatusErr   string
		preludeMu       sync.Mutex
		preludeWG       sync.WaitGroup
		linkedRes       linkedResult
		linkedReady     sync.WaitGroup
		branchReady     sync.WaitGroup
		mainStatusReady sync.WaitGroup
		setPreludeErr   = func(err error) {
			preludeMu.Lock()
			defer preludeMu.Unlock()
			if preludeErr == nil && err != nil {
				preludeErr = err
			}
		}
		setMainStatusErr = func(msg string) {
			preludeMu.Lock()
			defer preludeMu.Unlock()
			if mainStatusErr == "" && msg != "" {
				mainStatusErr = msg
			}
		}
		branchFailed = func() bool {
			preludeMu.Lock()
			defer preludeMu.Unlock()
			return preludeErr != nil
		}
	)

	data.mainRepoPath = mainRepoPath
	linkedReady.Add(1)
	branchReady.Add(1)
	mainStatusReady.Add(1)

	fetchCh := startFetchAsync(mainRepoPath, fetchEnabled)
	remoteCh := startRemoteCompareAsync(mainRepoPath, fetchCh, &branchReady, &mainStatusReady, branchFailed, func() string { return data.branch }, func() status.WrkCounts { return data.counts }, colorEnabled)

	preludeWG.Add(4)
	go func() {
		defer preludeWG.Done()
		defer linkedReady.Done()
		entries, err := projectsPerfTimedValue(mainRepoPath, "list_linked", func() ([]worktree.Entry, error) {
			return worktree.ListLinked(mainRepoPath)
		})
		linkedRes = linkedResult{entries: entries, err: err}
	}()
	go func() {
		defer preludeWG.Done()
		defer branchReady.Done()
		branch, err := projectsPerfTimedValue(mainRepoPath, "main_branch", func() (string, error) {
			return gitOutputNoOptionalLocks(mainRepoPath, "rev-parse", "--abbrev-ref", "HEAD")
		})
		if err != nil {
			setPreludeErr(err)
			return
		}
		data.branch = branch
	}()
	go func() {
		defer preludeWG.Done()
		short, subject, err := projectsPerfTimedCommit(mainRepoPath)
		if err != nil {
			setPreludeErr(err)
			return
		}
		data.short = short
		data.subject = subject
	}()
	go func() {
		defer preludeWG.Done()
		defer mainStatusReady.Done()
		linkedReady.Wait()
		var skip map[string]struct{}
		if linkedRes.err == nil {
			skip = skipUntrackedRelPaths(mainRepoPath, linkedRes.entries)
		}
		counts, err := projectsPerfTimedValue(mainRepoPath, "main_status", func() (status.WrkCounts, error) {
			return gitProjectStatusCountsWithSkip(mainRepoPath, skip)
		})
		if err != nil {
			setMainStatusErr(gitCombinedOutputError(mainRepoPath, "status", "--porcelain"))
			setPreludeErr(err)
			return
		}
		data.counts = counts
	}()

	linkedReady.Wait()

	var linkedEntries []worktree.Entry
	if linkedRes.err == nil {
		linkedEntries = linkedRes.entries
	}

	alive := aliveLinkedEntries(linkedEntries)
	for _, entry := range linkedEntries {
		if worktree.IsDead(entry.Path) {
			data.pruneCount++
		}
	}

	var (
		clean     int
		dirty     int
		wtChecks  int
		wtElapsed time.Duration
		wtResults []wtCheckResult
		wtWG      sync.WaitGroup
	)

	if len(alive) > 0 {
		wtWG.Add(1)
		go func() {
			defer wtWG.Done()
			poolStart := time.Now()
			workers := minInt(projectsWorktreeWorkers(), len(alive))
			jobs := make(chan int, len(alive))
			for i := range alive {
				jobs <- i
			}
			close(jobs)

			results := make([]wtCheckResult, len(alive))
			var poolWG sync.WaitGroup
			for w := 0; w < workers; w++ {
				poolWG.Add(1)
				go func() {
					defer poolWG.Done()
					for i := range jobs {
						entry := alive[i]
						start := time.Now()
						isClean, err := gitWorktreeIsClean(entry.Path)
						results[i] = wtCheckResult{
							path:    entry.Path,
							elapsed: time.Since(start),
							isClean: isClean,
							err:     err,
						}
					}
				}()
			}
			poolWG.Wait()
			wtElapsed = time.Since(poolStart)
			wtResults = results
		}()
	}

	mainStatusReady.Wait()

	var remote remoteAsyncResult
	var endWG sync.WaitGroup
	endWG.Add(1)
	go func() {
		defer endWG.Done()
		remote = <-remoteCh
	}()
	if len(alive) > 0 {
		endWG.Add(1)
		go func() {
			defer endWG.Done()
			wtWG.Wait()
		}()
	}
	preludeWG.Wait()
	endWG.Wait()

	if preludeErr != nil {
		data.mainRepoError = mainStatusErr
		if data.mainRepoError == "" {
			data.mainRepoError = preludeErr.Error()
		}
		return data, nil
	}

	if remote.err != nil {
		data.remoteLine = "Remote:       error: " + remote.err.Error()
		data.remoteRelation = git.BranchRelationSame
	} else {
		data.remoteLine = remote.remoteLine
		data.remoteRelation = remote.relation
	}

	for _, result := range wtResults {
		wtChecks++
		if result.err != nil {
			data.worktreeErrors++
			data.errorDetails = append(data.errorDetails, worktreeErrorDetail{
				path: storage.NormalizePath(result.path),
				msg:  result.err.Error(),
			})
		} else if result.isClean {
			clean++
		} else {
			dirty++
		}
		recordProjectsPerfWorktree(mainRepoPath, result.path, result.elapsed)
	}
	recordProjectsPerfAggregate(mainRepoPath, "worktree_status_all", wtChecks, wtElapsed)

	summary, err := projectsPerfTimedValue(mainRepoPath, "worktree_summary", func() (string, error) {
		return formatLinkedWorktreeSummary(clean, dirty, data.worktreeErrors, data.pruneCount, colorEnabled), nil
	})
	if err != nil {
		return data, nil
	}
	data.worktreeSummary = summary
	data.dirtyWorktrees = dirty
	return data, nil
}

func projectsPerfTimedCommit(mainRepoPath string) (short, subject string, err error) {
	p := currentProjectsPerf
	start := time.Now()
	short, subject, err = gitCommitShortSubject(mainRepoPath)
	if p != nil {
		elapsed := time.Since(start)
		half := elapsed / 2
		recordProjectsPerfPhase(mainRepoPath, "main_commit_short", half)
		recordProjectsPerfPhase(mainRepoPath, "main_commit_subject", elapsed-half)
	}
	return short, subject, err
}

func gitCommitShortSubject(repoPath string) (short, subject string, err error) {
	out, err := gitOutputNoOptionalLocks(repoPath, "log", "-1", "--pretty=format:%h %s")
	if err != nil {
		return "", "", err
	}
	if i := strings.IndexByte(out, ' '); i >= 0 {
		return out[:i], out[i+1:], nil
	}
	return out, "", nil
}

func printProjectStatusFromData(data projectStatusData, colorEnabled bool) {
	if data.mainRepoError != "" {
		fmt.Printf("Dir:          %s\n", data.mainRepoPath)
		statusVal := "error: " + data.mainRepoError
		if colorEnabled {
			statusVal = colorize(statusVal, ansiRed)
		}
		fmt.Printf("Status:       %s\n", statusVal)
		return
	}

	fmt.Printf("Dir:          %s\n", data.mainRepoPath)
	fmt.Printf("Branch:       %s\n", data.branch)
	fmt.Printf("Commit:       %s  %s\n", data.short, data.subject)
	statusLine := formatStatusCounts(data.counts, colorEnabled, false)
	fmt.Printf("Status:       %s\n", statusLine)
	fmt.Println(data.remoteLine)
	if len(data.errorDetails) > 0 {
		fmt.Printf("Worktrees:    %s\n", data.worktreeSummary)
		for _, detail := range data.errorDetails {
			line := formatWorktreeErrorDetailLine(detail.path, detail.msg, colorEnabled)
			fmt.Printf("%s\n", line)
		}
		return
	}
	fmt.Printf("Worktrees:    %s\n", data.worktreeSummary)
}

func formatWorktreeErrorDetailLine(path, msg string, colorEnabled bool) string {
	errVal := "error: " + msg
	if colorEnabled {
		errVal = colorize(errVal, ansiRed)
	}
	return fmt.Sprintf("  %s  %s", path, errVal)
}

func printProjectStatusBlock(mainRepoPath string, colorEnabled bool, fetchEnabled bool) error {
	data, err := gatherProjectStatus(mainRepoPath, colorEnabled, fetchEnabled)
	if err != nil {
		return err
	}
	printProjectStatusFromData(data, colorEnabled)
	return nil
}

func aliveLinkedEntries(linked []worktree.Entry) []worktree.Entry {
	var alive []worktree.Entry
	for _, entry := range linked {
		if worktree.IsDead(entry.Path) {
			continue
		}
		alive = append(alive, entry)
	}
	return alive
}

func skipUntrackedRelPaths(mainRepo string, linked []worktree.Entry) map[string]struct{} {
	skip := make(map[string]struct{})
	for _, entry := range linked {
		if worktree.IsDead(entry.Path) {
			continue
		}
		rel, err := filepath.Rel(mainRepo, entry.Path)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		skip[rel] = struct{}{}
	}
	return skip
}

func gitProjectStatusCountsWithSkip(repoPath string, skipUntracked map[string]struct{}) (status.WrkCounts, error) {
	out, err := gitOutputNoOptionalLocks(repoPath, "status", "--porcelain")
	if err != nil {
		return status.WrkCounts{}, err
	}
	return parseProjectStatusCounts(out, skipUntracked), nil
}

func formatLinkedWorktreeSummary(clean, dirty, errors, prunes int, colorEnabled bool) string {
	total := clean + dirty + errors
	parts := []string{fmt.Sprintf("%d total", total)}
	if colorEnabled && dirty > 0 {
		parts = append(parts, colorize(fmt.Sprintf("%d dirty", dirty), ansiRed))
	} else {
		parts = append(parts, fmt.Sprintf("%d dirty", dirty))
	}
	if errors > 0 {
		errPart := fmt.Sprintf("%d error", errors)
		if colorEnabled {
			errPart = colorize(errPart, ansiRed)
		}
		parts = append(parts, errPart)
	}
	if prunes > 0 {
		parts = append(parts, fmt.Sprintf("%d prune", prunes))
	}
	return strings.Join(parts, ", ")
}