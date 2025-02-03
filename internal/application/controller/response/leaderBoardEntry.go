package response

type LeaderboardEntry struct {
	UserID int64 `json:"user_id"`
	Score  int64 `json:"score"`
	Rank   int64 `json:"rank"`
}
