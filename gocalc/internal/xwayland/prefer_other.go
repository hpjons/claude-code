//go:build !linux

package xwayland

// Prefer is a no-op off Linux, where Wails doesn't use WebKitGTK.
func Prefer() {}
