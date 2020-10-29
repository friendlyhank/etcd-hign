package osutil

import "os"

// Exit calls os.Exit
func Exit(code int) {
	os.Exit(code)
}
