package main

import (
	"log"
	"os"

	"EchoBridge/db"
	"EchoBridge/internal/handlers"
	"EchoBridge/internal/scheduler"
	"EchoBridge/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Worker Pool
	syncWorker := worker.NewWorkerPool(100)
	syncWorker.Start(5) // Start 5 workers

	// FIX: Assign the initialized pool to the handlers package variable
	handlers.WorkerPool = syncWorker

	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Start Scheduler
	scheduler.StartCategorizationScheduler(syncWorker)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Your Frontend URL
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	handlers.RegisterAuthRoutes(r)
	handlers.RegisterPlaylistRoutes(r)
	// Share Routes are included in RegisterPlaylistRoutes in the updated handler file,
	// but if you want to keep them separate in structure:
	// handlers.RegisterShareRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
