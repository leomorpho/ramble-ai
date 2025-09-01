//go:build production
// +build production

package config

const (
	// Production configuration - compiled into the binary
	PRODUCTION_BUILD      = true
	USE_REMOTE_AI_BACKEND = true
	REMOTE_AI_BACKEND_URL = "https://api.ramble.goosebyteshq.com"
	RAMBLE_FRONTEND_URL   = "https://ramble.goosebyteshq.com"
)