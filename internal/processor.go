package chessarchive

import (
	"chess-archive/pkg/google/drive"
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Processor interface {
	Process(ctx context.Context, game *Game) error
}

type DriveStoreProcessor struct {
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
) *DriveStoreProcessor {
	return &DriveStoreProcessor{
		folderID:    folderID,
		gdClient:    gdClient,
		transformer: transformer,
		logger:      logger,
	}
}

func (d DriveStoreProcessor) Process(ctx context.Context, g *Game) error {
	d.logger.Infoln(g.ID)
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
