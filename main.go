package main

import (
	"log"
	"os"

	"EchoBridge/db"
	"EchoBridge/internal/auth"
	"EchoBridge/internal/handlers"
	"EchoBridge/internal/scheduler"
	"EchoBridge/internal/temporal"
	"EchoBridge/internal/worker"

	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Initialize OAuth configurations (must be after .env is loaded)
	auth.InitAuth()

	// Initialize Worker Pool
	syncWorker := worker.NewWorkerPool(100)
	syncWorker.Start(5) // Start 5 workers

	// FIX: Assign the initialized pool to the handlers package variable
	handlers.WorkerPool = syncWorker
	auth.WorkerPool = syncWorker

	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Temporal Client
	if err := temporal.InitClient(); err != nil {
		log.Printf("⚠️ Failed to connect to Temporal (running without workflows): %v", err)
	} else {
		defer temporal.Close()
		// Start Temporal Worker in background
		go func() {
			if err := temporal.StartWorker(); err != nil {
				log.Printf("Temporal worker stopped: %v", err)
			}
		}()
	}

	//  Start Scheduler
	scheduler.StartCategorizationScheduler(syncWorker)

	r := gin.Default()
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL, "http://localhost:5173"}, // Allow both prod and local
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	handlers.RegisterAuthRoutes(r)
	handlers.RegisterPlaylistRoutes(r)

	// Serve Frontend Static Files
	r.Static("/_app", "./web/_app")
	r.StaticFile("/favicon.png", "./web/favicon.png")

	// SPA Fallback
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
