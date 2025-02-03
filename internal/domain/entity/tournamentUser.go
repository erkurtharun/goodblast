package entity

import "time"

type TournamentUser struct {
	ID           int64     `bun:"id,pk,autoincrement"`
	TournamentID int64     `bun:"tournament_id,notnull"`
	UserID       int64     `bun:"user_id,notnull"`
	GroupID      int64     `bun:"group_id,notnull"`
	Score        int       `bun:"score,default:0"`
	CreatedAt    time.Time `bun:"created_at,default:current_timestamp"`
}
