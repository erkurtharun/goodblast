package repository

import (
	"context"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"goodblast/internal/domain/entity"
	"goodblast/pkg/log"
)

type ITournamentRewardRepository interface {
	CreateRewards(ctx context.Context, rewards []entity.TournamentReward) error
	GetUnclaimedRewardsByUser(ctx context.Context, userID int64) ([]entity.TournamentReward, error)
	ClaimReward(ctx context.Context, rewardID int64) error
}

type TournamentRewardRepository struct {
	db *bun.DB
}

func NewTournamentRewardRepository(db *bun.DB) ITournamentRewardRepository {
	return &TournamentRewardRepository{db: db}
}

func (r *TournamentRewardRepository) CreateRewards(ctx context.Context, rewards []entity.TournamentReward) error {
	_, err := r.db.NewInsert().
		Model(&rewards).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create tournament rewards")
	}
	return nil
}

func (r *TournamentRewardRepository) GetUnclaimedRewardsByUser(ctx context.Context, userID int64) ([]entity.TournamentReward, error) {
	var list []entity.TournamentReward
	err := r.db.NewSelect().
		Model(&list).
		Where("user_id = ?", userID).
		Where("claimed = false").
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		log.GetLogger().Warnf("No unclaimed rewards found for user ID: %d", userID)
		return nil, errors.Wrap(err, "failed to fetch unclaimed rewards")
	}
	return list, nil
}

func (r *TournamentRewardRepository) ClaimReward(ctx context.Context, rewardID int64) error {
	_, err := r.db.NewUpdate().
		Model((*entity.TournamentReward)(nil)).
		Set("claimed = true").
		Where("id = ?", rewardID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to claim reward")
	}
	return nil
}
