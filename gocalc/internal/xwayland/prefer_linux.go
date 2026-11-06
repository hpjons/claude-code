//go:build linux

package xwayland

import "os"

// Prefer points GTK at XWayland when the workaround applies (see plan). It
// must run before Wails initializes GTK.
func Prefer() {
	var sockets []string
	if entries, err := os.ReadDir("/tmp/.X11-unix"); err == nil {
		for _, e := range entries {
			sockets = append(sockets, e.Name())
		}
	}
	for k, v := range plan(os.LookupEnv, sockets) {
		os.Setenv(k, v)
	}
}
