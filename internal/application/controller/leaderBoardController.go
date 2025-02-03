package controller

import (
	"github.com/gin-gonic/gin"
	"goodblast/internal/application/service"
	"net/http"
	"strconv"
)

type LeaderboardController struct {
	leaderboardService service.ILeaderboardService
}

func NewLeaderboardController(service service.ILeaderboardService) *LeaderboardController {
	return &LeaderboardController{leaderboardService: service}
}

// GetGlobalLeaderboard godoc
// @Summary     Get Global Leaderboard
// @Description Retrieves the top 1000 users from the global leaderboard.
// @Tags        Leaderboard
// @Produce     json
// @Success     200 {array} response.LeaderboardEntry
// @Failure     500 {object} map[string]string "Internal Server Error"
// @Router      /leaderboard/global [get]
func (c *LeaderboardController) GetGlobalLeaderboard(ctx *gin.Context) {
	leaderboard, err := c.leaderboardService.GetGlobalLeaderboard(1000)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, leaderboard)
}

// GetCountryLeaderboard godoc
// @Summary     Get Country Leaderboard
// @Description Retrieves the top 1000 users from the leaderboard for a specific country.
// @Tags        Leaderboard
// @Produce     json
// @Param       country path string true "Country Code (e.g., US, TR, DE)"
// @Success     200 {array} response.LeaderboardEntry
// @Failure     500 {object} map[string]string "Internal Server Error"
// @Router      /leaderboard/country/{country} [get]
func (c *LeaderboardController) GetCountryLeaderboard(ctx *gin.Context) {
	country := ctx.Param("country")
	leaderboard, err := c.leaderboardService.GetCountryLeaderboard(country, 1000)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, leaderboard)
}

// GetUserRank godoc
// @Summary     Get User Rank
// @Description Retrieves the global rank of a specific user.
// @Tags        Leaderboard
// @Produce     json
// @Param       userId path int true "User ID"
// @Success     200 {object} map[string]interface{} "User Rank"
// @Failure     404 {object} map[string]string "User Not Found"
// @Router      /leaderboard/user/{userId}/rank [get]
func (c *LeaderboardController) GetUserRank(ctx *gin.Context) {
	userID, _ := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	rank, err := c.leaderboardService.GetUserRank(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user_id": userID, "rank": rank})
}
