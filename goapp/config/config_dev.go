//go:build !production
// +build !production

package config

const (
	// Development configuration - can be overridden by environment variables
	PRODUCTION_BUILD      = false
	USE_REMOTE_AI_BACKEND = false
	REMOTE_AI_BACKEND_URL = "http://localhost:8090"
	RAMBLE_FRONTEND_URL   = "https://ramble.goosebyteshq.com"
)