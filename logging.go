package tea

import (
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
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
