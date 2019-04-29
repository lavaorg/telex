package logger

import (
	"github.com/lavaorg/lrt/mlog"
)

// force mlog package; which installs into go log
// currently telex uses standard go logging; it will now pass through mlog

// SetupLogging configures the logging output.
//   debug   will set the log level to DEBUG
//   quiet   will set the log level to ERROR
//   logfile will direct the logging output to a file. Empty string is
//           interpreted as stderr. If there is an error opening the file the
//           logger will fallback to stderr.
func SetupLogging(debug, quiet bool, logfile string) {

	mlog.EnableDebug(debug)
	mlog.EnableDebug(!quiet)
}
