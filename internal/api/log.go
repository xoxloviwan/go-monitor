package api

import "os"

// LogFatal logs an error message and exits the program.
func LogFatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
