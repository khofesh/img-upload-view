package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rs/zerolog"
)

type IImageModel interface {
	Insert(image *Image) error
	GetAll(limit, offset int64) ([]*Image, int64, error)
	GetByID(id int64) (*Image, error)
	Delete(id int64) error
	GetByFilename(filename string) (*Image, error)
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

func (m ImageModel) Insert(image *Image) error {
	query := `
		INSERT INTO images (filename, original_filename, url, file_size, content_type, upload_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	args := []any{
		image.Filename,
		image.OriginalFilename,
		image.URL,
		image.FileSize,
		image.ContentType,
		image.UploadTimestamp,
	}

	ctx := context.Background()
	row := m.postgresDB.QueryRowContext(ctx, query, args...)

	var createdAt, updatedAt time.Time
	err := row.Scan(&image.ID, &createdAt, &updatedAt)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to insert image")
		return err
	}

	m.logger.Info().
		Int64("image_id", image.ID).
		Str("filename", image.Filename).
		Msg("Image inserted successfully")

	return nil
}

func (m ImageModel) GetAll(limit, offset int64) ([]*Image, int64, error) {
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM images`

	ctx := context.Background()
	err := m.postgresDB.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to get total image count")
		return nil, 0, err
	}

	query := `
		SELECT id, filename, original_filename, url, file_size, content_type, upload_timestamp
		FROM images 
		ORDER BY upload_timestamp DESC 
		LIMIT $1 OFFSET $2`

	args := []any{limit, offset}

	rows, err := m.postgresDB.QueryContext(ctx, query, args...)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to query images")
		return nil, 0, err
	}
	defer rows.Close()

	images := []*Image{}

	for rows.Next() {
		var image Image
		err := rows.Scan(
			&image.ID,
			&image.Filename,
			&image.OriginalFilename,
			&image.URL,
			&image.FileSize,
			&image.ContentType,
			&image.UploadTimestamp,
		)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to scan image row")
			return nil, 0, err
		}
		images = append(images, &image)
	}

	if err = rows.Err(); err != nil {
		m.logger.Error().Err(err).Msg("Error occurred during row iteration")
		return nil, 0, err
	}

	m.logger.Info().
		Int64("total_count", totalCount).
		Int("returned_count", len(images)).
		Int64("limit", limit).
		Int64("offset", offset).
		Msg("Images retrieved successfully")

	return images, totalCount, nil
}

func (m ImageModel) GetByID(id int64) (*Image, error) {
	if id < 1 {
		return nil, errors.New("invalid image ID")
	}

	query := `
		SELECT id, filename, original_filename, url, file_size, content_type, upload_timestamp
		FROM images 
		WHERE id = $1`

	var image Image
	ctx := context.Background()

	err := m.postgresDB.QueryRowContext(ctx, query, id).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.URL,
		&image.FileSize,
		&image.ContentType,
		&image.UploadTimestamp,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.logger.Warn().Int64("image_id", id).Msg("Image not found")
			return nil, errors.New("record not found")
		}
		m.logger.Error().Err(err).Int64("image_id", id).Msg("Failed to get image by ID")
		return nil, err
	}

	m.logger.Info().Int64("image_id", id).Msg("Image retrieved successfully")
	return &image, nil
}

func (m ImageModel) Delete(id int64) error {
	if id < 1 {
		return errors.New("invalid image ID")
	}

	query := `DELETE FROM images WHERE id = $1`

	ctx := context.Background()
	result, err := m.postgresDB.ExecContext(ctx, query, id)
	if err != nil {
		m.logger.Error().Err(err).Int64("image_id", id).Msg("Failed to delete image")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("record not found")
	}

	m.logger.Info().Int64("image_id", id).Msg("Image deleted successfully")
	return nil
}

func (m ImageModel) GetByFilename(filename string) (*Image, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	query := `
		SELECT id, filename, original_filename, url, file_size, content_type, upload_timestamp
		FROM images 
		WHERE filename = $1`

	var image Image
	ctx := context.Background()

	err := m.postgresDB.QueryRowContext(ctx, query, filename).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.URL,
		&image.FileSize,
		&image.ContentType,
		&image.UploadTimestamp,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("record not found")
		}
		m.logger.Error().Err(err).Str("filename", filename).Msg("Failed to get image by filename")
		return nil, err
	}

	return &image, nil
}
