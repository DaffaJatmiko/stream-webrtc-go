package handlers

import (
	"net/http"

	"github.com/DaffaJatmiko/stream_camera/internal/domain/models"
	"github.com/DaffaJatmiko/stream_camera/internal/usecase"
	"github.com/gin-gonic/gin"
)

type StreamHandler struct {
	streamUseCase usecase.StreamUseCase
}

func NewStreamHandler(streamUseCase usecase.StreamUseCase) *StreamHandler {
	return &StreamHandler{
		streamUseCase: streamUseCase,
	}
}

func (h *StreamHandler) GetStreamList(c *gin.Context) {
	streams, err := h.streamUseCase.GetAllStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, streams)
}

func (h *StreamHandler) GetStream(c *gin.Context) {
	uuid := c.Param("uuid")
	stream, err := h.streamUseCase.GetStream(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}
	c.JSON(http.StatusOK, stream)
}

func (h *StreamHandler) CreateStream(c *gin.Context) {
	var stream models.Stream
	if err := c.ShouldBindJSON(&stream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.streamUseCase.CreateStream(&stream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stream)
}

func (h *StreamHandler) UpdateStream(c *gin.Context) {
	uuid := c.Param("uuid")
	var stream models.Stream
	if err := c.ShouldBindJSON(&stream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.streamUseCase.UpdateStream(uuid, &stream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stream)
}

func (h *StreamHandler) DeleteStream(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := h.streamUseCase.DeleteStream(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stream deleted successfully"})
}
