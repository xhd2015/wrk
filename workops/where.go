package workops

// WhereMain resolves the main repository absolute path for a checkout.
// Linked worktree → main abs; main checkout → cleaned self path.
func WhereMain(checkout string) (mainAbs string, err error) {
	return resolveMainRepo(checkout)
}
