package chessarchive

import (
	"chess-archive/config"
	"context"

	"github.com/VMAnalytic/lichess-api-client/lichess"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Archiver struct {
	logger        logrus.FieldLogger
	cfg           *config.Config
	transformer   *LichessTransformer
	chessProvider *lichess.Client
	gameStorage   GameStorage
	processors    []Processor
}

func NewArchiver(
	logger logrus.FieldLogger,
	cfg *config.Config,
	transformer *LichessTransformer,
	chessProvider *lichess.Client,
	gameStorage GameStorage,
	processors []Processor,
) *Archiver {
	return &Archiver{
		logger:        logger,
		cfg:           cfg,
		transformer:   transformer,
		chessProvider: chessProvider,
		gameStorage:   gameStorage,
		processors:    processors,
	}
}

func (a Archiver) Run(ctx context.Context) error {
	a.logger.Infoln("process started...")

	since := int64(0)
	latest, err := a.gameStorage.Last(ctx)

	if err != nil {
		return errors.WithStack(err)
	}

	if latest != nil {
		since = latest.PlayedAt
	}

	games, _, err := a.chessProvider.Games.List(ctx, a.cfg.Lichess.Username, lichess.ListOptions{Since: since})
	if err != nil {
		return errors.WithStack(err)
	}

	group, gctx := errgroup.WithContext(ctx)

	for i, g := range games {
		if i == 10 {
			break
		}

		game, err := a.transformer.Transform(g)
		if err != nil {
			return errors.WithStack(err)
		}

		for _, p := range a.processors {
			proc := p

			group.Go(func() error {
				err = proc.Process(gctx, game)
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

	a.logger.Infoln("process finished...")

	return nil
}
