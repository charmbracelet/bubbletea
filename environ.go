package tea

import uv "github.com/charmbracelet/ultraviolet"

// EnvMsg is a message that represents the environment variables of the
// program. This is useful for getting the environment variables of programs
// running in a remote session like SSH. In that case, using [os.Getenv] would
// return the server's environment variables, not the client's.
//
// This message is sent to the program when it starts.
//
// Example:
//
//	switch msg := msg.(type) {
//	case EnvMsg:
//	  // What terminal type is being used?
//	  term := msg.Getenv("TERM")
//	}
type EnvMsg uv.Environ

// Getenv returns the value of the environment variable named by the key. If
// the variable is not present in the environment, the value returned will be
// the empty string.
func (msg EnvMsg) Getenv(key string) (v string) {
	return uv.Environ(msg).Getenv(key)
}

// LookupEnv retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned and the boolean is true. Otherwise the returned value will be
// empty and the boolean will be false.
func (msg EnvMsg) LookupEnv(key string) (s string, v bool) {
	return uv.Environ(msg).LookupEnv(key)
}
