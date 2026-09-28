package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/khofesh/img-upload-view/internal/app/bootstrap"
	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/db"
	readconfig "github.com/khofesh/img-upload-view/pkg/read-config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const defaultStaleAge = time.Hour

func Cli() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "cleanup":
		runCleanup(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cli cleanup [-config-path path] [-older-than duration]")
}

func runCleanup(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	cfgPath := fs.String("config-path", "config.dev.yaml", "path to config")
	olderThan := fs.Duration("older-than", defaultStaleAge, "age after which pending uploads are removed")
	fs.Parse(args)

	var cfg config.Config
	if err := readconfig.ReadConfigFromFile(*cfgPath, &cfg); err != nil {
		panic(err)
	}
	cfg.ApplyDefaults()

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(os.Stdout)).With().Timestamp().Logger()

	postgresDB, err := db.OpenDB(cfg.Db)
	if err != nil {
		log.Err(err).Msg("unable to open database")
		os.Exit(1)
	}
	defer postgresDB.Close()

	models := data.NewModels(postgresDB, &log.Logger)
	store := bootstrap.NewStorage(&cfg, &log.Logger)

	stale, err := models.Image.DeleteStalePending(*olderThan)
	if err != nil {
		log.Err(err).Msg("unable to delete stale pending images")
		os.Exit(1)
	}

	ctx := context.Background()
	for _, image := range stale {
		if err := store.Delete(ctx, image.ObjectKey); err != nil {
			log.Warn().Err(err).Str("object_key", image.ObjectKey).Msg("unable to delete stale object")
			continue
		}
		log.Info().Int64("image_id", image.ID).Str("object_key", image.ObjectKey).Msg("stale pending image removed")
	}

	log.Info().Int("count", len(stale)).Msg("cleanup complete")
}
