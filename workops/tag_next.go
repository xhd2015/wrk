package workops

import (
	"context"
	"fmt"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

// TagNextResult is the multi-scope outcome of TagNextAll.
type TagNextResult struct {
	// Tags are planned next tag names (DryRun) or created tag names (apply),
	// covering all scopes that planned a NextTag (root first when present).
	Tags []string
	// MainRepo is the resolved main repository absolute path.
	MainRepo string
}

// TagNext plans or applies the next release tag(s) at main tip and returns the
// primary tag string (first of TagNextAll.Tags — root-primary when present).
// DryRun returns the planned primary tag without creating refs.
// Kept as a BC helper for callers that need a single string.
func TagNext(ctx context.Context, opts TagNextOptions) (tag string, err error) {
	res, err := TagNextAll(ctx, opts)
	if err != nil {
		return "", err
	}
	if len(res.Tags) == 0 {
		return "", fmt.Errorf("workops: TagNext: no next tag planned")
	}
	return res.Tags[0], nil
}

// TagNextAll plans or applies next release tags for all scopes at main tip
// (parity with wrkcli runTagNextAtResult core). DryRun creates no tag refs.
// HeadRef (optional on TagNextOptions) defaults to "HEAD".
func TagNextAll(ctx context.Context, opts TagNextOptions) (TagNextResult, error) {
	var out TagNextResult
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if opts.Checkout == "" {
		return out, fmt.Errorf("workops: TagNext requires Checkout")
	}

	mainRepo, err := resolveMainRepo(opts.Checkout)
	if err != nil {
		return out, err
	}
	out.MainRepo = mainRepo

	headRef := opts.HeadRef
	if headRef == "" {
		headRef = "HEAD"
	}

	plan, _, err := tagscope.Plan(mainRepo, headRef)
	if err != nil {
		return out, err
	}

	planned := plannedTagNames(plan)
	if opts.DryRun {
		// Empty plan is OK for TagNextAll (CLI prints skip plan); TagNext
		// (singular) still errors when primary is empty.
		out.Tags = planned
		return out, nil
	}

	result, err := tagscope.Apply(mainRepo, plan, headRef, tagscope.ApplyOptions{
		DryRun: false,
		Push:   false,
	})
	if err != nil {
		return out, err
	}
	// Empty Created is OK when no scope planned a NextTag (CLI skip plan).
	// TagNext (singular) still errors when Tags is empty.
	out.Tags = result.Created
	return out, nil
}

// plannedTagNames returns NextTag values from the plan, root scope first then
// nested scopes in plan decision order.
func plannedTagNames(plan tagscope.ChangePlan) []string {
	var root, nested []string
	for _, d := range plan.Decisions {
		if d.NextTag == "" {
			continue
		}
		if d.Scope.PathPrefix == "" {
			root = append(root, d.NextTag)
		} else {
			nested = append(nested, d.NextTag)
		}
	}
	return append(root, nested...)
}
