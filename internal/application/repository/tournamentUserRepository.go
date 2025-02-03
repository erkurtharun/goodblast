package repository

import (
	"context"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"goodblast/internal/domain/entity"
	"goodblast/pkg/log"
)

type ITournamentUserRepository interface {
	CreateTournamentUserTx(ctx context.Context, tx bun.Tx, tu *entity.TournamentUser) error
	GetTournamentUser(ctx context.Context, tournamentID, userID int64) (*entity.TournamentUser, error)
	UpdateScore(ctx context.Context, tu *entity.TournamentUser) error
	GetTournamentUsersByTournament(ctx context.Context, tournamentID int64) ([]entity.TournamentUser, error)
	GetTournamentUsersByGroup(ctx context.Context, tournamentID int64, groupID int) ([]entity.TournamentUser, error)
}

type TournamentUserRepository struct {
	db *bun.DB
}

func NewTournamentUserRepository(db *bun.DB) ITournamentUserRepository {
	return &TournamentUserRepository{db: db}
}

func (r *TournamentUserRepository) CreateTournamentUserTx(ctx context.Context, tx bun.Tx, tu *entity.TournamentUser) error {
	_, err := tx.NewInsert().
		Model(tu).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create tournament user")
	}
	return nil
}

func (r *TournamentUserRepository) GetTournamentUser(ctx context.Context, tournamentID, userID int64) (*entity.TournamentUser, error) {
	var tu entity.TournamentUser
	err := r.db.NewSelect().
		Model(&tu).
		Where("tournament_id = ?", tournamentID).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		log.GetLogger().Warnf("Tournament user not found for tournament ID: %d, user ID: %d", tournamentID, userID)
		return nil, errors.Wrap(err, "failed to fetch tournament user")
	}
	return &tu, nil
}

func (r *TournamentUserRepository) UpdateScore(ctx context.Context, tu *entity.TournamentUser) error {
	_, err := r.db.NewUpdate().
		Model(tu).
		Set("score = ?", tu.Score).
		Where("id = ?", tu.ID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update tournament user score")
	}
	return nil
}

func (r *TournamentUserRepository) GetTournamentUsersByTournament(ctx context.Context, tournamentID int64) ([]entity.TournamentUser, error) {
	var list []entity.TournamentUser
	err := r.db.NewSelect().
		Model(&list).
		Where("tournament_id = ?", tournamentID).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch tournament users by tournament ID")
	}
	return list, nil
}

func (r *TournamentUserRepository) GetTournamentUsersByGroup(ctx context.Context, tournamentID int64, groupID int) ([]entity.TournamentUser, error) {
	var list []entity.TournamentUser
	err := r.db.NewSelect().
		Model(&list).
		Where("tournament_id = ?", tournamentID).
		Where("group_id = ?", groupID).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch tournament users by group")
	}
	return list, nil
}
