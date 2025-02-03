package events

type ProgressUpdateMessage struct {
	UserID  int64  `json:"user_id"`
	Country string `json:"country"`
}
