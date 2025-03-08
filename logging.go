package tea

import (
	"fmt"
	"io"
	"log"
	"os"
	"unicode"
)

// LogToFile sets up default logging to log to a file. This is helpful as we
// can't print to the terminal since our TUI is occupying it. If the file
// doesn't exist it will be created.
//
// Don't forget to close the file when you're done with it.
//
//	  f, err := LogToFile("debug.log", "debug")
//	  if err != nil {
//			fmt.Println("fatal:", err)
//			os.Exit(1)
//	  }
//	  defer f.Close()
func LogToFile(path string, prefix string) (*os.File, error) {
	return LogToFileWith(path, prefix, log.Default())
}

// LogOptionsSetter is an interface implemented by stdlib's log and charm's log
// libraries.
type LogOptionsSetter interface {
	SetOutput(io.Writer)
	SetPrefix(string)
}

// LogToFileWith does allows to call LogToFile with a custom LogOptionsSetter.
func LogToFileWith(path string, prefix string, log LogOptionsSetter) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600) //nolint:mnd
	if err != nil {
		return nil, fmt.Errorf("error opening file for logging: %w", err)
	}
	log.SetOutput(f)

	// Add a space after the prefix if a prefix is being specified and it
	// doesn't already have a trailing space.
	if len(prefix) > 0 {
		finalChar := prefix[len(prefix)-1]
		if !unicode.IsSpace(rune(finalChar)) {
			prefix += " "
		}
	}
	log.SetPrefix(prefix)

	return f, nil
}
