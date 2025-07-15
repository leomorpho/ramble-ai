//go:build darwin && !production
// +build darwin,!production

package binaries

// FFmpegBinary is empty in development/test mode
var FFmpegBinary []byte

const FFmpegExtension = ""