package chessarchive

import (
	"chess-archive/config"
	"fmt"
	"time"
)

type Source int
type UserResult string

const (
	_ = iota
	lichessorg
)

const (
	win  = UserResult("win")
	lose = UserResult("lose")
	draw = UserResult("draw")
)

var format = "2006-01-02 15:04:05"

type Game struct {
	ID         string     `firestore:"id"`
	Source     Source     `firestore:"source"`
	Speed      string     `firestore:"speed"`
	Duration   uint16     `firestore:"duration"`
	Status     string     `firestore:"status"`
	UserResult UserResult `firestore:"result"`
	PlayedAt   int64      `firestore:"played_at"` //milliseconds
	Winner     string     `firestore:"winner"`
	PGN        string     `firestore:"pgn"`
	Opening    *Opening   `firestore:"opening,omitempty"`
	Players    struct {
		White Player `firestore:"white"`
		Black Player `firestore:"black"`
	} `firestore:"players"`
}

type Player struct {
	ID       string    `firestore:"id"`
	Name     string    `firestore:"name"`
	Rating   uint16    `firestore:"rating"`
	Analysis *Analysis `firestore:"analysis,omitempty"`
}

type Analysis struct {
	Inaccuracy uint16 `firestore:"inaccuracy"`
	Mistake    uint16 `firestore:"mistake"`
	Blunder    uint16 `firestore:"blunder"`
	ACPL       uint16 `firestore:"acpl"`
}

type Opening struct {
	Name    string `firestore:"name"`
	ECOCode string `firestore:"eco_code"`
}

func (g *Game) Name() string {
	return fmt.Sprintf(
		"%s | %s | %s - %s.pgn",
		g.PlayedAtTime().Format(format),
		g.Result(),
		g.Players.White.Name,
		g.Players.Black.Name,
	)
}

func (g *Game) PlayedAtTime() time.Time {
	loc, _ := time.LoadLocation(config.TimeZone)

	return time.Unix(0, g.PlayedAt*int64(time.Millisecond)).In(loc)
}

func (g *Game) Result() string {
	switch g.Winner {
	case "black":
		return "0-1"
	case "white":
		return "1-0"
	default:
		return "1/2 - 1/2"
	}
}
