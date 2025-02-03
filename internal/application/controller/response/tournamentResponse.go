package response

type CreateDailyTournamentResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type GetActiveTournamentResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type CloseTournamentResponse struct {
	Status string `json:"status"`
}

type StartTournamentResponse struct {
	Status string `json:"status"`
}
