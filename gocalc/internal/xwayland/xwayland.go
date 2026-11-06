// Package xwayland works around WebKitGTK rendering a completely blank
// window under native Wayland by steering GTK onto XWayland instead.
//
// This was observed on the author's machine (COSMIC desktop); neither
// WEBKIT_DISABLE_DMABUF_RENDERER nor WEBKIT_DISABLE_COMPOSITING_MODE fixed
// it, while XWayland renders correctly every time. The workaround is only
// built on Linux (see prefer_linux.go) and can be turned off by setting
// GOCALC_NATIVE_WAYLAND=1.
package xwayland

import "regexp"

// OptOutEnv disables the workaround when set to any non-empty value other
// than "0".
const OptOutEnv = "GOCALC_NATIVE_WAYLAND"

// desktopDefaultBackend is the GDK_BACKEND value COSMIC's session sets by
// default. It makes GTK try (and succeed at connecting to, just not
// rendering into) Wayland first, so it's overridden like an unset value.
// Any other value was presumably chosen by the user and is respected.
const desktopDefaultBackend = "wayland,x11"

var x11SocketPattern = regexp.MustCompile(`^X(\d+)$`)

// plan decides which environment variables to set, given a way to look up
// the current environment and the entry names in /tmp/.X11-unix. It
// returns nil when nothing should change.
//
// GDK_BACKEND=x11 is only chosen when an X server is actually reachable
// (an existing DISPLAY, or a live socket in /tmp/.X11-unix): forcing it
// with no reachable X server crashes GTK outright.
func plan(lookup func(string) (string, bool), x11Sockets []string) map[string]string {
	if v, _ := lookup(OptOutEnv); v != "" && v != "0" {
		return nil
	}
	if backend, _ := lookup("GDK_BACKEND"); backend != "" && backend != desktopDefaultBackend {
		return nil
	}

	if display, _ := lookup("DISPLAY"); display != "" {
		return map[string]string{"GDK_BACKEND": "x11"}
	}
	for _, name := range x11Sockets {
		if m := x11SocketPattern.FindStringSubmatch(name); m != nil {
			return map[string]string{"GDK_BACKEND": "x11", "DISPLAY": ":" + m[1]}
		}
	}
	return nil
}
