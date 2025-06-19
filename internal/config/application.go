package config

import (
	errres "github.com/khofesh/img-upload-view/pkg/errors"

	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/rs/zerolog"
)

type Application struct {
	Logger        *zerolog.Logger
	Config        *Config
	Models        data.Models
	ErrorResponse errres.ErrorResponse
}
