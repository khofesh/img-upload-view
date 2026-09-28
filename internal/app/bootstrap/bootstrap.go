package bootstrap

import (
	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/storage"
	"github.com/rs/zerolog"
)

func NewStorage(cfg *config.Config, logger *zerolog.Logger) storage.ObjectStore {
	return storage.NewS3Store(storage.Options{
		Endpoint:       cfg.Storage.Endpoint,
		PublicEndpoint: cfg.Storage.PublicEndpoint,
		Region:         cfg.Storage.Region,
		Bucket:         cfg.Storage.Bucket,
		AccessKey:      cfg.Storage.AccessKey,
		SecretKey:      cfg.Storage.SecretKey,
		UsePathStyle:   cfg.Storage.UsePathStyle,
		PresignPutTTL:  cfg.Storage.PresignPutTTL,
		PresignGetTTL:  cfg.Storage.PresignGetTTL,
		CORSOrigins:    cfg.TrustedOrigins,
	}, logger)
}
