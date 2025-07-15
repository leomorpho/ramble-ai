//go:build windows
// +build windows

package binaries

import _ "embed"

//go:embed static/ffmpeg-windows-amd64.exe
var FFmpegBinary []byte

const FFmpegExtension = ".exe"