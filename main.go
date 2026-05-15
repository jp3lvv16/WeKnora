// WeKnora - A collaborative knowledge management platform
// Fork of Tencent/WeKnora
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// Version information, injected at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	// Load environment variables from .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Print startup banner
	fmt.Printf("WeKnora v%s (build: %s, commit: %s)\n", Version, BuildTime, Commit)

	// Initialize application configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	registerRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals
	go func() {
		log.Printf("Server listening on %s:%s", cfg.Host, cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Allow up to 30 seconds for active connections to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// AppConfig holds the application-level configuration
type AppConfig struct {
	Host string
	Port string
	Env  string
}

// loadConfig reads configuration from environment variables
func loadConfig() (*AppConfig, error) {
	host := getEnv("APP_HOST", "0.0.0.0")
	port := getEnv("APP_PORT", "8080")
	env := getEnv("APP_ENV", "development")

	return &AppConfig{
		Host: host,
		Port: port,
		Env:  env,
	}, nil
}

// registerRoutes sets up the HTTP route handlers
func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", notFoundHandler)
}

// healthHandler returns a simple health check response
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","version":%q}`, Version)
}

// notFoundHandler returns a 404 response for unmatched routes
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, `{"error":"not found"}`)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
