package usecase

import (
	"errors"
	"github.com/DaffaJatmiko/stream_camera/internal/domain/models"
	"log"
	"time"

	"github.com/DaffaJatmiko/stream_camera/internal/repository"
	"github.com/DaffaJatmiko/stream_camera/pkg/config"
	"github.com/deepch/vdk/av"
	webrtc "github.com/deepch/vdk/format/webrtcv3"
)

// WebRTCUseCase defines the interface for WebRTC operations
type WebRTCUseCase interface {
	HandleWebRTC(url string, sdp64 string) (*WebRTCResponse, error)
}

type webrtcUseCase struct {
	cfg        *config.Config
	streamRepo repository.StreamRepository
}

// WebRTCResponse represents the response structure for WebRTC operations
type WebRTCResponse struct {
	Tracks []string `json:"tracks"`
	Sdp64  string   `json:"sdp64"`
}

// NewWebRTCUseCase creates a new instance of WebRTCUseCase
func NewWebRTCUseCase(streamRepo repository.StreamRepository) WebRTCUseCase {
	return &webrtcUseCase{
		cfg:        config.GetInstance(),
		streamRepo: streamRepo,
	}
}

// HandleWebRTC processes a WebRTC connection request
func (u *webrtcUseCase) HandleWebRTC(url string, sdp64 string) (*WebRTCResponse, error) {
	// Get or create stream
	stream, err := u.getOrCreateStream(url)
	if err != nil {
		return nil, err
	}

	// Get stream codecs
	codecs := u.cfg.GetStreamCodecs(stream.UUID)
	if codecs == nil {
		return nil, errors.New("stream codec not found")
	}

	// Setup WebRTC muxer
	muxerWebRTC := u.createWebRTCMuxer()

	// Create answer for WebRTC connection
	answer, err := muxerWebRTC.WriteHeader(codecs, sdp64)
	if err != nil {
		log.Printf("[HandleWebRTC] WriteHeader error: %v", err)
		return nil, err
	}

	// Prepare response
	response := &WebRTCResponse{
		Sdp64: answer,
	}

	// Process codecs and build tracks
	response.Tracks = u.buildTracksFromCodecs(codecs)

	// Start stream handling in background
	go u.handleStreamConnection(stream.UUID, muxerWebRTC, codecs)

	return response, nil
}

// getOrCreateStream retrieves existing stream or creates a new one
func (u *webrtcUseCase) getOrCreateStream(url string) (*models.Stream, error) {
	stream, err := u.streamRepo.GetByURL(url)
	if err != nil {
		log.Printf("[getOrCreateStream] Creating new stream for URL: %s", url)
		stream = &models.Stream{
			URL:      url,
			OnDemand: true,
		}
		err = u.streamRepo.Create(stream)
		if err != nil {
			log.Printf("[getOrCreateStream] Error creating stream: %v", err)
			return nil, err
		}
	}
	return stream, nil
}

// createWebRTCMuxer creates a new WebRTC muxer with configured options
func (u *webrtcUseCase) createWebRTCMuxer() *webrtc.Muxer {
	return webrtc.NewMuxer(webrtc.Options{
		ICEServers:    u.cfg.GetICEServers(),
		ICEUsername:   u.cfg.GetICEUsername(),
		ICECredential: u.cfg.GetICECredential(),
		PortMin:       u.cfg.GetWebRTCPortMin(),
		PortMax:       u.cfg.GetWebRTCPortMax(),
	})
}

// buildTracksFromCodecs processes codecs and returns track types
func (u *webrtcUseCase) buildTracksFromCodecs(codecs []av.CodecData) []string {
	var tracks []string
	for _, codec := range codecs {
		if !u.isCodecSupported(codec) {
			log.Printf("[buildTracksFromCodecs] Codec not supported: %v", codec.Type())
			continue
		}

		if codec.Type().IsVideo() {
			tracks = append(tracks, "video")
		} else {
			tracks = append(tracks, "audio")
		}
	}
	return tracks
}

// isCodecSupported checks if the codec is supported for WebRTC
func (u *webrtcUseCase) isCodecSupported(codec av.CodecData) bool {
	return codec.Type() == av.H264 ||
		codec.Type() == av.PCM_ALAW ||
		codec.Type() == av.PCM_MULAW ||
		codec.Type() == av.OPUS
}

// handleStreamConnection manages the WebRTC stream connection
func (u *webrtcUseCase) handleStreamConnection(streamID string, muxerWebRTC *webrtc.Muxer, codecs []av.CodecData) {
	isAudioOnly := len(codecs) == 1 && codecs[0].Type().IsAudio()
	viewerID, packetChannel := u.cfg.AddViewer(streamID)

	defer func() {
		u.cfg.RemoveViewer(streamID, viewerID)
		muxerWebRTC.Close()
	}()

	var videoStart bool
	noVideo := time.NewTimer(10 * time.Second)
	defer noVideo.Stop()

	for {
		select {
		case <-noVideo.C:
			log.Printf("[handleStreamConnection] No video timeout for stream: %s", streamID)
			return
		case packet := <-packetChannel:
			if u.shouldStartVideo(packet, isAudioOnly, &videoStart) {
				noVideo.Reset(10 * time.Second)
			}

			if !videoStart && !isAudioOnly {
				continue
			}

			if err := muxerWebRTC.WritePacket(packet); err != nil {
				log.Printf("[handleStreamConnection] WritePacket error: %v", err)
				return
			}
		}
	}
}

// shouldStartVideo determines if video streaming should start
func (u *webrtcUseCase) shouldStartVideo(packet av.Packet, isAudioOnly bool, videoStart *bool) bool {
	if packet.IsKeyFrame || isAudioOnly {
		*videoStart = true
		return true
	}
	return false
}
