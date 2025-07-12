// Package chatbot provides AI-powered conversational interfaces for video editing assistance.
//
// This package implements a chatbot service that can:
// - Handle conversational AI interactions
// - Execute function calls for highlight reordering
// - Manage chat history and sessions
// - Interface with OpenRouter API for LLM capabilities
//
// The package is organized into several files:
// - types.go: Data structures and type definitions
// - service.go: Main service logic and core functionality
// - functions.go: Function calling and execution logic
// - api.go: OpenRouter API communication
//
// Usage:
//   service := chatbot.NewChatbotService(client, ctx)
//   response, err := service.SendMessage(request, getAPIKeyFunc)
//
package chatbot