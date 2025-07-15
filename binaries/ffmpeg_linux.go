//go:build linux && production
// +build linux,production

package binaries

import _ "embed"

//go:embed static/ffmpeg-linux-amd64
var FFmpegBinary []byte

const FFmpegExtension = ""
