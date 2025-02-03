package entity

type Group struct {
	ID           int64 `bun:"id,pk,autoincrement"`
	TournamentID int64 `bun:"tournament_id,notnull"`
	GroupNumber  int   `bun:"group_number,notnull"`
	CurrentSize  int   `bun:"current_size,notnull,default:0"`
}
