//go:build production
// +build production

package main

import "embed"

//go:embed all:frontend/build
var assets embed.FS