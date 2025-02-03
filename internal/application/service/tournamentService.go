package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/uptrace/bun"
	appconfig "goodblast/config"
	"goodblast/internal/application/repository"
	"goodblast/internal/domain/entity"
	domainErr "goodblast/internal/domain/errors"
	"goodblast/internal/domain/events"
	kafkautil "goodblast/internal/infrastructure/kafka/producer"
	"goodblast/pkg/log"
	"sort"
	"time"
)

type ITournamentService interface {
	CreateDailyTournament(ctx context.Context) (*entity.Tournament, error)
	StartTournament(ctx context.Context, id int64) error
	CloseTournament(ctx context.Context, id int64) error
	GetActiveTournament(ctx context.Context) (*entity.Tournament, error)
	EnterTournamentAsync(ctx context.Context, userID int64) error
	EnterTournament(ctx context.Context, userID int64) error
	UpdateTournamentScore(ctx context.Context, message events.ProgressUpdateMessage) error
	StoreRewards(ctx context.Context, tournamentID int64) error
	ClaimReward(ctx context.Context, userID int64) error
}

type TournamentService struct {
	db                   *bun.DB
	tRepo                repository.ITournamentRepository
	gRepo                repository.IGroupRepository
	tuRepo               repository.ITournamentUserRepository
	uRepo                repository.IUserRepository
	tournamentRewardRepo repository.ITournamentRewardRepository
	dynamicConfigService appconfig.IDynamicConfigService
	producer             *ckafka.Producer
}

func NewTournamentService(
	db *bun.DB,
	tRepo repository.ITournamentRepository,
	gRepo repository.IGroupRepository,
	tuRepo repository.ITournamentUserRepository,
	uRepo repository.IUserRepository,
	tournamentRewardRepo repository.ITournamentRewardRepository,
	dynamicConfigService appconfig.IDynamicConfigService,
	producer *ckafka.Producer,
) ITournamentService {
	return &TournamentService{
		db:                   db,
		tRepo:                tRepo,
		gRepo:                gRepo,
		tuRepo:               tuRepo,
		uRepo:                uRepo,
		tournamentRewardRepo: tournamentRewardRepo,
		dynamicConfigService: dynamicConfigService,
		producer:             producer,
	}
}

func (s *TournamentService) CreateDailyTournament(ctx context.Context) (*entity.Tournament, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	start := today
	end := start.Add(24*time.Hour - time.Second)

	tournament := &entity.Tournament{
		StartDate: start,
		EndDate:   end,
		Status:    entity.TournamentStatusPlanned,
	}

	err := s.tRepo.CreateTournament(ctx, tournament)
	if err != nil {
		return nil, err
	}

	return tournament, nil
}

func (s *TournamentService) StartTournament(ctx context.Context, id int64) error {
	tournament, err := s.tRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if tournament.HasEnded() {
		return domainErr.ErrTournamentAlreadyEnded
	}

	tournament.Activate()

	return s.tRepo.UpdateTournament(ctx, tournament)
}

func (s *TournamentService) CloseTournament(ctx context.Context, id int64) error {
	tournament, err := s.tRepo.FindByID(ctx, id)
	if err != nil {
		return domainErr.ErrTournamentNotFound
	}
	if tournament.Status == entity.TournamentStatusClosed {
		return domainErr.ErrTournamentAlreadyEnded
	}

	tournament.Close()

	return s.tRepo.UpdateTournament(ctx, tournament)
}

func (s *TournamentService) GetActiveTournament(ctx context.Context) (*entity.Tournament, error) {
	return s.tRepo.GetActiveTournament(ctx)
}

func (s *TournamentService) EnterTournamentAsync(ctx context.Context, userID int64) error {
	cutoffHour := s.dynamicConfigService.GetConfig().TournamentCutoffHour
	if time.Now().UTC().Hour() >= cutoffHour {
		return domainErr.ErrTournamentRegistrationClosed
	}

	user, err := s.uRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return domainErr.ErrUserNotFound
	}

	if user.Level < s.dynamicConfigService.GetConfig().MinimumTournamentEntryLevel {
		return domainErr.ErrLevelTooLowToEnterTournament
	}

	entranceCoins := int64(s.dynamicConfigService.GetConfig().TournamentEntranceCoins)
	if user.Coins < entranceCoins {
		return domainErr.ErrInsufficientCoins
	}

	payload := events.EnterTournamentPayload{UserID: userID}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = kafkautil.ProduceFireAndForget(s.producer, s.dynamicConfigService.GetConfig().TournamentEntryTopic, data)
	if err != nil {
		return err
	}

	log.GetLogger().Infof(
		"Fire-and-forget message for userID=%d has been produced to topic=%s (no delivery check).",
		userID, s.dynamicConfigService.GetConfig().TournamentEntryTopic,
	)

	return nil
}

func (s *TournamentService) EnterTournament(ctx context.Context, userID int64) error {

	tournament, err := s.tRepo.GetActiveTournament(ctx)
	if err != nil {
		return err
	}
	if tournament == nil {
		return domainErr.ErrNoActiveTournament
	}

	err = s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		user, err := s.uRepo.FindUserForUpdateTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return domainErr.ErrUserNotFound
		}

		tournamentEntranceCoins := int64(s.dynamicConfigService.GetConfig().TournamentEntranceCoins)

		user.Coins -= tournamentEntranceCoins
		if err := s.uRepo.UpdateUserTx(ctx, tx, user); err != nil {
			return err
		}

		lastGroup, err := s.gRepo.FindLastGroupForTournamentForUpdate(ctx, tx, tournament.ID)
		if err != nil {
			return err
		}

		var group *entity.Group
		if lastGroup == nil {
			group = &entity.Group{
				TournamentID: tournament.ID,
				GroupNumber:  1,
				CurrentSize:  0,
			}
			if err := s.gRepo.CreateGroupTx(ctx, tx, group); err != nil {
				return err
			}
		} else {
			if lastGroup.CurrentSize < 35 {
				group = lastGroup
			} else {
				group = &entity.Group{
					TournamentID: tournament.ID,
					GroupNumber:  lastGroup.GroupNumber + 1,
					CurrentSize:  0,
				}
				if err := s.gRepo.CreateGroupTx(ctx, tx, group); err != nil {
					return err
				}
			}
		}

		group.CurrentSize++
		if err := s.gRepo.UpdateGroupTx(ctx, tx, group); err != nil {
			return err
		}

		tu := entity.TournamentUser{
			TournamentID: tournament.ID,
			UserID:       user.ID,
			GroupID:      group.ID,
			Score:        0,
		}
		if err := s.tuRepo.CreateTournamentUserTx(ctx, tx, &tu); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.GetLogger().Info(fmt.Sprintf("User %d joined tournament %d in group ???", userID, tournament.ID))
	return nil
}

func (s *TournamentService) UpdateTournamentScore(ctx context.Context, message events.ProgressUpdateMessage) error {
	tournament, err := s.tRepo.GetActiveTournament(ctx)
	if err != nil {
		return domainErr.ErrInternalServerError
	}
	if tournament == nil {
		return nil
	}

	tournamentUser, err := s.tuRepo.GetTournamentUser(ctx, tournament.ID, message.UserID)
	if err != nil {
		return nil
	}
	if tournamentUser == nil {
		return nil
	}

	tournamentUser.Score++
	err = s.tuRepo.UpdateScore(ctx, tournamentUser)
	if err != nil {
		return domainErr.ErrInternalServerError
	}

	payload := events.LeaderboardUpdateMessage{
		UserID:       message.UserID,
		TournamentID: tournament.ID,
		Country:      message.Country,
		Score:        tournamentUser.Score,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = kafkautil.ProduceFireAndForget(s.producer, s.dynamicConfigService.GetConfig().LeaderboardUpdateTopic, data)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to send leaderboard update for user %d to Kafka: %v", message.UserID, err))
	}

	return nil
}
func (s *TournamentService) StoreRewards(ctx context.Context, tournamentID int64) error {
	tournamentUsers, err := s.tuRepo.GetTournamentUsersByTournament(ctx, tournamentID)
	if err != nil {
		return err
	}

	sort.Slice(tournamentUsers, func(i, j int) bool {
		return tournamentUsers[i].Score > tournamentUsers[j].Score
	})

	if len(tournamentUsers) > 10 {
		tournamentUsers = tournamentUsers[:10]
	}

	reward1 := s.dynamicConfigService.GetConfig().Reward1
	reward2 := s.dynamicConfigService.GetConfig().Reward2
	reward3 := s.dynamicConfigService.GetConfig().Reward3
	reward4to10 := s.dynamicConfigService.GetConfig().Reward4to10

	var rewards []entity.TournamentReward

	for i, tu := range tournamentUsers {
		rank := i + 1

		var coins int
		switch {
		case rank == 1:
			coins = reward1
		case rank == 2:
			coins = reward2
		case rank == 3:
			coins = reward3
		case rank >= 4 && rank <= 10:
			coins = reward4to10
		default:
			coins = 0
		}

		rw := entity.TournamentReward{
			TournamentID: tournamentID,
			UserID:       tu.UserID,
			Rank:         rank,
			RewardCoins:  coins,
			Claimed:      false,
		}
		rewards = append(rewards, rw)
	}

	if len(rewards) > 0 {
		err = s.tournamentRewardRepo.CreateRewards(ctx, rewards)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TournamentService) ClaimReward(ctx context.Context, userID int64) error {
	unclaimedList, err := s.tournamentRewardRepo.GetUnclaimedRewardsByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(unclaimedList) == 0 {
		return domainErr.ErrNoUnclaimedReward
	}

	var totalCoins int64 = 0
	for _, rw := range unclaimedList {
		totalCoins += int64(rw.RewardCoins)
	}

	user, err := s.uRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return domainErr.ErrUserNotFound
	}
	user.Coins += totalCoins

	err = s.uRepo.UpdateProgress(ctx, user)
	if err != nil {
		return err
	}

	for _, rw := range unclaimedList {
		if err := s.tournamentRewardRepo.ClaimReward(ctx, rw.ID); err != nil {
			return err
		}
	}

	return nil
}
