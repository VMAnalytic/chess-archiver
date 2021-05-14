package chessarchive

import (
	"chess-archive/config"
	"chess-archive/pkg/google/drive"
	"context"
	"time"

	"github.com/VMAnalytic/lichess-api-client/lichess"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Archiver struct {
	logger      logrus.FieldLogger
	cfg         *config.Config
	Transformer *LichessTransformer
	processors  []Processor
}

func NewArchiver(
	logger logrus.FieldLogger,
	cfg *config.Config,
	transformer *LichessTransformer,
	processors []Processor,
) *Archiver {
	return &Archiver{
		logger:      logger,
		cfg:         cfg,
		Transformer: transformer,
		processors:  processors,
	}
}

func (a Archiver) Run(ctx context.Context) error {
	lichessClient := lichess.NewClient(a.cfg.Lichess.APIKey, nil)
	err := lichessClient.SetLimits(1*time.Second, 20)

	if err != nil {
		return errors.WithStack(err)
	}

	gdClient, err := drive.NewHTTPtClient(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	since := int64(0)
	latest, err := gdClient.Latest(ctx, a.cfg.Google.ArchiveFolderID)

	if err != nil {
		return errors.WithStack(err)
	}

	if latest != nil && latest.CreatedAt() != nil {
		since = latest.CreatedAt().Unix()
	}

	games, _, err := lichessClient.Games.List(ctx, a.cfg.Lichess.Username, lichess.ListOptions{Since: since})
	if err != nil {
		return errors.WithStack(err)
	}

	group, gctx := errgroup.WithContext(ctx)

	for _, g := range games {
		for _, p := range a.processors {
			game, err := a.Transformer.Transform(g)
			if err != nil {
				return errors.WithStack(err)
			}

			group.Go(func() error {
				err = p.Process(gctx, game)
				if err != nil {
					return errors.WithStack(err)
				}

				return nil
			})
		}
	}

	if err = group.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
