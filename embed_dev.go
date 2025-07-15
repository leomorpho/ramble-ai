//go:build !production
// +build !production

package main

import (
	"embed"
)

// In development/test mode, provide an empty embed.FS
var assets embed.FS

func init() {
	// Create a minimal filesystem structure for development
	// This prevents errors when Wails tries to access the assets
	// In dev mode, Wails will use the vite dev server instead
}