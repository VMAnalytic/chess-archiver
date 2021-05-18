package chessarchive

import (
	"chess-archive/pkg/google/drive"
	"context"

	"cloud.google.com/go/firestore"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Processor interface {
	Process(ctx context.Context, game *Game) error
}

type GDriveStoreProcessor struct {
	folderID    string
	gdClient    drive.GDriveClient
	transformer *LichessTransformer
	logger      logrus.FieldLogger
}

func NewDriveStoreProcessor(
	folderID string,
	gdClient drive.GDriveClient,
	transformer *LichessTransformer,
	logger logrus.FieldLogger,
) *GDriveStoreProcessor {
	return &GDriveStoreProcessor{
		folderID:    folderID,
		gdClient:    gdClient,
		transformer: transformer,
		logger:      logger,
	}
}

func (d *GDriveStoreProcessor) Process(ctx context.Context, g *Game) error {
	d.logger.Debugf("GDriveStoreProcessor process game ID: %s", g.ID)

	file, err := d.transformer.TransformToFile(g)

	if err != nil {
		return errors.WithStack(err)
	}

	_, err = d.gdClient.Create(ctx, d.folderID, file)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type DataStoreProcessor struct {
	logger          logrus.FieldLogger
	transformer     *LichessTransformer
	datastoreClient *firestore.Client
}

func NewDataStoreProcessor(
	logger logrus.FieldLogger,
	transformer *LichessTransformer,
	datastoreClient *firestore.Client,
) *DataStoreProcessor {
	return &DataStoreProcessor{
		logger:          logger,
		transformer:     transformer,
		datastoreClient: datastoreClient,
	}
}

func (d *DataStoreProcessor) Process(ctx context.Context, g *Game) error {
	d.logger.Debugf("DataStoreProcessor process game ID: %s", g.ID)

	_, err := d.datastoreClient.Collection("games").Doc(g.ID).Set(ctx, g)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
