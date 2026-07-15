package wrkcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gitcmd "github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/checkout"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"github.com/xhd2015/gitops/git"
)

var wrkCheckoutOpts = checkout.Options{
	StatusStyle:        status.StyleWrk,
	PorcelainUntracked: true,
}

type statusBlockPrintOpts struct {
	showMaster *bool
}

// statusDirLine formats a Dir: value relative to invocation cwd.
// Rel fail or more than two leading ".." segments → absolute NormalizePath.
func statusDirLine(displayCwd, repoPath string) string {
	base := storage.NormalizePath(displayCwd)
	target := storage.NormalizePath(repoPath)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	rel = filepath.Clean(rel)
	slash := filepath.ToSlash(rel)
	leading := 0
	for _, p := range strings.Split(slash, "/") {
		if p == ".." {
			leading++
			continue
		}
		break
	}
	if leading > 2 {
		return target
	}
	return slash
}

func sameNormalizedPath(a, b string) bool {
	return storage.NormalizePath(a) == storage.NormalizePath(b)
}

// runStatus prints status for statusRoot. displayCwd is the invocation work
// directory used only for Dir: labels (kept when --main rewrites status root).
func runStatus(statusRoot, displayCwd string, colorEnabled bool, fetchEnabled bool) error {
	cwd, err := filepath.Abs(statusRoot)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}
	displayBase, err := filepath.Abs(displayCwd)
	if err != nil {
		return fmt.Errorf("resolve display cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	if mainRepo, ok := linkedInTreeMainRepo(cwd); ok {
		return runStatusLinkedInTreeCwd(displayBase, cwd, mainRepo, colorEnabled)
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return err
	}

	scanPathList := make([]string, 0, len(repos))
	scanPathSet := make(map[string]struct{}, len(repos))
	for _, repo := range repos {
		scanPathList = append(scanPathList, repo.Path)
		scanPathSet[storage.NormalizePath(repo.Path)] = struct{}{}
	}

	showRemote := worktree.IsMainRepo(checkoutRoot)
	effectiveFetch := fetchEnabled && showRemote

	// Non-main checkouts: scan-order only (no ListLinked partition / external header).
	if !showRemote {
		scanColorEnabled := colorEnabled
		blocksPrinted := 0
		for _, repo := range repos {
			if blocksPrinted > 0 {
				fmt.Println()
			}
			if err := printStatusBlock(displayBase, checkoutRoot, repo.Path, colorEnabled, scanColorEnabled, false, false, statusBlockPrintOpts{}); err != nil {
				return err
			}
			blocksPrinted++
		}
		return nil
	}

	// Main-repo: primary = main + ListLinked (porcelain); external = other scan paths.
	linked, err := worktree.ListLinked(checkoutRoot)
	if err != nil {
		return err
	}
	linkedOrdered := make([]string, 0, len(linked))
	for _, entry := range linked {
		linkedOrdered = append(linkedOrdered, entry.Path)
	}
	lists := PartitionStatusPaths(checkoutRoot, scanPathList, linkedOrdered)

	// Preserve prior scan-color quirk: disable Status/Master/Remote coloring on
	// printStatusBlock when any primary path uses appended-style presentation
	// (out-of-tree or dead linked). Broken appended blocks still color via colorEnabled.
	hasAppendedStyle := false
	for _, p := range lists.Primary {
		if statusPathNeedsAppendedPresentation(p, scanPathSet) {
			hasAppendedStyle = true
			break
		}
	}
	scanColorEnabled := colorEnabled && !hasAppendedStyle

	blocksPrinted := 0
	for _, path := range lists.Primary {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		if statusPathNeedsAppendedPresentation(path, scanPathSet) {
			printAppendedLinkedBlock(displayBase, path, colorEnabled)
		} else {
			if err := printStatusBlock(displayBase, checkoutRoot, path, colorEnabled, scanColorEnabled, showRemote, effectiveFetch, statusBlockPrintOpts{}); err != nil {
				return err
			}
		}
		blocksPrinted++
	}

	if len(lists.External) > 0 {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		// Gray ANSI when colorEnabled (P3); plain ASCII otherwise.
		if colorEnabled {
			fmt.Println(colorize("---- external ----", ansiGrey))
		} else {
			fmt.Println("---- external ----")
		}
		for _, path := range lists.External {
			fmt.Println()
			// External nested: printStatusBlock; Remote only for main identity (none here).
			if err := printStatusBlock(displayBase, checkoutRoot, path, colorEnabled, scanColorEnabled, showRemote, effectiveFetch, statusBlockPrintOpts{}); err != nil {
				return err
			}
			blocksPrinted++
		}
	}
	return nil
}

// statusPathNeedsAppendedPresentation is true for dead/prunable or out-of-scan
// linked paths that historically used printAppendedLinkedBlock (out-of-tree WRK
// worktrees). In-scan healthy paths (main + in-tree linked) use printStatusBlock.
func statusPathNeedsAppendedPresentation(path string, scanPathSet map[string]struct{}) bool {
	if worktree.IsDead(path) {
		return true
	}
	_, inScan := scanPathSet[storage.NormalizePath(path)]
	return !inScan
}

func linkedInTreeMainRepo(cwd string) (string, bool) {
	if !worktree.IsLinked(cwd) {
		return "", false
	}
	mainRepo, err := worktree.ReadMainRepo(cwd)
	if err != nil {
		return "", false
	}
	cleanMain := filepath.Clean(mainRepo)
	cleanCwd := filepath.Clean(cwd)
	if cleanCwd == cleanMain {
		return "", false
	}
	rel, err := filepath.Rel(cleanMain, cleanCwd)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return mainRepo, true
}

func runStatusLinkedInTreeCwd(displayCwd, cwd, mainRepo string, colorEnabled bool) error {
	repos, err := discoverStatusRepos(context.Background(), mainRepo)
	if err != nil {
		return err
	}

	blocksPrinted := 0
	printBlock := func(repoPath string, opts statusBlockPrintOpts) error {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		blocksPrinted++
		// showRemote false: linked-in-tree cwd path never prints Remote.
		return printStatusBlock(displayCwd, mainRepo, repoPath, colorEnabled, colorEnabled, false, false, opts)
	}

	showMasterFalse := false
	if err := printBlock(cwd, statusBlockPrintOpts{showMaster: &showMasterFalse}); err != nil {
		return err
	}

	showMasterTrue := true
	for _, repo := range repos {
		if worktree.IsMainRepo(repo.Path) || !worktree.IsLinked(repo.Path) {
			continue
		}
		if err := printBlock(repo.Path, statusBlockPrintOpts{showMaster: &showMasterTrue}); err != nil {
			return err
		}
	}
	return nil
}

func runRepos(workDir string) error {
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

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		rel, err := filepath.Rel(checkoutRoot, repo.Path)
		if err != nil {
			return fmt.Errorf("resolve relative repo path: %w", err)
		}
		fmt.Println(filepath.ToSlash(rel))
	}
	return nil
}

func discoverStatusRepos(ctx context.Context, root string) ([]scan_repo.Repo, error) {
	result, err := scan_repo.Scan(ctx, scan_repo.Options{Roots: []string{root}})
	return result.Repos, err
}

func printAppendedLinkedBlock(displayCwd, repoPath string, colorEnabled bool) {
	dirLine := statusDirLine(displayCwd, repoPath)

	if worktree.IsDead(repoPath) {
		fmt.Printf("Dir:          %s\n", dirLine)
		fmt.Printf("Status:       prunable\n")
		return
	}

	meta := checkout.Enrich(context.Background(), repoPath, wrkCheckoutOpts)
	if meta.Error != "" {
		printBrokenStatusBlock(dirLine, brokenStatusMessage(meta, repoPath), colorEnabled)
		return
	}
	applyWrkStatusWithLinkedSkip(repoPath, &meta)
	masterBrief, _, err := masterBriefForRepo(repoPath, meta.Branch, colorEnabled)
	if err != nil {
		printBrokenStatusBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}

	fmt.Printf("Dir:          %s\n", dirLine)
	fmt.Printf("Branch:       %s\n", meta.Branch)
	fmt.Printf("Commit:       %s  %s\n", meta.CommitSHA, meta.CommitMsg)
	fmt.Printf("Status:       %s\n", formatStatusText(meta.Status, colorEnabled, true))
	fmt.Printf("Master:       %s\n", masterBrief)
}

func printBrokenStatusBlock(dirLine, msg string, colorEnabled bool) {
	statusVal := "error: " + msg
	if colorEnabled {
		statusVal = colorize(statusVal, ansiRed)
	}
	fmt.Printf("Dir:          %s\n", dirLine)
	fmt.Printf("Status:       %s\n", statusVal)
}

func brokenStatusMessage(meta checkout.Meta, repoPath string) string {
	if meta.Error != "" {
		return meta.Error
	}
	return gitCombinedOutputError(repoPath, "status", "--porcelain")
}

// printStatusBlock prints one status block. statusRoot is the status checkout root
// (ShowToplevel of status cwd); Remote is printed only for that block when showRemote
// is set — not for nested main-repo blocks under a multi-repo scan, and not via Dir==".".
func printStatusBlock(displayCwd, statusRoot, repoPath string, colorEnabled, scanColorEnabled bool, showRemote bool, fetchEnabled bool, opts statusBlockPrintOpts) error {
	dirLine := statusDirLine(displayCwd, repoPath)

	meta := checkout.Enrich(context.Background(), repoPath, wrkCheckoutOpts)
	if meta.Error != "" {
		printBrokenStatusBlock(dirLine, brokenStatusMessage(meta, repoPath), colorEnabled)
		return nil
	}
	applyWrkStatusWithLinkedSkip(repoPath, &meta)

	hasMaster := worktree.IsLinked(repoPath)
	if opts.showMaster != nil {
		hasMaster = *opts.showMaster
	}
	var masterBrief string
	if hasMaster {
		var err error
		masterBrief, _, err = masterBriefForRepo(repoPath, meta.Branch, scanColorEnabled)
		if err != nil {
			printBrokenStatusBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
			return nil
		}
	}

	fmt.Printf("Dir:          %s\n", dirLine)
	fmt.Printf("Branch:       %s\n", meta.Branch)
	fmt.Printf("Commit:       %s  %s\n", meta.CommitSHA, meta.CommitMsg)

	statusLine := formatStatusText(meta.Status, scanColorEnabled, true)
	fmt.Printf("Status:       %s\n", statusLine)
	if hasMaster {
		fmt.Printf("Master:       %s\n", masterBrief)
	} else if showRemote && sameNormalizedPath(repoPath, statusRoot) {
		// Primary status root only (may be Dir ../.. when cwd is a subdir — not Dir==".").
		remoteLine, err := formatStatusRemoteLine(repoPath, meta.Branch, scanColorEnabled, fetchEnabled, meta.Status == "clean")
		if err != nil {
			return err
		}
		fmt.Println(remoteLine)
	}
	return nil
}

// applyWrkStatusWithLinkedSkip recomputes Meta.Status including untracked files
// but excluding in-tree linked worktree paths (same skip as wrk --projects).
// Git reports those checkouts as ?? under the main tree; they must not mark
// the parent dirty when they are reported as their own status blocks.
func applyWrkStatusWithLinkedSkip(repoPath string, meta *checkout.Meta) {
	if meta == nil || meta.Error != "" {
		return
	}
	var skip map[string]struct{}
	if worktree.IsMainRepo(repoPath) {
		if linked, err := worktree.ListLinked(repoPath); err == nil && len(linked) > 0 {
			skip = skipUntrackedRelPaths(repoPath, linked)
		}
	}
	if len(skip) == 0 {
		// No linked paths to exclude; Enrich status already includes untracked.
		return
	}
	counts, err := gitProjectStatusCountsWithSkip(repoPath, skip)
	if err != nil {
		return
	}
	meta.Status = status.FormatWrk(counts)
}

func formatStatusRemoteLine(mainRepoPath, currentBranch string, colorEnabled bool, fetchEnabled bool, isClean bool) (string, error) {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return "", err
	}
	if upstream == "" {
		return "Remote:       (no upstream)", nil
	}
	if fetchEnabled {
		if err := gitFetchUpstreamQuietNoOptionalLocks(mainRepoPath, upstream); err != nil {
			return "Remote:       error: " + err.Error(), nil
		}
	}
	remoteColor := colorEnabled && isClean
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return "Remote:       error: " + err.Error(), nil
	}
	return "Remote:       " + FormatRemoteBrief(result, remoteColor), nil
}

func projectBlockUsesColor(colorEnabled bool, counts status.WrkCounts, remoteRelation git.BranchRelation, dirtyWorktrees, worktreeErrors int) bool {
	if !colorEnabled {
		return false
	}
	if counts.Added != 0 || counts.Changed != 0 || counts.Renamed != 0 || counts.Deleted != 0 {
		return true
	}
	if dirtyWorktrees > 0 {
		return true
	}
	if worktreeErrors > 0 {
		return true
	}
	switch remoteRelation {
	case git.BranchRelationAIsAncestorOfB, git.BranchRelationBIsAncestorOfA, git.BranchRelationDiverged:
		return true
	default:
		return false
	}
}

func statusBlockUsesColor(colorEnabled bool, counts status.WrkCounts, hasMaster bool, masterRelation git.BranchRelation) bool {
	if !colorEnabled {
		return false
	}
	if counts.Added != 0 || counts.Changed != 0 || counts.Renamed != 0 || counts.Deleted != 0 {
		return true
	}
	if counts.Added == 0 && counts.Changed == 0 && counts.Renamed == 0 && counts.Deleted == 0 {
		return true
	}
	if hasMaster {
		switch masterRelation {
		case git.BranchRelationSame, git.BranchRelationAIsAncestorOfB, git.BranchRelationBIsAncestorOfA, git.BranchRelationDiverged:
			return true
		}
	}
	return false
}

func masterBriefForRepo(repoPath, wtBranch string, colorEnabled bool) (string, git.BranchRelation, error) {
	mainRepo, err := worktree.ReadMainRepo(repoPath)
	if err != nil {
		return "", 0, err
	}
	mainBranch, err := gitcmd.Run(context.Background(), mainRepo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", 0, err
	}
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		return "", 0, err
	}
	return FormatMasterBrief(result, colorEnabled), result.Relation, nil
}

func formatCompareWithRemote(mainRepoPath, currentBranch string, colorEnabled bool, fetchEnabled bool) (string, error) {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return "", err
	}
	if upstream == "" {
		return "Remote:       (no upstream)", nil
	}
	if fetchEnabled {
		if err := gitFetchUpstreamQuietNoOptionalLocks(mainRepoPath, upstream); err != nil {
			return "", err
		}
	}
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return "", err
	}
	return "Remote:       " + FormatRemoteBrief(result, colorEnabled), nil
}

func gitUpstreamRef(repoPath string) (string, error) {
	upstream, ok, err := gitcmd.RunOptional(context.Background(), repoPath, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err != nil || !ok {
		// Match legacy gitOutput behavior: missing upstream is not an error.
		return "", nil
	}
	return upstream, nil
}

func gitWorktreeIsClean(repoPath string) (bool, error) {
	return worktree.IsCleanWrk(repoPath)
}

func gitCombinedOutput(repoPath string, args ...string) ([]byte, error) {
	cmd := gitCommandDir(repoPath, args...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	return cmd.CombinedOutput()
}

func gitCombinedOutputError(repoPath string, args ...string) string {
	out, err := gitCombinedOutput(repoPath, args...)
	if err == nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitOutputNoOptionalLocks(repoPath string, args ...string) (string, error) {
	return gitcmd.Run(context.Background(), repoPath, args...)
}

func parseProjectStatusCounts(out string, skipUntracked map[string]struct{}) status.WrkCounts {
	if len(skipUntracked) == 0 {
		return status.ParsePorcelainWrk(out)
	}
	var filtered strings.Builder
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			path := strings.TrimSpace(line[3:])
			path = strings.TrimSuffix(path, "/")
			if _, ok := skipUntracked[path]; ok {
				continue
			}
		}
		if filtered.Len() > 0 {
			filtered.WriteByte('\n')
		}
		filtered.WriteString(line)
	}
	return status.ParsePorcelainWrk(filtered.String())
}

func gitFetchQuiet(repoPath string) error {
	cmd := gitCommand("-C", repoPath, "fetch", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func gitFetchQuietNoOptionalLocks(repoPath string) error {
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, "fetch", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func gitFetchUpstreamQuietNoOptionalLocks(repoPath, upstream string) error {
	remote, branch, ok := strings.Cut(upstream, "/")
	if !ok || remote == "" || branch == "" {
		return gitFetchQuietNoOptionalLocks(repoPath)
	}
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, "fetch", "--quiet", remote, branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func formatStatusCounts(counts status.WrkCounts, colorEnabled bool, greenClean bool) string {
	return formatStatusText(status.FormatWrk(counts), colorEnabled, greenClean)
}

func formatStatusText(plain string, colorEnabled bool, greenClean bool) string {
	if plain == "clean" {
		if colorEnabled && greenClean {
			return colorize("clean", ansiGreen)
		}
		return "clean"
	}
	if !colorEnabled {
		return plain
	}
	if !strings.HasPrefix(plain, "dirty (") || !strings.HasSuffix(plain, ")") {
		return plain
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(plain, "dirty ("), ")")
	return colorize("dirty", ansiRed) + " (" + colorizeStatusSegments(inner) + ")"
}

func colorizeStatusSegments(inner string) string {
	parts := strings.Split(inner, ", ")
	for i, part := range parts {
		fields := strings.SplitN(part, " ", 2)
		if len(fields) != 2 {
			continue
		}
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		if n > 0 {
			parts[i] = formatStatusCountSegment(n, fields[1])
		} else {
			parts[i] = colorize(part, ansiGrey)
		}
	}
	return strings.Join(parts, ", ")
}

func formatStatusCountSegment(n int, kind string) string {
	s := fmt.Sprintf("%d %s", n, kind)
	if n > 0 {
		return colorize(s, ansiRed)
	}
	return colorize(s, ansiGrey)
}
