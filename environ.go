package tea

import "strings"

// getenv is a function that returns the value of the environment variable named
// by the key. If the variable is not present in the environment, the value
// returned will be the empty string.
// This function traverses the environment variables in reverse order, so that
// the last value set for the key is the one returned.
func (p *Program) getenv(key string) (v string) {
	for i := len(p.environ) - 1; i >= 0; i-- {
		if strings.HasPrefix(p.environ[i], key+"=") {
			v = strings.TrimPrefix(p.environ[i], key+"=")
			break
		}
	}
	return
}
