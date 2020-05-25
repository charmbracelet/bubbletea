// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"log"
	"log/syslog"
	"os"
)

// LogToFile sets up default logging to log to a file This is helpful as we
// can't print to the terminal since our TUI is occupying it. If the file
// doesn't exist it will be created.
//
// Don't forget to close the file when you're done with it.
//
//   f, err := LogToFile("debug.log", "debug")
//   if err != nil {
//		fmt.Println("fatal:", err)
//		os.Exit(1)
//   }
//   defer f.Close()
func LogToFile(path string, prefix string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	return f, nil
}

// UseSysLog sets up logging to log the system log. This becomes helpful when
// debugging since we can't print to the terminal since our TUI is occupying it.
//
// On macOS this is a just a matter of: tail -f /var/log/system.log
// On Linux this varies depending on distribution.
func UseSysLog(programName string) error {
	l, err := syslog.New(syslog.LOG_NOTICE, programName)
	if err != nil {
		return err
	}
	log.SetOutput(l)
	return nil
}
