// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package boba

import (
	"log"
	"log/syslog"
)

// UseSysLog sets up logging to log the system log. This becomes helpful when
// debugging since we can't easily print to the terminal since our TUI is
// occupying it!
//
// On macOS this is a just a matter of: tail -f /var/log/system.log
func UseSysLog(programName string) error {
	l, err := syslog.New(syslog.LOG_NOTICE, programName)
	if err != nil {
		return err
	}
	log.SetOutput(l)
	return nil
}
