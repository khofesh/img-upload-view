package data

import (
	"database/sql"

	"github.com/rs/zerolog"
)

type Models struct {
	Image IImageModel
}

func NewModels(db *sql.DB, logger *zerolog.Logger) Models {
	return Models{
		Image: ImageModel{postgresDB: db, logger: logger},
	}
}
