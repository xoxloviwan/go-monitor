package api

import "os"

func LogFatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
