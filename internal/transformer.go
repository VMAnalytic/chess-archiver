package chessarchive

import (
	"chess-archive/pkg/google/drive"
	"strings"

	"github.com/VMAnalytic/lichess-api-client/lichess"
	"github.com/fatih/structs"
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
		return nil, errors.New("game should not be nil")
	}

	var g Game

	g.ID = lg.ID
	g.Source = lichessorg
	g.Speed = lg.Speed
	g.PlayedAt = lg.CreatedAt
	g.Winner = lg.Winner
	g.Status = lg.Status
	g.PGN = lg.Pgn

	g.Players.White.Name = lg.Players.White.User.Name
	g.Players.White.Rating = uint16(lg.Players.White.Rating)

	if lg.Players.White.Analysis != nil {
		g.Players.White.Analysis = Analysis{
			Inaccuracy: uint16(lg.Players.White.Analysis.Inaccuracy),
			Mistake:    uint16(lg.Players.White.Analysis.Mistake),
			Blunder:    uint16(lg.Players.White.Analysis.Blunder),
			ACPL:       uint16(lg.Players.White.Analysis.ACPL),
		}
	}

	g.Players.Black.Name = lg.Players.Black.User.Name
	g.Players.Black.Rating = uint16(lg.Players.Black.Rating)

	if lg.Players.Black.Analysis != nil {
		g.Players.Black.Analysis = Analysis{
			Inaccuracy: uint16(lg.Players.Black.Analysis.Inaccuracy),
			Mistake:    uint16(lg.Players.Black.Analysis.Mistake),
			Blunder:    uint16(lg.Players.Black.Analysis.Blunder),
			ACPL:       uint16(lg.Players.Black.Analysis.ACPL),
		}
	}

	g.Opening = Opening{
		Name:    lg.Opening.Name,
		ECOCode: lg.Opening.Eco,
	}

	return &g, nil
}

func (t LichessTransformer) TransformToFile(game *Game) (*drive.File, error) {
	var f drive.File

	f.Name = game.Name()
	f.Media = strings.NewReader(game.PGN)
	f.Description = "Test"

	return &f, nil
}

func (t LichessTransformer) TransformToMap(game *Game) map[string]interface{} {
	data := structs.Map(game)

	return data
}
