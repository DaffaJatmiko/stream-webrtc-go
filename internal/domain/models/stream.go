package models

import (
	"gorm.io/gorm"
)

type Stream struct {
	gorm.Model
	UUID     string `json:"uuid"`
	URL      string `json:"url"`
	OnDemand bool   `json:"on_demand" gorm:"default:false"`
	Debug    bool   `json:"debug" gorm:"default:false"`
}

type StreamResponse struct {
	UUID     string `json:"uuid"`
	URL      string `json:"url"`
	OnDemand bool   `json:"on_demand"`
	Debug    bool   `json:"debug"`
}
