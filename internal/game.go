package chessarchive

import (
	"fmt"
	"io"
	"time"
)

type Game struct {
	ID       string
	Source   string
	Speed    string
	PlayedAt time.Time
	Winner   string
	Duration int
	PGN      io.Reader
	Players  struct {
		White Player
		Black Player
	}
}

type Player struct {
	Name   string
	Rating uint16
}

type Opening struct {
	Name    string
	ECOCode string
}

func (g Game) Name() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}
