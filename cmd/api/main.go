package main

import (
	"flag"
	"os"

	"github.com/khofesh/img-upload-view/internal/app/api"
	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/db"
	"github.com/khofesh/img-upload-view/pkg/errors"
	readconfig "github.com/khofesh/img-upload-view/pkg/read-config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var cfg config.Config
	var cfgPath string
	flag.StringVar(&cfgPath, "config-path", "/etc/secrets/config.yaml", "path to config")

	flag.Parse()
	err := readconfig.ReadConfigFromFile(cfgPath, &cfg)
	if err != nil {
		panic(err)
	}

	// zerolog
	multiWriters := zerolog.MultiLevelWriter(os.Stdout)
	log.Logger = zerolog.New(multiWriters).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	db, err := db.OpenDB(cfg.Db)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	log.Info().Msg("database connection pool established.")

	app := &config.Application{
		Logger:        &log.Logger,
		Config:        &cfg,
		Models:        data.NewModels(db, &log.Logger),
		ErrorResponse: errors.NewErrorResponse(&log.Logger),
	}

	err = api.Serve(app)
	if err != nil {
		log.Err(err)
		os.Exit(1)
	}
}
