//go:build windows && !production
// +build windows,!production

package binaries

// FFmpegBinary is empty in development/test mode
var FFmpegBinary []byte

const FFmpegExtension = ".exe"