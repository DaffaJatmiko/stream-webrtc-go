package streaming

import (
	"log"
	"sync"

	"github.com/DaffaJatmiko/stream_camera/internal/repository"
	"github.com/DaffaJatmiko/stream_camera/pkg/config"
)

type Manager struct {
	streamRepo repository.StreamRepository
	cfg        *config.Config
	mu         sync.RWMutex
}

func NewManager(streamRepo repository.StreamRepository) *Manager {
	return &Manager{
		streamRepo: streamRepo,
		cfg:        config.GetInstance(),
	}
}

func (m *Manager) Start() {
	streams, err := m.streamRepo.GetNonDemandStreams()
	if err != nil {
		log.Printf("Error loading streams: %v", err)
		return
	}

	for _, stream := range streams {
		// Tambahkan stream ke config
		m.cfg.AddStream(stream.UUID, config.StreamConfig{
			URL:      stream.URL,
			Status:   true,
			OnDemand: stream.OnDemand,
			Debug:    stream.Debug,
			Viewers:  make(map[string]config.ViewerConfig),
		})

		// Start RTSP worker
		go StartRTSPWorker(stream.UUID, stream.URL, stream.OnDemand, stream.Debug)
	}
}
