// pkg/config/config.go
package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/DaffaJatmiko/stream_camera/pkg/utils"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtspv2"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	mutex     sync.RWMutex
	Server    ServerConfig            `json:"server"`
	Streams   map[string]StreamConfig `json:"streams"`
	LastError error
}

type ServerConfig struct {
	HTTPPort      string   `json:"http_port"`
	ICEServers    []string `json:"ice_servers"`
	ICEUsername   string   `json:"ice_username"`
	ICECredential string   `json:"ice_credential"`
	WebRTCPortMin uint16   `json:"webrtc_port_min"`
	WebRTCPortMax uint16   `json:"webrtc_port_max"`
}

type StreamConfig struct {
	URL          string `json:"url"`
	Status       bool   `json:"status"`
	OnDemand     bool   `json:"on_demand"`
	DisableAudio bool   `json:"disable_audio"`
	Debug        bool   `json:"debug"`
	RunLock      bool   `json:"-"`
	Codecs       []av.CodecData
	Viewers      map[string]ViewerConfig // Renamed from Cl for clarity
}

type ViewerConfig struct {
	PacketChannel chan av.Packet // Renamed from C for clarity
}

// GetInstance returns singleton instance of Config
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.loadConfiguration()
	})
	return instance
}

// loadConfiguration loads config from file or sets default values
func (c *Config) loadConfiguration() error {
	data, err := ioutil.ReadFile("config/config.json")
	if err == nil {
		err = json.Unmarshal(data, c)
		if err != nil {
			return err
		}

		// Initialize viewers map for each stream
		for id, stream := range c.Streams {
			if stream.Viewers == nil {
				stream.Viewers = make(map[string]ViewerConfig)
			}
			c.Streams[id] = stream
		}
	} else {
		c.loadDefaultConfiguration()
	}
	return nil
}

// loadDefaultConfiguration sets default values for config
func (c *Config) loadDefaultConfiguration() {
	addr := flag.String("listen", ":8083", "HTTP host:port")
	udpMin := flag.Int("udp_min", 0, "WebRTC UDP port min")
	udpMax := flag.Int("udp_max", 0, "WebRTC UDP port max")
	iceServer := flag.String("ice_server", "", "ICE Server")
	flag.Parse()

	c.Server.HTTPPort = *addr
	c.Server.WebRTCPortMin = uint16(*udpMin)
	c.Server.WebRTCPortMax = uint16(*udpMax)
	if len(*iceServer) > 0 {
		c.Server.ICEServers = []string{*iceServer}
	}

	c.Streams = make(map[string]StreamConfig)
}

// Server configuration getters
func (c *Config) GetICEServers() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Server.ICEServers
}

func (c *Config) GetICEUsername() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Server.ICEUsername
}

func (c *Config) GetICECredential() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Server.ICECredential
}

func (c *Config) GetWebRTCPortMin() uint16 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Server.WebRTCPortMin
}

func (c *Config) GetWebRTCPortMax() uint16 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Server.WebRTCPortMax
}

// Stream management methods
func (c *Config) StreamExists(streamID string) bool { // Renamed from Ext
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, exists := c.Streams[streamID]
	return exists
}

func (c *Config) AddStream(streamID string, cfg StreamConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Streams[streamID] = cfg
}

func (c *Config) GetStreamCodecs(streamID string) []av.CodecData { // Renamed from CoGe
	maxRetries := 100
	retryInterval := 50 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		c.mutex.RLock()
		stream, exists := c.Streams[streamID]
		c.mutex.RUnlock()

		if !exists {
			return nil
		}
		if stream.Codecs != nil {
			return stream.Codecs
		}
		time.Sleep(retryInterval)
	}
	return nil
}

// Viewer management methods
func (c *Config) AddViewer(streamID string) (string, chan av.Packet) { // Renamed from ClAd
	c.mutex.Lock()
	defer c.mutex.Unlock()

	viewerID := utils.GenerateUUID()
	packetChannel := make(chan av.Packet, 100)

	stream, exists := c.Streams[streamID]
	if !exists {
		stream = StreamConfig{
			Viewers: make(map[string]ViewerConfig),
		}
	}
	if stream.Viewers == nil {
		stream.Viewers = make(map[string]ViewerConfig)
	}

	stream.Viewers[viewerID] = ViewerConfig{PacketChannel: packetChannel}
	c.Streams[streamID] = stream

	return viewerID, packetChannel
}

func (c *Config) RemoveViewer(streamID, viewerID string) { // Renamed from ClDe
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists {
		delete(stream.Viewers, viewerID)
	}
}

// Stream operation methods
func (c *Config) StartStreamIfNotRunning(streamID string) { // Renamed from RunIFNotRun
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists && stream.OnDemand {
		go c.startStreamWorker(streamID, stream.URL, stream.OnDemand, stream.Debug)
	}
}

func (c *Config) BroadcastPacket(streamID string, packet av.Packet) { // Renamed from Cast
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists {
		for _, viewer := range stream.Viewers {
			select {
			case viewer.PacketChannel <- packet:
			default:
			}
		}
	}
}

func (c *Config) UpdateStreamCodecs(streamID string, codecs []av.CodecData) { // Renamed from CoAd
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists {
		stream.Codecs = codecs
		c.Streams[streamID] = stream
	}
}

// Error handling methods
func (c *Config) SetLastError(err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.LastError = err
}

func (c *Config) GetLastError() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.LastError
}

// Stream status methods
func (c *Config) HasViewers(streamID string) bool { // Renamed from HasViewer
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists {
		return len(stream.Viewers) > 0
	}
	return false
}

func (c *Config) UnlockStream(streamID string) { // Renamed from RunUnlock
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if stream, exists := c.Streams[streamID]; exists {
		if stream.OnDemand && stream.RunLock {
			stream.RunLock = false
			c.Streams[streamID] = stream
		}
	}
}

// RTSP worker methods
func (c *Config) startStreamWorker(streamID string, url string, onDemand bool, debug bool) {
	for {
		log.Printf("Attempting to connect to stream: %s", streamID)
		err := c.handleRTSPStream(streamID, url, debug)
		if err != nil {
			log.Printf("Stream error: %v", err)
			c.SetLastError(err)
		}
		if onDemand {
			return
		}
		time.Sleep(time.Second)
	}
}

func (c *Config) handleRTSPStream(streamID string, url string, debug bool) error {
	client, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              url,
		DisableAudio:     true,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            debug,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	if client.CodecData != nil {
		c.UpdateStreamCodecs(streamID, client.CodecData)
	}

	for {
		packet := <-client.OutgoingPacketQueue
		c.BroadcastPacket(streamID, *packet)
	}
}

// Tambahkan method di config.go
func (c *Config) SyncStreamConfig(streamID string, url string, onDemand bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Streams[streamID] = StreamConfig{
		URL:      url,
		Status:   true,
		OnDemand: onDemand,
		Viewers:  make(map[string]ViewerConfig),
	}
}
