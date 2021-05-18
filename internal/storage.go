package chessarchive

import (
	"chess-archive/pkg/google/drive"
	"context"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type GameStorage interface {
	Last(ctx context.Context) (*Game, error)
}

type GDriveGameStorage struct {
	folderID     string
	transformer  *LichessTransformer
	gDriveClient drive.GDriveClient
}

func (gds *GDriveGameStorage) Last(ctx context.Context) (*Game, error) {
	f, err := gds.gDriveClient.Latest(ctx, gds.folderID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	game, err := gds.transformer.Transform(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return game, nil
}

type DataStoreGameStorage struct {
	logger          logrus.FieldLogger
	datastoreClient *firestore.Client
}

func NewDataStoreGameStorage(
	logger logrus.FieldLogger,
	datastoreClient *firestore.Client,
) *DataStoreGameStorage {
	return &DataStoreGameStorage{
		logger:          logger,
		datastoreClient: datastoreClient,
	}
}

func (ds *DataStoreGameStorage) Last(ctx context.Context) (*Game, error) {
	var g Game

	query := ds.datastoreClient.Collection("games").OrderBy("played_at", firestore.Desc).Limit(1)
	iter := query.Documents(ctx)
	doc, err := iter.Next()

	if err != nil {
		if err == iterator.Done {
			return nil, nil
		}

		return nil, errors.WithStack(err)
	}

	err = doc.DataTo(&g)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &g, nil
}
