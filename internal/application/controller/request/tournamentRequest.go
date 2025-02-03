package request

type CreateDailyTournamentRequest struct {
}

type StartTournamentReq struct {
	ID int64 `json:"id" binding:"required"`
}

type CloseTournamentReq struct {
	ID int64 `json:"id" binding:"required"`
}
