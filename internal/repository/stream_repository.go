package repository

import (
	"github.com/DaffaJatmiko/stream_camera/internal/domain/models"
	"gorm.io/gorm"
)

type StreamRepository interface {
	GetAll() ([]models.Stream, error)
	GetByUUID(uuid string) (*models.Stream, error)
	GetByURL(url string) (*models.Stream, error)
	Create(stream *models.Stream) error
	Update(stream *models.Stream) error
	Delete(uuid string) error
	GetNonDemandStreams() ([]models.Stream, error)
}

type streamRepository struct {
	db *gorm.DB
}

func NewStreamRepository(db *gorm.DB) StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) GetAll() ([]models.Stream, error) {
	var streams []models.Stream
	err := r.db.Find(&streams).Error
	return streams, err
}

func (r *streamRepository) GetByUUID(uuid string) (*models.Stream, error) {
	var stream models.Stream
	err := r.db.Where("uuid = ?", uuid).First(&stream).Error
	return &stream, err
}

func (r *streamRepository) Create(stream *models.Stream) error {
	return r.db.Create(stream).Error
}

func (r *streamRepository) Update(stream *models.Stream) error {
	return r.db.Save(stream).Error
}

func (r *streamRepository) Delete(uuid string) error {
	return r.db.Where("uuid = ?", uuid).Delete(&models.Stream{}).Error
}

func (r *streamRepository) GetNonDemandStreams() ([]models.Stream, error) {
	var streams []models.Stream
	err := r.db.Where("on_demand = ?", false).Find(&streams).Error
	return streams, err
}

func (r *streamRepository) GetByURL(url string) (*models.Stream, error) {
	var stream models.Stream
	result := r.db.Where("url = ?", url).First(&stream)
	if result.Error != nil {
		return nil, result.Error
	}
	return &stream, nil
}
