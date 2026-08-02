package workops

import (
	"context"
	"fmt"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

// TagNext plans or applies the next root release tag at main tip.
// DryRun returns the planned next tag string without creating refs.
// Returns the singular root-scope next tag (e.g. v0.0.2 after v0.0.1).
func TagNext(ctx context.Context, opts TagNextOptions) (tag string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if opts.Checkout == "" {
		return "", fmt.Errorf("workops: TagNext requires Checkout")
	}

	// Plan against main object DB (WhereMain-equivalent resolve).
	mainRepo, err := resolveMainRepo(opts.Checkout)
	if err != nil {
		return "", err
	}

	plan, _, err := tagscope.Plan(mainRepo, "HEAD")
	if err != nil {
		return "", err
	}

	rootTag := rootNextTag(plan)
	if rootTag == "" {
		// Fall back to first planned tag if root scope has no NextTag.
		for _, d := range plan.Decisions {
			if d.NextTag != "" {
				rootTag = d.NextTag
				break
			}
		}
	}

	if opts.DryRun {
		if rootTag == "" {
			return "", fmt.Errorf("workops: TagNext dry-run: no next tag planned")
		}
		return rootTag, nil
	}

	result, err := tagscope.Apply(mainRepo, plan, "HEAD", tagscope.ApplyOptions{
		DryRun: false,
		Push:   false,
	})
	if err != nil {
		return "", err
	}
	if rootTag != "" {
		return rootTag, nil
	}
	if len(result.Created) > 0 {
		return result.Created[0], nil
	}
	return "", fmt.Errorf("workops: TagNext: no tag created")
}

// rootNextTag returns the planned NextTag for the root scope (empty PathPrefix).
func rootNextTag(plan tagscope.ChangePlan) string {
	for _, d := range plan.Decisions {
		if d.Scope.PathPrefix == "" && d.NextTag != "" {
			return d.NextTag
		}
	}
	return ""
}
