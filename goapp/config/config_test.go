package config

import (
	"os"
	"testing"
)

func TestDevelopmentConfig(t *testing.T) {
	// In development (no production tag), we should get dev defaults
	if IsProduction() {
		t.Skip("Skipping development test in production build")
	}

	// Test default values
	if USE_REMOTE_AI_BACKEND != false {
		t.Errorf("Expected USE_REMOTE_AI_BACKEND to be false in dev, got %v", USE_REMOTE_AI_BACKEND)
	}

	if REMOTE_AI_BACKEND_URL != "http://localhost:8090" {
		t.Errorf("Expected REMOTE_AI_BACKEND_URL to be localhost in dev, got %s", REMOTE_AI_BACKEND_URL)
	}

	// Test env override
	os.Setenv("USE_REMOTE_AI_BACKEND", "true")
	defer os.Unsetenv("USE_REMOTE_AI_BACKEND")
	
	if !ShouldUseRemoteBackend() {
		t.Error("Expected ShouldUseRemoteBackend to return true with env override")
	}
}

func TestProductionConfig(t *testing.T) {
	// In production build, values should be hardcoded
	if !IsProduction() {
		t.Skip("Skipping production test in development build")
	}

	if !USE_REMOTE_AI_BACKEND {
		t.Error("Expected USE_REMOTE_AI_BACKEND to be true in production")
	}

	if REMOTE_AI_BACKEND_URL != "https://api.ramble.goosebyteshq.com" {
		t.Errorf("Expected production backend URL, got %s", REMOTE_AI_BACKEND_URL)
	}

	if RAMBLE_FRONTEND_URL != "https://ramble.goosebyteshq.com" {
		t.Errorf("Expected production frontend URL, got %s", RAMBLE_FRONTEND_URL)
	}
}