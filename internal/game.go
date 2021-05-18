package chessarchive

import (
	"fmt"
)

type Source int

const (
	_ = iota
	lichessorg
)

type Game struct {
	ID       string  `firestore:"id"`
	Source   Source  `firestore:"source"`
	Speed    string  `firestore:"speed"`
	Duration int16   `firestore:"duration"`
	Status   string  `firestore:"status"`
	PlayedAt int64   `firestore:"played_at"`
	Winner   string  `firestore:"winner"`
	PGN      string  `firestore:"pgn"`
	Opening  Opening `firestore:"opening"`
	Players  struct {
		White Player `firestore:"white"`
		Black Player `firestore:"black"`
	} `firestore:"players"`
}

type Player struct {
	Name     string   `firestore:"name"`
	Rating   uint16   `firestore:"rating"`
	Analysis Analysis `firestore:"analysis"`
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

func (g Game) Name() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}

func (g Game) WinnerName() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}

func (g Game) Moves() string {
	return fmt.Sprintf("%s - %s.pgn", g.Players.White.Name, g.Players.Black.Name)
}
