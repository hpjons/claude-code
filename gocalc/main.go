package main

import (
	"embed"
	"os"
	"regexp"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var x11SocketPattern = regexp.MustCompile(`^X(\d+)$`)

// preferXWayland forces GTK onto XWayland instead of native Wayland.
// WebKitGTK renders a completely blank window under native Wayland on this
// machine (confirmed: neither WEBKIT_DISABLE_DMABUF_RENDERER nor
// WEBKIT_DISABLE_COMPOSITING_MODE fix it), while XWayland renders correctly
// every time. COSMIC's session sets GDK_BACKEND=wayland,x11 by default,
// which makes GTK try (and succeed at connecting to, just not rendering
// into) Wayland first, so this always overrides it to x11 rather than
// deferring to whatever's already set. It only does so when an X server is
// actually reachable (via an existing DISPLAY, or by finding a live socket
// in /tmp/.X11-unix) — forcing GDK_BACKEND=x11 with no reachable X server
// crashes GTK outright.
func preferXWayland() {
	if _, set := os.LookupEnv("DISPLAY"); set {
		os.Setenv("GDK_BACKEND", "x11")
		return
	}

	entries, err := os.ReadDir("/tmp/.X11-unix")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if m := x11SocketPattern.FindStringSubmatch(entry.Name()); m != nil {
			os.Setenv("DISPLAY", ":"+m[1])
			os.Setenv("GDK_BACKEND", "x11")
			return
		}
	}
}

func main() {
	preferXWayland()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "gocalc",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
