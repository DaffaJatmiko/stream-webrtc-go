package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DaffaJatmiko/stream_camera/internal/delivery/http"
	"github.com/DaffaJatmiko/stream_camera/internal/repository"
	"github.com/DaffaJatmiko/stream_camera/internal/usecase"
	"github.com/DaffaJatmiko/stream_camera/pkg/config"
	"github.com/DaffaJatmiko/stream_camera/pkg/database"
	"github.com/DaffaJatmiko/stream_camera/pkg/streaming"
)

func main() {
	// Initialize config
	cfg := config.GetInstance()

	// Initialize database
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize repository
	streamRepo := repository.NewStreamRepository(db.DB)

	// Initialize streaming manager
	streamManager := streaming.NewManager(streamRepo)
	go streamManager.Start()

	// Initialize usecases
	streamUsecase := usecase.NewStreamUseCase(streamRepo)
	webrtcUsecase := usecase.NewWebRTCUseCase(streamRepo)

	// Initialize HTTP server
	router := http.NewRouter(streamUsecase, webrtcUsecase)
	go router.Run(cfg.Server.HTTPPort)
	//go func() {
	//	if err := router.Run(cfg.Server.HTTPPort); err != nil {
	//		log.Fatal("Failed to start HTTP server:", err)
	//	}
	//}()

	// Wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println(sig)
		done <- true
	}()

	log.Println("Server Start Awaiting Signal")
	<-done
	log.Println("Exiting")
}
