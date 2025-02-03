package service

import (
	"context"
	"encoding/json"
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	appconfig "goodblast/config"
	"goodblast/internal/application/controller/request"
	"goodblast/internal/application/repository"
	"goodblast/internal/domain/entity"
	domain "goodblast/internal/domain/errors"
	"goodblast/internal/domain/events"
	kafkautil "goodblast/internal/infrastructure/kafka/producer"
	"goodblast/pkg/log"
)

type IUserService interface {
	CreateUser(ctx context.Context, userRequest request.CreateUserRequest) (*int64, error)
	Login(ctx context.Context, userRequest request.UserLoginRequest) (*entity.User, error)
	UpdateProgress(ctx context.Context, userID int64) error
}

type UserService struct {
	userRepository           repository.IUserRepository
	tournamentRepository     repository.ITournamentRepository
	tournamentUserRepository repository.ITournamentUserRepository
	dynamicConfigService     appconfig.IDynamicConfigService
	producer                 *ckafka.Producer
}

func NewUserService(
	userRepository repository.IUserRepository,
	tournamentRepository repository.ITournamentRepository,
	tournamentUserRepository repository.ITournamentUserRepository,
	dynamicConfigService appconfig.IDynamicConfigService,
	producer *ckafka.Producer,

) IUserService {
	return &UserService{
		userRepository:           userRepository,
		tournamentRepository:     tournamentRepository,
		tournamentUserRepository: tournamentUserRepository,
		dynamicConfigService:     dynamicConfigService,
		producer:                 producer,
	}
}

func (usrServ *UserService) CreateUser(ctx context.Context, userRequest request.CreateUserRequest) (*int64, error) {
	user := entity.NewUserFromRequest(userRequest)

	userId, err := usrServ.userRepository.CreateUser(ctx, &user)

	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("UserService.CreateUser - Error: %v", err.Error()))
		return nil, domain.ErrUserAlreadyExists
	}

	return &userId, nil
}

func (usrServ *UserService) Login(ctx context.Context, userRequest request.UserLoginRequest) (*entity.User, error) {
	user, err := usrServ.userRepository.GetUserByUsername(ctx, userRequest.Username)

	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("UserService.Login - Error: %v", err.Error()))
		return nil, domain.ErrUserNotFound
	}

	if err := user.CheckPassword(userRequest.Password); err != nil {
		log.GetLogger().Error(fmt.Sprintf("UserService.Login - Error: %v, userId: %v", err.Error(), user.ID))
		return nil, domain.ErrInvalidPassword
	}

	return user, nil
}

func (usrServ *UserService) UpdateProgress(ctx context.Context, userID int64) error {
	user, err := usrServ.userRepository.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return domain.ErrUserNotFound
	}

	coinPerLevel := usrServ.dynamicConfigService.GetConfig().CoinPerLevel
	user.IncrementLevel(coinPerLevel)
	err = usrServ.userRepository.UpdateProgress(ctx, user)
	if err != nil {
		return domain.ErrInternalServerError
	}

	payload := events.ProgressUpdateMessage{UserID: userID, Country: user.Country}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = kafkautil.ProduceFireAndForget(usrServ.producer, usrServ.dynamicConfigService.GetConfig().UserProgressUpdateTopic, data)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Failed to send progress update for user %d to Kafka: %v", user.ID, err))
	}

	return nil
}
