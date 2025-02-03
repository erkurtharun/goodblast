package entity

import "time"

type TournamentReward struct {
	ID           int64     `bun:"id,pk,autoincrement"`
	TournamentID int64     `bun:"tournament_id,notnull"`
	UserID       int64     `bun:"user_id,notnull"`
	Rank         int       `bun:"rank,notnull"`
	RewardCoins  int       `bun:"reward_coins,notnull"`
	Claimed      bool      `bun:"claimed,default:false"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
