package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"go.avagenc.com/spotify/db"
	"go.avagenc.com/spotify/handlers"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/joho/godotenv"
	"go.naturallyfunny.dev/api/identity"
)

func main() {
	// Load .env file if it exists (ignored in Lambda where env vars are set natively)
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	}

	// Initialize Database
	dbURL := os.Getenv("SPOTIFY_DATABASE_URL")
	if dbURL == "" {
		log.Println("Warning: SPOTIFY_DATABASE_URL is not set")
	} else {
		if err := db.Init(dbURL); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
	}

	// Register routes
	authMiddleware := identity.WithUserIDFromHeader("x-user-id")
	http.Handle("/play", authMiddleware(http.HandlerFunc(handlers.PlayHandler)))
	http.Handle("/get-music", authMiddleware(http.HandlerFunc(handlers.GetMusicHandler)))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Check if running inside AWS Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("AWS Lambda environment detected. Starting Lambda handler...")
		lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
		return
	}

	// Local mode
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Spotify API Agent starting on http://localhost%s", addr)
	log.Printf("Endpoints:")
	log.Printf("  POST /play")
	log.Printf("  GET  /get-music")
	log.Printf("  GET  /health")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
