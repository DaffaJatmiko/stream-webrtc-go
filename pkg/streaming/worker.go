// pkg/streaming/worker.go

package streaming

import (
	"errors"
	"log"
	"time"

	"github.com/DaffaJatmiko/stream_camera/pkg/config"
	"github.com/deepch/vdk/format/rtspv2"
)

var (
	ErrorStreamExitNoVideoOnStream = errors.New("stream exit no video on stream")
	ErrorStreamExitRtspDisconnect  = errors.New("stream exit rtsp disconnect")
	ErrorStreamExitNoViewer        = errors.New("stream exit on demand no viewer")
)

func StartRTSPWorker(uuid string, url string, onDemand bool, debug bool) {
	for {
		log.Println("Stream Try Connect", uuid)
		err := RTSPWorker(uuid, url, onDemand, debug)
		if err != nil {
			log.Println(err)
			config.GetInstance().SetLastError(err)
		}
		if onDemand {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func RTSPWorker(uuid string, url string, onDemand bool, debug bool) error {
	keyTest := time.NewTimer(20 * time.Second)
	clientTest := time.NewTimer(20 * time.Second)
	defer keyTest.Stop()
	defer clientTest.Stop()

	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              url,
		DisableAudio:     true,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            debug,
	})
	if err != nil {
		return err
	}
	defer RTSPClient.Close()

	cfg := config.GetInstance()
	if RTSPClient.CodecData != nil {
		cfg.UpdateStreamCodecs(uuid, RTSPClient.CodecData)
	}

	AudioOnly := len(RTSPClient.CodecData) == 1 && RTSPClient.CodecData[0].Type().IsAudio()

	for {
		select {
		case <-clientTest.C:
			if onDemand {
				if !cfg.HasViewers(uuid) {
					return ErrorStreamExitNoViewer
				}
				clientTest.Reset(20 * time.Second)
			}
		case <-keyTest.C:
			return ErrorStreamExitNoVideoOnStream
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				cfg.UpdateStreamCodecs(uuid, RTSPClient.CodecData)
			case rtspv2.SignalStreamRTPStop:
				return ErrorStreamExitRtspDisconnect
			}
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if AudioOnly || packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
			}
			cfg.BroadcastPacket(uuid, *packetAV)
		}
	}
}
