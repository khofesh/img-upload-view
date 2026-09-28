package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rs/zerolog"
)

const (
	StatusPending = "pending"
	StatusReady   = "ready"
)

type IImageModel interface {
	InsertPending(image *Image) error
	MarkReady(id, size int64, contentType string) error
	GetAll(limit, offset int64) ([]*Image, int64, error)
	GetByID(id int64) (*Image, error)
	Delete(id int64) error
	DeleteStalePending(olderThan time.Duration) ([]*Image, error)
}

type Image struct {
	ID               int64     `json:"id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	ObjectKey        string    `json:"-"`
	URL              string    `json:"url"`
	FileSize         int64     `json:"file_size"`
	ContentType      string    `json:"content_type"`
	Status           string    `json:"status"`
	UploadTimestamp  time.Time `json:"upload_timestamp"`
}

type ImageModel struct {
	postgresDB *sql.DB
	logger     *zerolog.Logger
}

func (m ImageModel) InsertPending(image *Image) error {
	query := `
		INSERT INTO images (filename, original_filename, object_key, file_size, content_type, status, upload_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	args := []any{
		image.Filename,
		image.OriginalFilename,
		image.ObjectKey,
		image.FileSize,
		image.ContentType,
		StatusPending,
		image.UploadTimestamp,
	}

	ctx := context.Background()
	row := m.postgresDB.QueryRowContext(ctx, query, args...)

	var createdAt, updatedAt time.Time
	err := row.Scan(&image.ID, &createdAt, &updatedAt)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to insert pending image")
		return err
	}

	image.Status = StatusPending

	m.logger.Info().
		Int64("image_id", image.ID).
		Str("filename", image.Filename).
		Msg("Pending image inserted successfully")

	return nil
}

func (m ImageModel) MarkReady(id, size int64, contentType string) error {
	query := `
		UPDATE images
		SET status = $2, file_size = $3, content_type = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	ctx := context.Background()
	result, err := m.postgresDB.ExecContext(ctx, query, id, StatusReady, size, contentType)
	if err != nil {
		m.logger.Error().Err(err).Int64("image_id", id).Msg("Failed to mark image ready")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("record not found")
	}

	m.logger.Info().Int64("image_id", id).Msg("Image marked ready successfully")
	return nil
}

func (m ImageModel) GetAll(limit, offset int64) ([]*Image, int64, error) {
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM images WHERE status = $1`

	ctx := context.Background()
	err := m.postgresDB.QueryRowContext(ctx, countQuery, StatusReady).Scan(&totalCount)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to get total image count")
		return nil, 0, err
	}

	query := `
		SELECT id, filename, original_filename, object_key, file_size, content_type, status, upload_timestamp
		FROM images 
		WHERE status = $1
		ORDER BY upload_timestamp DESC 
		LIMIT $2 OFFSET $3`

	args := []any{StatusReady, limit, offset}

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
			&image.ObjectKey,
			&image.FileSize,
			&image.ContentType,
			&image.Status,
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
		SELECT id, filename, original_filename, object_key, file_size, content_type, status, upload_timestamp
		FROM images 
		WHERE id = $1`

	var image Image
	ctx := context.Background()

	err := m.postgresDB.QueryRowContext(ctx, query, id).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.ObjectKey,
		&image.FileSize,
		&image.ContentType,
		&image.Status,
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

func (m ImageModel) DeleteStalePending(olderThan time.Duration) ([]*Image, error) {
	query := `
		DELETE FROM images
		WHERE status = $1 AND upload_timestamp < $2
		RETURNING id, filename, original_filename, object_key, file_size, content_type, status, upload_timestamp`

	ctx := context.Background()
	rows, err := m.postgresDB.QueryContext(ctx, query, StatusPending, time.Now().Add(-olderThan))
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to delete stale pending images")
		return nil, err
	}
	defer rows.Close()

	images := []*Image{}
	for rows.Next() {
		var image Image
		err := rows.Scan(
			&image.ID,
			&image.Filename,
			&image.OriginalFilename,
			&image.ObjectKey,
			&image.FileSize,
			&image.ContentType,
			&image.Status,
			&image.UploadTimestamp,
		)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to scan stale pending image row")
			return nil, err
		}
		images = append(images, &image)
	}

	if err = rows.Err(); err != nil {
		m.logger.Error().Err(err).Msg("Error occurred during stale pending row iteration")
		return nil, err
	}

	m.logger.Info().Int("count", len(images)).Msg("Stale pending images deleted successfully")
	return images, nil
}
