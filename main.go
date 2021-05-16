package main

import (
	"chess-archive/config"
	chessArchive "chess-archive/internal"
	"chess-archive/pkg/google/drive"
	"chess-archive/pkg/google/logging"
	"context"

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

	transformer := chessArchive.NewGameTransformer()
	gdClient, _ := drive.NewHTTPtClient(ctx)
	gcloudProcessor := chessArchive.NewDriveStoreProcessor(cfg.Google.ArchiveFolderID, gdClient, transformer, logger)

	dataStoreClient, err := firestore.NewClient(ctx, cfg.Google.ProjectID)

	if err != nil {
		logger.Fatalln(err)
	}

	dataProcessor := chessArchive.NewDataStoreProcessor(logger, transformer, dataStoreClient)

	arch := chessArchive.NewArchiver(logger, cfg, transformer, []chessArchive.Processor{gcloudProcessor, dataProcessor})

	err = arch.Run(ctx)

	if err != nil {
		logger.Fatalln(err)
	}

	logger.Infoln("success")
}
