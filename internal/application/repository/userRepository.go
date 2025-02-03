package repository

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"goodblast/internal/domain/entity"
	"goodblast/pkg/log"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByID(ctx context.Context, userId int64) (*entity.User, error)
	UpdateProgress(ctx context.Context, user *entity.User) error
	FindUserForUpdateTx(ctx context.Context, tx bun.Tx, userID int64) (*entity.User, error)
	UpdateUserTx(ctx context.Context, tx bun.Tx, u *entity.User) error
}

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(database *bun.DB) IUserRepository {
	return &UserRepository{db: database}
}

func (usrRepo *UserRepository) CreateUser(ctx context.Context, user *entity.User) (int64, error) {
	_, err := usrRepo.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create user")
	}
	return user.ID, nil
}

func (usrRepo *UserRepository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := usrRepo.db.NewSelect().Model(&user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.GetLogger().Warnf("User not found: %s", username)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to fetch user by username")
	}
	return &user, nil
}

func (usrRepo *UserRepository) GetUserByID(ctx context.Context, userId int64) (*entity.User, error) {
	var user entity.User
	err := usrRepo.db.NewSelect().Model(&user).Where("id = ?", userId).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.GetLogger().Warnf("User not found: %d", userId)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to fetch user by ID")
	}
	return &user, nil
}

func (usrRepo *UserRepository) UpdateProgress(ctx context.Context, user *entity.User) error {
	_, err := usrRepo.db.NewUpdate().
		Model(user).
		Set("level = ?", user.Level).
		Set("coins = ?", user.Coins).
		Where("id = ?", user.ID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update user progress")
	}
	return nil
}

func (usrRepo *UserRepository) FindUserForUpdateTx(ctx context.Context, tx bun.Tx, userID int64) (*entity.User, error) {
	var u entity.User
	if err := tx.NewSelect().
		Model(&u).
		Where("id = ?", userID).
		For("UPDATE").
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.GetLogger().Warnf("User not found for update: %d", userID)
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to find user for update")
	}
	return &u, nil
}

func (usrRepo *UserRepository) UpdateUserTx(ctx context.Context, tx bun.Tx, u *entity.User) error {
	_, err := tx.NewUpdate().
		Model(u).
		Set("coins = ?", u.Coins).
		Set("level = ?", u.Level).
		Where("id = ?", u.ID).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update user transactionally")
	}
	return nil
}
