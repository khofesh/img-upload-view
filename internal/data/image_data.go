package data

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog"
)

type IImageModel interface {
	Insert(Image) error
	GetAll(limit, offset int) ([]*Image, int, error)
	GetByID(id int64) (*Image, error)
}

type Image struct {
	ID               int64     `json:"id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	URL              string    `json:"url"`
	FileSize         int64     `json:"file_size"`
	ContentType      string    `json:"content_type"`
	UploadTimestamp  time.Time `json:"upload_timestamp"`
}

type ImageModel struct {
	postgresDB *sql.DB
	logger     *zerolog.Logger
}

func (m ImageModel) Insert(Image) error {
	return nil
}
func (m ImageModel) GetAll(limit, offset int) ([]*Image, int, error) {
	return nil, 0, nil
}
func (m ImageModel) GetByID(id int64) (*Image, error) {
	return nil, nil
}
