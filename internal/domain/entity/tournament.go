package entity

import (
	"time"
)

type TournamentStatus string

const (
	TournamentStatusPlanned TournamentStatus = "planned"
	TournamentStatusActive  TournamentStatus = "active"
	TournamentStatusClosed  TournamentStatus = "closed"
)

type Tournament struct {
	ID        int64            `bun:"id,pk,autoincrement"`
	StartDate time.Time        `bun:"start_date,notnull"`
	EndDate   time.Time        `bun:"end_date,notnull"`
	Status    TournamentStatus `bun:"status,type:varchar(16),default:'planned'"`
}

func (t *Tournament) IsActive() bool {
	now := time.Now().UTC()
	return t.Status == TournamentStatusActive &&
		now.After(t.StartDate) &&
		now.Before(t.EndDate)
}

func (t *Tournament) HasEnded() bool {
	now := time.Now().UTC()
	return t.Status == TournamentStatusClosed || now.After(t.EndDate)
}

func (t *Tournament) Activate() {
	t.Status = TournamentStatusActive
}

func (t *Tournament) Close() {
	t.Status = TournamentStatusClosed
}
