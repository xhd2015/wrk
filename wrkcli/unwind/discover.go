package unwind

import (
	"context"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func discoverStatusRepos(ctx context.Context, root string) ([]scan_repo.Repo, error) {
	result, err := scan_repo.Scan(ctx, scan_repo.Options{
		Roots:   []string{root},
		NoCache: true,
	})
	return result.Repos, err
}
