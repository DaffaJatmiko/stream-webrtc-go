package usecase

import (
	"github.com/DaffaJatmiko/stream_camera/internal/domain/models"
	"github.com/DaffaJatmiko/stream_camera/internal/repository"
	"github.com/DaffaJatmiko/stream_camera/pkg/utils"
)

type StreamUseCase interface {
	GetAllStreams() ([]models.StreamResponse, error)
	GetStream(uuid string) (*models.StreamResponse, error)
	CreateStream(stream *models.Stream) error
	UpdateStream(uuid string, stream *models.Stream) error
	DeleteStream(uuid string) error
}

type streamUseCase struct {
	streamRepo repository.StreamRepository
}

func NewStreamUseCase(streamRepo repository.StreamRepository) StreamUseCase {
	return &streamUseCase{
		streamRepo: streamRepo,
	}
}

func (u *streamUseCase) GetAllStreams() ([]models.StreamResponse, error) {
	streams, err := u.streamRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []models.StreamResponse
	for _, stream := range streams {
		response = append(response, models.StreamResponse{
			UUID:     stream.UUID,
			URL:      stream.URL,
			OnDemand: stream.OnDemand,
			Debug:    stream.Debug,
		})
	}
	return response, nil
}

func (u *streamUseCase) GetStream(uuid string) (*models.StreamResponse, error) {
	stream, err := u.streamRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}

	return &models.StreamResponse{
		UUID:     stream.UUID,
		URL:      stream.URL,
		OnDemand: stream.OnDemand,
		Debug:    stream.Debug,
	}, nil
}

func (u *streamUseCase) CreateStream(stream *models.Stream) error {
	stream.UUID = utils.GenerateUUID()
	return u.streamRepo.Create(stream)
}

func (u *streamUseCase) UpdateStream(uuid string, stream *models.Stream) error {
	existingStream, err := u.streamRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}

	existingStream.URL = stream.URL
	existingStream.OnDemand = stream.OnDemand
	existingStream.Debug = stream.Debug

	return u.streamRepo.Update(existingStream)
}

func (u *streamUseCase) DeleteStream(uuid string) error {
	return u.streamRepo.Delete(uuid)
}
