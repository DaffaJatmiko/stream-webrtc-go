package handlers

import (
	"encoding/json"
	"github.com/deepch/vdk/av"
	"log"
	"net/http"
	"time"

	"github.com/DaffaJatmiko/stream_camera/internal/usecase"
	"github.com/DaffaJatmiko/stream_camera/pkg/config"
	webrtc "github.com/deepch/vdk/format/webrtcv3"
	"github.com/gin-gonic/gin"
)

type CodecInfo struct {
	Type string `json:"Type"`
}

type WebRTCHandler struct {
	webrtcUseCase usecase.WebRTCUseCase
	cfg           *config.Config
}

func NewWebRTCHandler(webrtcUseCase usecase.WebRTCUseCase) *WebRTCHandler {
	return &WebRTCHandler{
		webrtcUseCase: webrtcUseCase,
		cfg:           config.GetInstance(),
	}
}

// GetStreamCodec handles codec information retrieval for a stream
func (h *WebRTCHandler) GetStreamCodec(c *gin.Context) {
	streamID := c.Param("uuid")
	log.Printf("[GetStreamCodec] Called with Stream ID: %s", streamID)

	if !h.cfg.StreamExists(streamID) {
		log.Printf("[GetStreamCodec] Stream %s not found", streamID)
		c.Writer.Write([]byte(""))
		return
	}

	h.cfg.StartStreamIfNotRunning(streamID)
	codecs := h.cfg.GetStreamCodecs(streamID)
	if codecs == nil {
		log.Printf("[GetStreamCodec] No codecs for stream %s", streamID)
		c.Writer.Write([]byte(""))
		return
	}

	var tmpCodec []CodecInfo
	for _, codec := range codecs {
		if codec.Type() != av.H264 &&
			codec.Type() != av.PCM_ALAW &&
			codec.Type() != av.PCM_MULAW &&
			codec.Type() != av.OPUS {
			log.Printf("[GetStreamCodec] Codec Not Supported WebRTC ignore this track: %v", codec.Type())
			continue
		}
		if codec.Type().IsVideo() {
			tmpCodec = append(tmpCodec, CodecInfo{Type: "video"})
			log.Printf("[GetStreamCodec] Added video codec")
		} else {
			tmpCodec = append(tmpCodec, CodecInfo{Type: "audio"})
			log.Printf("[GetStreamCodec] Added audio codec")
		}
	}

	b, err := json.Marshal(tmpCodec)
	if err != nil {
		log.Printf("[GetStreamCodec] Error marshaling codec info: %v", err)
		c.Writer.Write([]byte(""))
		return
	}

	log.Printf("[GetStreamCodec] Sending codec response: %s", string(b))
	c.Writer.Write(b)
}

// HandleWebRTCWithUUID processes WebRTC connections for a specific stream
func (h *WebRTCHandler) HandleWebRTCWithUUID(c *gin.Context) {
	streamID := c.Param("uuid")
	sdp64 := c.PostForm("data")
	log.Printf("[HandleWebRTCWithUUID] Called with Stream ID: %s", streamID)

	if !h.cfg.StreamExists(streamID) {
		log.Printf("[HandleWebRTCWithUUID] Stream %s not found", streamID)
		return
	}

	h.cfg.StartStreamIfNotRunning(streamID)
	codecs := h.cfg.GetStreamCodecs(streamID)
	if codecs == nil {
		log.Printf("[HandleWebRTCWithUUID] Stream %s codec not found", streamID)
		return
	}

	var AudioOnly bool
	if len(codecs) == 1 && codecs[0].Type().IsAudio() {
		AudioOnly = true
		log.Printf("[HandleWebRTCWithUUID] Stream %s is audio only", streamID)
	}

	log.Printf("[HandleWebRTCWithUUID] Setting up ICE Servers: %v", h.cfg.GetICEServers())
	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{
		ICEServers:    h.cfg.GetICEServers(),
		ICEUsername:   h.cfg.GetICEUsername(),
		ICECredential: h.cfg.GetICECredential(),
		PortMin:       h.cfg.GetWebRTCPortMin(),
		PortMax:       h.cfg.GetWebRTCPortMax(),
	})

	answer, err := muxerWebRTC.WriteHeader(codecs, sdp64)
	if err != nil {
		log.Printf("[HandleWebRTCWithUUID] WriteHeader error: %v", err)
		return
	}

	log.Printf("[HandleWebRTCWithUUID] Successfully created WebRTC answer")
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		log.Printf("[HandleWebRTCWithUUID] Write error: %v", err)
		return
	}

	go h.handleStreamConnection(streamID, muxerWebRTC, AudioOnly)
}

// HandleWebRTC processes WebRTC connections with URL
func (h *WebRTCHandler) HandleWebRTC(c *gin.Context) {
	url := c.PostForm("url")
	sdp64 := c.PostForm("sdp64")

	response, err := h.webrtcUseCase.HandleWebRTC(url, sdp64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleStreamConnection manages the WebRTC stream connection
func (h *WebRTCHandler) handleStreamConnection(streamID string, muxerWebRTC *webrtc.Muxer, AudioOnly bool) {
	viewerID, packetChannel := h.cfg.AddViewer(streamID)
	defer h.cfg.RemoveViewer(streamID, viewerID)
	defer muxerWebRTC.Close()

	var videoStart bool
	noVideo := time.NewTimer(10 * time.Second)
	defer noVideo.Stop()

	for {
		select {
		case <-noVideo.C:
			log.Printf("[handleStreamConnection] No video timeout for stream %s", streamID)
			return
		case packet := <-packetChannel:
			if packet.IsKeyFrame || AudioOnly {
				noVideo.Reset(10 * time.Second)
				videoStart = true
			}
			if !videoStart && !AudioOnly {
				continue
			}
			if err := muxerWebRTC.WritePacket(packet); err != nil {
				log.Printf("[handleStreamConnection] WritePacket error for stream %s: %v", streamID, err)
				return
			}
		}
	}
}
