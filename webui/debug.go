package webui

import "log"

var debugEnabled bool

func SetDebug(enabled bool) {
	debugEnabled = enabled
}

func DebugEnabled() bool {
	return debugEnabled
}

func debugLog(format string, args ...any) {
	if !debugEnabled {
		return
	}
	log.Printf("WebUI: "+format, args...)
}
