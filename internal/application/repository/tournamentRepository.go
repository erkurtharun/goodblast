package repository

import (
	"context"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"goodblast/internal/domain/entity"
	"goodblast/pkg/log"
	"time"
)

type ITournamentRepository interface {
	CreateTournament(ctx context.Context, t *entity.Tournament) error
	FindByID(ctx context.Context, id int64) (*entity.Tournament, error)
	UpdateTournament(ctx context.Context, t *entity.Tournament) error
	GetActiveTournament(ctx context.Context) (*entity.Tournament, error)
}

type TournamentRepository struct {
	db *bun.DB
}

func NewTournamentRepository(db *bun.DB) ITournamentRepository {
	return &TournamentRepository{
		db: db,
	}
}

func (r *TournamentRepository) CreateTournament(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.NewInsert().Model(t).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create tournament")
	}
	return nil
}

func (r *TournamentRepository) FindByID(ctx context.Context, id int64) (*entity.Tournament, error) {
	var tournament entity.Tournament
	err := r.db.NewSelect().
		Model(&tournament).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		log.GetLogger().Warnf("Tournament not found: %d", id)
		return nil, errors.Wrap(err, "failed to fetch tournament by ID")
	}
	return &tournament, nil
}

func (r *TournamentRepository) UpdateTournament(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.NewUpdate().
		Model(t).
		Column("status").
		Where("id = ?", t.ID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update tournament")
	}
	return nil
}

func (r *TournamentRepository) GetActiveTournament(ctx context.Context) (*entity.Tournament, error) {
	var tournament entity.Tournament
	now := time.Now().UTC()

	err := r.db.NewSelect().
		Model(&tournament).
		Where("status = ?", entity.TournamentStatusActive).
		Where("start_date <= ?", now).
		Where("end_date >= ?", now).
		Scan(ctx)
	if err != nil {
		log.GetLogger().Warn("No active tournament found")
		return nil, errors.Wrap(err, "failed to fetch active tournament")
	}
	return &tournament, nil
}
