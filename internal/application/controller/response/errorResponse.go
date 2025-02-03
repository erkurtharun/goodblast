package response

type ErrorResponse struct {
	Status      int    `json:"status"`
	Description string `json:"description"`
}
