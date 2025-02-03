package domain

import "net/http"

type CustomError struct {
	Message string
	Status  int
}

func (e *CustomError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound                 = &CustomError{"user not found", http.StatusNotFound}
	ErrInvalidPassword              = &CustomError{"invalid password", http.StatusUnauthorized}
	ErrUserAlreadyExists            = &CustomError{"user already exists", http.StatusConflict}
	ErrInternalServerError          = &CustomError{"internal server error", http.StatusInternalServerError}
	ErrTournamentNotFound           = &CustomError{"tournament not found", http.StatusNotFound}
	ErrTournamentAlreadyEnded       = &CustomError{"tournament has already ended or closed", http.StatusConflict}
	ErrNoActiveTournament           = &CustomError{"no active tournament", http.StatusNotFound}
	ErrLevelTooLowToEnterTournament = &CustomError{"level too low to enter", http.StatusForbidden}
	ErrInsufficientCoins            = &CustomError{"insufficient coins", http.StatusForbidden}
	ErrTournamentRegistrationClosed = &CustomError{"tournament registration closed", http.StatusConflict}
	ErrNoUnclaimedReward            = &CustomError{"no unclaimed reward", http.StatusNotFound}
)
