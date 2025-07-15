//go:build windows && production
// +build windows,production

package binaries

import _ "embed"

//go:embed static/ffmpeg-windows-amd64.exe
var FFmpegBinary []byte

const FFmpegExtension = ".exe"
