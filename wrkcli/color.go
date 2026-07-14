package wrkcli

const (
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiOrange = "\x1b[33m"
	ansiGrey   = "\x1b[90m"
	ansiReset  = "\x1b[0m"
)

func colorize(s, code string) string {
	return code + s + ansiReset
}