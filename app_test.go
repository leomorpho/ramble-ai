package main

import (
	"testing"
)

// Note: The filterValidHighlightSuggestions functionality is now tested
// in the highlights package tests (goapp/highlights/highlights_test.go)
// since it's a private method of AIService.

func TestAppStructCreation(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Error("NewApp() should return a non-nil App instance")
	}
	if app.client == nil {
		t.Error("App should have a non-nil client")
	}
}
