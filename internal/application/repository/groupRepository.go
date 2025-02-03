package repository

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"goodblast/internal/domain/entity"
	"goodblast/pkg/log"
)

type IGroupRepository interface {
	FindLastGroupForTournamentForUpdate(ctx context.Context, tx bun.Tx, tournamentID int64) (*entity.Group, error)
	CreateGroupTx(ctx context.Context, tx bun.Tx, g *entity.Group) error
	UpdateGroupTx(ctx context.Context, tx bun.Tx, g *entity.Group) error
}

type GroupRepository struct {
	db *bun.DB
}

func NewGroupRepository(db *bun.DB) IGroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) FindLastGroupForTournamentForUpdate(ctx context.Context, tx bun.Tx, tournamentID int64) (*entity.Group, error) {
	var g entity.Group
	err := tx.NewSelect().
		Model(&g).
		Where("tournament_id = ?", tournamentID).
		OrderExpr("group_number DESC").
		Limit(1).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.GetLogger().Warnf("No group found for tournament ID: %d", tournamentID)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to find last group for tournament")
	}
	return &g, nil
}

func (r *GroupRepository) CreateGroupTx(ctx context.Context, tx bun.Tx, g *entity.Group) error {
	_, err := tx.NewInsert().
		Model(g).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create group")
	}
	return nil
}

func (r *GroupRepository) UpdateGroupTx(ctx context.Context, tx bun.Tx, g *entity.Group) error {
	_, err := tx.NewUpdate().
		Model(g).
		Set("current_size = ?", g.CurrentSize).
		Where("id = ?", g.ID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update group")
	}
	return nil
}
