package config

import "os"

// IsProduction returns true if the app is built with production tag
func IsProduction() bool {
	// This will be set by the build tags
	return PRODUCTION_BUILD
}

// GetRemoteBackendURL returns the backend URL, allowing env override in development
func GetRemoteBackendURL() string {
	if !IsProduction() {
		if envURL := os.Getenv("REMOTE_AI_BACKEND_URL"); envURL != "" {
			return envURL
		}
	}
	return REMOTE_AI_BACKEND_URL
}

// GetFrontendURL returns the frontend URL, allowing env override in development
func GetFrontendURL() string {
	if !IsProduction() {
		if envURL := os.Getenv("RAMBLE_FRONTEND_URL"); envURL != "" {
			return envURL
		}
	}
	return RAMBLE_FRONTEND_URL
}

// ShouldUseRemoteBackend returns whether to use remote backend
func ShouldUseRemoteBackend() bool {
	if !IsProduction() {
		if envValue := os.Getenv("USE_REMOTE_AI_BACKEND"); envValue != "" {
			return envValue == "true"
		}
	}
	return USE_REMOTE_AI_BACKEND
}// test comment
