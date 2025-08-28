package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/stripe/stripe-go/v79"

	aihandlers "pocketbase/internal/ai"
	bannerhandlers "pocketbase/internal/banners"
	otphandlers "pocketbase/internal/otp"
	stripehandlers "pocketbase/internal/stripe"
	"pocketbase/webauthn"
	_ "pocketbase/migrations"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	app := pocketbase.New()

	// Check if we're running with `go run`
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	// Register the migrate command
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun, // Auto-migrate during development
	})

	// Configure Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Register WebAuthn
	webauthn.Register(app)

	// Configure SMTP settings on app initialization
	if err := configureEmailSettings(app); err != nil {
		log.Printf("Failed to configure SMTP: %v", err)
	}

	// Configure app settings for large file uploads
	app.OnBootstrap().BindFunc(func(be *core.BootstrapEvent) error {
		log.Println("Configuring PocketBase with large file upload support")
		
		// The file upload size limits are controlled by the middleware and server configuration
		// We'll configure the server to handle large requests in the OnServe hook
		
		return be.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Configure request body size limit for large audio files
		se.Server.MaxHeaderBytes = 1 << 20  // 1MB for headers
		se.Server.ReadTimeout = 300 * time.Second // 5 minutes for large files
		se.Server.WriteTimeout = 300 * time.Second
		
		// IMPORTANT: Configure body size limits BEFORE default middleware
		// PocketBase's default body limit is 32MB, we need to bypass this for audio uploads
		
		log.Printf("Server configured: ReadTimeout=%v, WriteTimeout=%v", 
			se.Server.ReadTimeout, se.Server.WriteTimeout)

		// Log Whisper configuration for audio processing
		logWhisperConfiguration()

		// Seed development data if in development mode
		if err := aihandlers.SeedDevelopmentData(app); err != nil {
			log.Printf("Warning: Failed to seed development data: %v", err)
		}


		// Stripe routes
		se.Router.POST("/create-checkout-session", func(e *core.RequestEvent) error {
			return stripehandlers.CreateCheckoutSession(e, app)
		})

		se.Router.POST("/create-portal-link", func(e *core.RequestEvent) error {
			return stripehandlers.CreatePortalLink(e, app)
		})

		se.Router.POST("/stripe", func(e *core.RequestEvent) error {
			return stripehandlers.HandleWebhook(e, app)
		})

		// OTP routes
		se.Router.POST("/send-otp", func(e *core.RequestEvent) error {
			return otphandlers.SendOTPHandler(e, app)
		})

		se.Router.POST("/verify-otp", func(e *core.RequestEvent) error {
			return otphandlers.VerifyOTPHandler(e, app)
		})

		// AI routes
		se.Router.POST("/api/ai/process-text", func(e *core.RequestEvent) error {
			return aihandlers.ProcessTextHandler(e, app)
		})

		// Audio processing route with streaming support and increased body limit
		// Override the default 32MB body limit to allow up to 2GB audio files
		se.Router.POST("/api/ai/process-audio", func(e *core.RequestEvent) error {
			log.Printf("🎵 Processing audio upload with 2GB body limit")
			return aihandlers.ProcessAudioHandler(e, app)
		}).Bind(apis.BodyLimit(2 << 30)) // 2GB body limit for audio uploads

		se.Router.POST("/api/generate-api-key", func(e *core.RequestEvent) error {
			return aihandlers.GenerateAPIKeyHandler(e, app)
		})

		// Usage tracking routes for Wails app (requires API key)
		se.Router.GET("/api/usage/summary", func(e *core.RequestEvent) error {
			return aihandlers.UsageSummaryHandler(e, app)
		})

		se.Router.GET("/api/usage/files", func(e *core.RequestEvent) error {
			return aihandlers.UsageFilesHandler(e, app)
		})

		se.Router.GET("/api/usage/stats", func(e *core.RequestEvent) error {
			return aihandlers.UsageStatsHandler(e, app)
		})

		// Banner routes
		se.Router.GET("/api/banners", func(e *core.RequestEvent) error {
			return bannerhandlers.GetPublicBannersHandler(e, app)
		})

		se.Router.GET("/api/banners/authenticated", func(e *core.RequestEvent) error {
			return bannerhandlers.GetAuthenticatedBannersHandler(e, app)
		})



		// Serve static files from the provided public dir (if exists)
		// This must be registered last as it's a catch-all route
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// configureEmailSettings sets up SMTP configuration for email verification
func configureEmailSettings(app *pocketbase.PocketBase) error {
	// Only configure if SMTP_HOST is set
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		log.Println("SMTP_HOST not set, email verification disabled")
		return nil
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		smtpPort = 587 // default
	}

	smtpTLS := os.Getenv("SMTP_TLS") == "true"
	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		emailFrom = "noreply@localhost"
	}
	emailFromName := os.Getenv("EMAIL_FROM_NAME")
	if emailFromName == "" {
		emailFromName = "Pulse"
	}

	// Configure SMTP settings
	app.Settings().SMTP.Enabled = true
	app.Settings().SMTP.Host = smtpHost
	app.Settings().SMTP.Port = smtpPort
	app.Settings().SMTP.Username = os.Getenv("SMTP_USERNAME")
	app.Settings().SMTP.Password = os.Getenv("SMTP_PASSWORD")
	app.Settings().SMTP.TLS = smtpTLS
	app.Settings().SMTP.AuthMethod = "PLAIN"

	// Configure email templates
	app.Settings().Meta.SenderName = emailFromName
	app.Settings().Meta.SenderAddress = emailFrom
	
	log.Printf("SMTP configured: %s:%d (TLS: %v)", smtpHost, smtpPort, smtpTLS)
	return nil
}

// logWhisperConfiguration logs the Whisper API configuration for audio processing
func logWhisperConfiguration() {
	var maxSize int64
	var source string
	
	if maxSizeStr := os.Getenv("WHISPER_MAX_FILE_SIZE"); maxSizeStr != "" {
		if parsedSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = parsedSize
			source = "environment variable"
		} else {
			log.Printf("Warning: Invalid WHISPER_MAX_FILE_SIZE value '%s', using default", maxSizeStr)
			maxSize = 25 * 1024 * 1024 // 25MB default
			source = "default (invalid env var)"
		}
	} else {
		maxSize = 25 * 1024 * 1024 // 25MB default
		source = "default"
	}
	
	sizeMB := float64(maxSize) / (1024 * 1024)
	log.Printf("[WHISPER_CONFIG] Max file size: %d bytes (%.1f MB) - source: %s", maxSize, sizeMB, source)
	
	// Also log the PocketBase body limit for comparison
	bodyLimitGB := float64(2<<30) / (1024 * 1024 * 1024)
	log.Printf("[WHISPER_CONFIG] PocketBase body limit: %.0f GB for audio uploads", bodyLimitGB)
}