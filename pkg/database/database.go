package database

import (
	"github.com/DaffaJatmiko/stream_camera/internal/domain/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Definisikan struct Database
type Database struct {
	DB *gorm.DB
}

// Membuat koneksi ke PostgreSQL dan mengembalikan instance Database
func NewPostgresDB() (*Database, error) {
	dsn := "host=localhost user=postgres password=jatming dbname=keran_kitera port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migrate the models
	err = db.AutoMigrate(&models.Stream{})
	if err != nil {
		return nil, err
	}

	// Mengembalikan instance Database
	return &Database{DB: db}, nil
}

// Fungsi untuk menutup koneksi database
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
