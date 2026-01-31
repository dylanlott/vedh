package server

import "time"

type GameStatus string

const (
	GameStatusInProgress GameStatus = "IN_PROGRESS"
	GameStatusFinished   GameStatus = "FINISHED"
)

type GameResult string

const (
	GameResultWin  GameResult = "WIN"
	GameResultDraw GameResult = "DRAW"
)

type PendingWinClaim struct {
	ClaimedBy string
	Condition *string
	Remaining []string
}

type GameLogEvent struct {
	ID        string
	GameID    string
	EventTime time.Time
	Type      string
	Actor     *string
	Payload   *string
}
