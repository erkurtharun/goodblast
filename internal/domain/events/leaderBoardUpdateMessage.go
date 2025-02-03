package events

type LeaderboardUpdateMessage struct {
	UserID       int64  `json:"user_id"`
	TournamentID int64  `json:"tournament_id"`
	Country      string `json:"country"`
	Score        int    `json:"score"`
}
