//go:build linux && !production
// +build linux,!production

package binaries

// FFmpegBinary is empty in development/test mode
var FFmpegBinary []byte

const FFmpegExtension = ""