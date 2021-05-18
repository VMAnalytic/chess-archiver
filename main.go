package main

import (
	"chess-archive/config"
	chessArchive "chess-archive/internal"
	"chess-archive/pkg/google/drive"
	"chess-archive/pkg/google/logging"
	"context"
	"time"

	"github.com/VMAnalytic/lichess-api-client/lichess"

	"cloud.google.com/go/firestore"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger := logging.NewLogger()
	cfg, err := config.NewConfig()

	if err != nil {
		logger.Fatalln(err)
	}

	ctx := context.Background()

	dataStoreClient, err := firestore.NewClient(ctx, cfg.Google.ProjectID)

	if err != nil {
		logger.Fatalln(err)
	}

	lichessClient := lichess.NewClient(cfg.Lichess.APIKey, nil)
	err = lichessClient.SetLimits(1*time.Second, uint(cfg.Lichess.LimitPerSec))

	if err != nil {
		logger.Fatalln(err)
	}

	gameStorage := chessArchive.NewDataStoreGameStorage(logger, dataStoreClient)

	transformer := chessArchive.NewGameTransformer()
	gdClient, _ := drive.NewHTTPtClient(ctx)
	gcloudProcessor := chessArchive.NewDriveStoreProcessor(cfg.Google.ArchiveFolderID, gdClient, transformer, logger)

	dataProcessor := chessArchive.NewDataStoreProcessor(logger, transformer, dataStoreClient)

	arch := chessArchive.NewArchiver(
		logger,
		cfg,
		transformer,
		lichessClient,
		gameStorage,
		[]chessArchive.Processor{gcloudProcessor, dataProcessor},
	)

	err = arch.Run(ctx)

	if err != nil {
		logger.Fatalln(err)
	}

	logger.Infoln("success")
}
