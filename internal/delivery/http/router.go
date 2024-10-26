package http

import (
	"github.com/DaffaJatmiko/stream_camera/internal/delivery/http/handlers"
	"github.com/DaffaJatmiko/stream_camera/internal/delivery/http/middleware"
	"github.com/DaffaJatmiko/stream_camera/internal/usecase"
	"github.com/gin-gonic/gin"
	"log"
)

type Router struct {
	engine        *gin.Engine
	streamHandler *handlers.StreamHandler
	webrtcHandler *handlers.WebRTCHandler
}

func NewRouter(streamUseCase usecase.StreamUseCase, webrtcUseCase usecase.WebRTCUseCase) *Router {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Add middlewares
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Logger())

	r := &Router{
		engine:        router,
		streamHandler: handlers.NewStreamHandler(streamUseCase),
		webrtcHandler: handlers.NewWebRTCHandler(webrtcUseCase),
	}

	// Setup routes immediately
	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	// Existing routes
	r.engine.GET("/streams", r.streamHandler.GetStreamList)
	r.engine.POST("/stream", r.webrtcHandler.HandleWebRTC)
	r.engine.POST("/stream/receiver/:uuid", r.webrtcHandler.HandleWebRTCWithUUID)
	r.engine.GET("/stream/codec/:uuid", r.webrtcHandler.GetStreamCodec)

	api := r.engine.Group("/api")
	{
		api.GET("/streams", r.streamHandler.GetStreamList)
		api.GET("/streams/:uuid", r.streamHandler.GetStream)
		api.POST("/streams", r.streamHandler.CreateStream)
		api.PUT("/streams/:uuid", r.streamHandler.UpdateStream)
		api.DELETE("/streams/:uuid", r.streamHandler.DeleteStream)
	}
}

func (r *Router) Run(addr string) error {
	log.Printf("Starting server on %s", addr)
	return r.engine.Run(addr)
}
