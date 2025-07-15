//go:build darwin
// +build darwin

package binaries

import _ "embed"

//go:embed static/ffmpeg-darwin-amd64
var FFmpegBinary []byte

const FFmpegExtension = ""
