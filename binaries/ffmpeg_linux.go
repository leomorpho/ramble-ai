//go:build linux
// +build linux

package binaries

import _ "embed"

//go:embed static/ffmpeg-linux-amd64
var FFmpegBinary []byte

const FFmpegExtension = ""