package wrkcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// composeShipOpts configures the post-land / activeRoot ship wave:
// (tag-next → push) ‖ sync ‖ reinstall-local.
type composeShipOpts struct {
	MainPath   string
	SourcePath string // linked worktree path (same-name origin update after push)
	Result     *worktree.MergeBackResult
	SameName   sameNameRemoteSnapshot

	DryRun           bool
	WithSync         bool
	WithTagNext      bool
	WithPush         bool
	ForcePush        bool
	WithReinstall    bool
	ReinstallUseMain bool // true after done/merge-back; false on activeRoot
	PretendMainAt    string
	Color            bool
	NoColor          bool
	Stage            *composeStageWriter // optional; when set, marks the ship stage
	StageIndex       int
	SoftSkipEmptyMod bool // done/merge-back: empty go.mod plan does not fail ship
}

type shipLaneID string

const (
	shipLaneTagPush   shipLaneID = "tag-next+push"
	shipLaneSync      shipLaneID = "sync"
	shipLaneReinstall shipLaneID = "reinstall-local"
)

type shipLaneResult struct {
	ID      shipLaneID
	Summary string
	Capture string
	Err     error
	Elapsed time.Duration
}

func shipStageIO(st *composeStageWriter) (out, errW io.Writer) {
	if st != nil {
		return st.Out(), st.Err()
	}
	return os.Stdout, os.Stderr
}

// runComposeShip runs optional ship stages. Dry-run stays serial and ordered
// (sync → tag-next → push → reinstall) for assertability. Apply runs enabled
// lanes concurrently with fail-fast cancel, unwind-style progress, then flushes
// full sync/tag+push bodies; reinstall stays summary-only.
// All body output under the open ship marker is kind-aligned via Stage.Out/Err.
func runComposeShip(opts composeShipOpts) error {
	if !opts.WithSync && !opts.WithTagNext && !opts.WithPush && !opts.WithReinstall {
		return nil
	}
	if opts.MainPath == "" {
		return fmt.Errorf("wrk: ship missing main path")
	}
	if opts.DryRun {
		return runComposeShipDry(opts)
	}
	return runComposeShipApply(opts)
}

func runComposeShipDry(opts composeShipOpts) error {
	st := opts.Stage
	out, errW := shipStageIO(st)
	if st != nil && opts.StageIndex > 0 {
		st.mark(opts.StageIndex, "ship")
		st.detail("dry-run · serial")
	}
	blankBefore := func() {
		fmt.Fprintln(out)
	}
	if opts.WithSync {
		blankBefore()
		if _, err := runSyncOpts(opts.MainPath, syncOpts{
			DryRun:        true,
			PretendMainAt: opts.PretendMainAt,
			Color:         opts.Color,
			NoColor:       opts.NoColor,
			Out:           out,
			Err:           errW,
		}); err != nil {
			return err
		}
	}
	var createdTags []string
	if opts.WithTagNext {
		blankBefore()
		headRef := "HEAD"
		if opts.PretendMainAt != "" {
			headRef = opts.PretendMainAt
		}
		tagRes, err := runTagNextAtResultTo(opts.MainPath, headRef, true, false, false, out)
		if err != nil {
			return err
		}
		createdTags = tagRes.Tags
	}
	if opts.WithPush {
		blankBefore()
		var tags []string
		if opts.WithTagNext {
			tags = createdTags
		}
		if err := runPushMainWrite(opts.MainPath, true, opts.ForcePush, tags, true, out); err != nil {
			return err
		}
		maybeUpdateSameNameOriginBranch(opts.MainPath, opts.SourcePath, opts.Result, opts.SameName, true)
	}
	if opts.WithReinstall {
		blankBefore()
		if _, err := runComposeShipReinstall(opts, out, errW); err != nil {
			return err
		}
	}
	return nil
}

func runComposeShipApply(opts composeShipOpts) error {
	type lane struct {
		id  shipLaneID
		run func(ctx context.Context, out, errW io.Writer) (summary string, err error)
	}
	var lanes []lane
	if opts.WithTagNext || opts.WithPush {
		lanes = append(lanes, lane{
			id: shipLaneTagPush,
			run: func(ctx context.Context, out, errW io.Writer) (string, error) {
				return runShipTagPushLane(ctx, opts, out)
			},
		})
	}
	if opts.WithSync {
		lanes = append(lanes, lane{
			id: shipLaneSync,
			run: func(ctx context.Context, out, errW io.Writer) (string, error) {
				return runShipSyncLane(ctx, opts, out, errW)
			},
		})
	}
	if opts.WithReinstall {
		lanes = append(lanes, lane{
			id: shipLaneReinstall,
			run: func(ctx context.Context, out, errW io.Writer) (string, error) {
				return runShipReinstallLane(ctx, opts, out, errW)
			},
		})
	}
	if len(lanes) == 0 {
		return nil
	}

	st := opts.Stage
	out, errW := shipStageIO(st)
	indent := ""
	if st != nil {
		indent = st.Indent()
	}
	if st != nil && opts.StageIndex > 0 {
		name := "ship"
		if len(lanes) > 1 {
			name = "ship · concurrent"
		}
		st.mark(opts.StageIndex, name)
	}

	// Single lane: live body through kind-aligned writers; reinstall captures.
	if len(lanes) == 1 {
		l := lanes[0]
		if l.id == shipLaneReinstall {
			var sink strings.Builder
			summary, err := l.run(context.Background(), &sink, &sink)
			if err != nil {
				dumpShipCaptureTo(errW, string(l.id), sink.String())
				return fmt.Errorf("wrk: %s failed: %w", l.id, err)
			}
			if summary != "" {
				fmt.Fprintln(out, summary)
			}
			return nil
		}
		if _, err := l.run(context.Background(), out, errW); err != nil {
			return fmt.Errorf("wrk: %s failed: %w", l.id, err)
		}
		return nil
	}

	ids := make([]string, len(lanes))
	for i, l := range lanes {
		ids[i] = string(l.id)
	}
	prog := newShipProgress(indent, resolveStderrColor(opts.Color, opts.NoColor), ids)
	prog.Begin()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pathMu sync.Mutex
	results := make([]shipLaneResult, len(lanes))
	var wg sync.WaitGroup
	var failOnce sync.Once
	var firstErr error

	for i, l := range lanes {
		i, l := i, l
		wg.Add(1)
		go func() {
			defer wg.Done()
			sink := prog.Sink(string(l.id))
			prog.Start(string(l.id))
			start := time.Now()
			run := l.run
			if l.id == shipLaneTagPush || l.id == shipLaneSync {
				inner := run
				run = func(ctx context.Context, out, errW io.Writer) (string, error) {
					pathMu.Lock()
					defer pathMu.Unlock()
					return inner(ctx, out, errW)
				}
			}
			summary, err := run(ctx, sink, sink)
			results[i] = shipLaneResult{
				ID:      l.id,
				Summary: summary,
				Capture: prog.Capture(string(l.id)),
				Err:     err,
				Elapsed: time.Since(start),
			}
			prog.Finish(string(l.id), summary, err)
			if err != nil {
				failOnce.Do(func() {
					firstErr = err
					cancel()
				})
			}
		}()
	}
	wg.Wait()
	prog.Close()

	if firstErr != nil {
		for _, r := range results {
			if r.Err != nil && !isContextCanceled(r.Err) {
				dumpShipCaptureTo(errW, string(r.ID), r.Capture)
				return fmt.Errorf("wrk: %s failed: %w", r.ID, r.Err)
			}
		}
		return fmt.Errorf("wrk: ship failed: %w", firstErr)
	}

	// Flush full bodies for sync/tag+push via kind-aligned stdout; reinstall summary-only.
	for _, r := range results {
		switch r.ID {
		case shipLaneTagPush, shipLaneSync:
			body := strings.TrimSpace(r.Capture)
			if body == "" {
				continue
			}
			fmt.Fprintln(out)
			fmt.Fprintln(out, body)
		}
	}
	return nil
}

func isContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	return err == context.Canceled || strings.Contains(err.Error(), "context canceled")
}

func runShipTagPushLane(ctx context.Context, opts composeShipOpts, out io.Writer) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var createdTags []string
	var parts []string
	if opts.WithTagNext {
		tagRes, err := runTagNextAtResultTo(opts.MainPath, "HEAD", false, false, false, out)
		if err != nil {
			return "", err
		}
		createdTags = tagRes.Tags
		switch {
		case len(createdTags) == 1:
			parts = append(parts, "tagged "+createdTags[0])
		case len(createdTags) > 1:
			parts = append(parts, fmt.Sprintf("tagged %d", len(createdTags)))
		default:
			parts = append(parts, "tag-next (none)")
		}
	}
	if err := ctx.Err(); err != nil {
		return strings.Join(parts, " · "), err
	}
	if opts.WithPush {
		var tags []string
		if opts.WithTagNext {
			tags = createdTags
		}
		if err := runPushMainWrite(opts.MainPath, false, opts.ForcePush, tags, true, out); err != nil {
			return strings.Join(parts, " · "), err
		}
		maybeUpdateSameNameOriginBranch(opts.MainPath, opts.SourcePath, opts.Result, opts.SameName, false)
		parts = append(parts, "pushed")
	}
	return strings.Join(parts, " · "), nil
}

func runShipSyncLane(ctx context.Context, opts composeShipOpts, out, errW io.Writer) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	res, err := runSyncOpts(opts.MainPath, syncOpts{
		DryRun:  false,
		Color:   opts.Color,
		NoColor: opts.NoColor,
		Out:     out,
		Err:     errW,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("synced: %d into main, %d into worktrees, %d skipped",
		res.IntoMain, res.IntoWT, res.Skipped), nil
}

func runShipReinstallLane(ctx context.Context, opts composeShipOpts, out, errW io.Writer) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return runComposeShipReinstall(opts, out, errW)
}

func runComposeShipReinstall(opts composeShipOpts, out, errW io.Writer) (string, error) {
	st, err := runReinstallLocalExTo(opts.MainPath, opts.DryRun, opts.ReinstallUseMain, opts.Color, opts.NoColor, nil, out, errW)
	if err == nil {
		return fmt.Sprintf("reinstalled %d, skipped %d, failed %d", st.Reinstalled, st.Skipped, st.Failed), nil
	}
	if opts.SoftSkipEmptyMod &&
		(strings.Contains(err.Error(), "no go.mod modules found") ||
			strings.Contains(err.Error(), "no go.mod found")) {
		if opts.DryRun {
			fmt.Fprintf(errW, "would: skip reinstall-local (%s)\n", err.Error())
		} else {
			fmt.Fprintf(errW, "skip reinstall-local: %s\n", err.Error())
		}
		return "skipped (no go.mod)", nil
	}
	return "", err
}

func formatShipElapsed(d time.Duration) string {
	if d < time.Second {
		ms := d.Milliseconds()
		if ms < 1 && d > 0 {
			ms = 1
		}
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d / time.Minute)
	secs := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

func dumpShipCapture(id, capture string) {
	dumpShipCaptureTo(os.Stderr, id, capture)
}

func dumpShipCaptureTo(errW io.Writer, id, capture string) {
	if errW == nil {
		errW = os.Stderr
	}
	capture = strings.TrimSpace(capture)
	fmt.Fprintf(errW, "---- %s ----\n", id)
	if capture == "" {
		fmt.Fprintln(errW, "(no output)")
		return
	}
	fmt.Fprintln(errW, capture)
}
