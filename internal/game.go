package chessarchive

import (
	"fmt"
	"io"
)

type Game struct {
	ID     string `firestore:"id"`
	Source string `firestore:"id"`
	Speed  string `firestore:"id"`
	//PlayedAt time.Time
	Winner   string `firestore:"id"`
	Duration int    `firestore:"id"`
	PGN      io.Reader
	Opening  Opening
	Players  struct {
		White Player
		Black Player
	}
}

type Player struct {
	Name     string
	Rating   uint16
	Analysis Analysis
}

type Analysis struct {
	inaccuracy string
	mistake    uint16
	blunder    uint16
	acpl       uint16
}

type Opening struct {
	Name    string
	ECOCode string
}

func (g Game) Name() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}

func (g Game) WinnerName() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}
