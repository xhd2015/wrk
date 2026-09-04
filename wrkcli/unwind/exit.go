package unwind

// ExitCodeError is a silent non-zero exit (verify FAIL). wrkcli maps this to
// wrkcli.ExitCodeError so Capture/main do not print Error:.
type ExitCodeError struct {
	Code int
}

func (e ExitCodeError) Error() string { return "" }
