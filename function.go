package gfunctions

import (
	"chess-archive/config"
	chessArchive "chess-archive/internal"
	"chess-archive/pkg/google/drive"
	"chess-archive/pkg/google/logging"
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/VMAnalytic/lichess-api-client/lichess"

	"github.com/sirupsen/logrus"
)

var (
	logger    logrus.FieldLogger
	cfg       *config.Config
	initError error
)

func init() {
	logger = logging.NewLogger()
	cfg, initError = config.NewConfig()

	if initError != nil {
		logger.Fatal(initError)
	}

	config.TimeZone = cfg.TimeZone
}

// PubSubMessage is the payload of a Google Pub/Sub event
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// TrackEvent consumes a Pub/Sub message.
func TrackEvent(ctx context.Context, m PubSubMessage) error {
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

	transformer := chessArchive.NewGameTransformer(cfg.Lichess.UserID)
	gdClient, err := drive.NewHTTPtClient(ctx)

	if err != nil {
		logger.Fatalln(err)
	}

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

	return nil
}
