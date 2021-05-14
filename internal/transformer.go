package chessarchive

import (
	"chess-archive/pkg/google/drive"
	"strings"
	"time"

	"github.com/VMAnalytic/lichess-api-client/lichess"
	"github.com/pkg/errors"
)

type LichessTransformer struct {
}

func NewGameTransformer() *LichessTransformer {
	return &LichessTransformer{}
}

func (t LichessTransformer) Transform(v interface{}) (*Game, error) {
	switch game := v.(type) {
	case *lichess.Game:
		return t.transformLichess(game)

	default:
		return nil, errors.New("unknown type")
	}
}

func (t LichessTransformer) transformLichess(lg *lichess.Game) (*Game, error) {
	if lg == nil {
		return nil, errors.New("nil game")
	}

	var g Game

	g.ID = lg.ID
	g.Speed = lg.Speed
	g.PlayedAt = time.Unix(lg.CreatedAt, 0)
	g.PGN = strings.NewReader(lg.Pgn)
	g.Players.White.Name = lg.Players.White.User.Name
	g.Players.White.Rating = uint16(lg.Players.White.Rating)
	g.Players.Black.Name = lg.Players.Black.User.Name
	g.Players.Black.Rating = uint16(lg.Players.Black.Rating)

	return &g, nil
}

func (t LichessTransformer) TransformToFile(game *Game) (*drive.File, error) {
	var f drive.File

	f.Name = game.Name()
	f.Media = game.PGN
	f.Description = "Test"

	return &f, nil
}
