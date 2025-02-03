package controller

import (
	"github.com/gin-gonic/gin"
	"goodblast/internal/application/controller/request"
	"goodblast/internal/application/controller/response"
	"goodblast/internal/application/service"
	"net/http"
)

type ITournamentController interface {
	CreateDailyTournament(ctx *gin.Context)
	StartTournament(ctx *gin.Context)
	CloseTournament(ctx *gin.Context)
	GetActiveTournament(ctx *gin.Context)
	EnterTournament(ctx *gin.Context)
	ClaimReward(ctx *gin.Context)
}

type TournamentController struct {
	service service.ITournamentService
}

func NewTournamentController(service service.ITournamentService) ITournamentController {
	return &TournamentController{
		service: service,
	}
}

// CreateDailyTournament godoc
// @Summary     Create a new daily tournament (00:00 - 23:59)
// @Description Creates a "planned" tournament for the current day.
// @Tags        Tournament
// @Accept      json
// @Produce     json
// @Success     200 {object} response.CreateDailyTournamentResponse
// @Failure     500 {object} map[string]string
// @Router      /internal/tournament/create-daily [post]
func (ctrl *TournamentController) CreateDailyTournament(ctx *gin.Context) {
	tournament, err := ctrl.service.CreateDailyTournament(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	resp := response.CreateDailyTournamentResponse{
		ID:        tournament.ID,
		Status:    string(tournament.Status),
		StartDate: tournament.StartDate.String(),
		EndDate:   tournament.EndDate.String(),
	}

	ctx.JSON(http.StatusOK, resp)
}

// StartTournament godoc
// @Summary     Start a tournament
// @Description Changes a tournament status from "planned" to "active" (if not ended).
// @Tags        Tournament
// @Accept      json
// @Produce     json
// @Param       requestBody body  request.StartTournamentReq true "Tournament ID to start"
// @Success     200   {object}  response.StartTournamentResponse
// @Failure     400   {object}  map[string]string "Bad Request"
// @Failure     409   {object}  map[string]string "Tournament Already Ended/Closed"
// @Failure     500   {object}  map[string]string "Internal Error"
// @Router      /internal/tournament/start [post]
func (ctrl *TournamentController) StartTournament(ctx *gin.Context) {
	var req request.StartTournamentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.service.StartTournament(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.StartTournamentResponse{
		Status: "tournament started",
	})
}

// CloseTournament godoc
// @Summary     Close a tournament
// @Description Moves a tournament status to "closed".
// @Tags        Tournament
// @Accept      json
// @Produce     json
// @Param       requestBody body  request.CloseTournamentReq true "Tournament ID to close"
// @Success     200   {object} response.CloseTournamentResponse
// @Failure     400   {object} map[string]string
// @Failure     409   {object} map[string]string "Tournament Already Ended/Closed"
// @Failure     500   {object} map[string]string
// @Router      /internal/tournament/close [post]
func (ctrl *TournamentController) CloseTournament(ctx *gin.Context) {
	var req request.CloseTournamentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.service.CloseTournament(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := ctrl.service.StoreRewards(ctx, req.ID); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, response.CloseTournamentResponse{Status: "tournament closed"})
}

// GetActiveTournament godoc
// @Summary     Get the currently active tournament
// @Description Returns the tournament that is currently marked "active" and within time range.
// @Tags        Tournament
// @Produce     json
// @Success     200 {object} response.GetActiveTournamentResponse
// @Failure     404 {object} map[string]string "No active tournament found"
// @Failure     500 {object} map[string]string
// @Router      /internal/tournament/active [get]
func (ctrl *TournamentController) GetActiveTournament(ctx *gin.Context) {
	tournament, err := ctrl.service.GetActiveTournament(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	if tournament == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active tournament"})
		return
	}

	resp := response.GetActiveTournamentResponse{
		ID:        tournament.ID,
		Status:    string(tournament.Status),
		StartDate: tournament.StartDate.String(),
		EndDate:   tournament.EndDate.String(),
	}
	ctx.JSON(http.StatusOK, resp)
}

// EnterTournament godoc
// @Summary     Enter the active tournament
// @Description Joins the currently active tournament if user meets level/coin requirements.
// @Tags        Tournament
// @Produce     json
// @Success     200 {object} map[string]string "Joined the tournament successfully"
// @Failure     401 {object} map[string]string "Unauthorized or invalid user ID"
// @Failure     403 {object} map[string]string "Forbidden if user level or coins are insufficient"
// @Failure     404 {object} map[string]string "No active tournament found"
// @Failure     409 {object} map[string]string "Tournament already ended (if needed)"
// @Failure     500 {object} map[string]string "Internal server error"
// @Security    BearerAuth
// @Router      /internal/tournament/enter [post]
func (ctrl *TournamentController) EnterTournament(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err := ctrl.service.EnterTournamentAsync(ctx.Request.Context(), userID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Joined the tournament successfully"})
}

// ClaimReward godoc
// @Summary     Claim all unclaimed rewards
// @Description Fetches all unclaimed rewards for the current user, sums their coin values, updates the user's coin balance, and marks each reward as claimed.
// @Tags        Tournament
// @Accept      json
// @Produce     json
// @Success     200 {object} map[string]string  "All rewards claimed successfully"
// @Failure     401 {object} map[string]string  "Unauthorized or invalid user ID"
// @Failure     404 {object} map[string]string  "No unclaimed rewards found"
// @Failure     500 {object} map[string]string  "Server error or database failure"
// @Security    BearerAuth
// @Router      /internal/tournament/reward/claim [post]
func (ctrl *TournamentController) ClaimReward(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err := ctrl.service.ClaimReward(ctx.Request.Context(), userID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "All rewards claimed"})
}
