package tea

import (
	"strings"
)

// environ is a slice of strings that represents the environment variables of
// the program.
type environ []string

// Getenv returns the value of the environment variable named by the key. If
// the variable is not present in the environment, the value returned will be
// the empty string.
func (p environ) Getenv(key string) (v string) {
	v, _ = p.LookupEnv(key)
	return
}

// LookupEnv retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned and the boolean is true. Otherwise the returned value will be
// empty and the boolean will be false.
func (p environ) LookupEnv(key string) (s string, v bool) {
	for i := len(p) - 1; i >= 0; i-- {
		if strings.HasPrefix(p[i], key+"=") {
			s = strings.TrimPrefix(p[i], key+"=")
			v = true
			break
		}
	}
	return
}

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
type EnvMsg environ

// Getenv returns the value of the environment variable named by the key. If
// the variable is not present in the environment, the value returned will be
// the empty string.
func (msg EnvMsg) Getenv(key string) (v string) {
	return environ(msg).Getenv(key)
}

// LookupEnv retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty)
// is returned and the boolean is true. Otherwise the returned value will be
// empty and the boolean will be false.
func (msg EnvMsg) LookupEnv(key string) (s string, v bool) {
	return environ(msg).LookupEnv(key)
}

// termType represents a terminal type. This is the value found in the TERM
// environment variable. It is used to determine the capabilities and type of
// the terminal.
type termType string

// Name returns the name of the terminal type. This is the first part of the
// TERM environment variable split by the first dash (-).
func (t termType) Name() string {
	parts := strings.Split(string(t), "-")
	return parts[0]
}

// String returns the string representation of the terminal type. This is the
// full value of the TERM environment variable.
func (t termType) String() string {
	return string(t)
}

// Is returns whether the terminal type name is equal to the given name.
func (t termType) Is(name string) bool {
	return strings.Contains(t.Name(), name)
}
