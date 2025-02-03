package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goodblast/internal/application/controller/request"
	"goodblast/internal/application/controller/response"
	"goodblast/internal/application/service"
	"goodblast/internal/validation"
	"goodblast/pkg/auth"
	"goodblast/pkg/log"
	"net/http"
)

type IUserController interface {
	CreateUser(ctx *gin.Context)
	Login(ctx *gin.Context)
	UpdateProgress(ctx *gin.Context)
}

type UserController struct {
	userService service.IUserService
	validator   validation.Validator
}

func NewUserController(userService service.IUserService, validator validation.Validator) IUserController {
	return &UserController{
		userService: userService,
		validator:   validator,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Tags User Controller
// @Accept json
// @Produce json
// @Param requestBody body request.CreateUserRequest true "User object that needs to be created"
// @Success 200 {object} int64
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /internal/user [post]
func (userController *UserController) CreateUser(ctx *gin.Context) {
	var req request.CreateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{Status: http.StatusBadRequest, Description: err.Error()})
		return
	}

	if err := userController.validator.Validate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Status:      http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	log.GetLogger().Info(fmt.Sprintf("UserController.CreateUser - Start Request: %#v", req))

	userId, err := userController.userService.CreateUser(ctx.Request.Context(), req)

	if err != nil {
		ctx.Error(err)
		return
	}

	log.GetLogger().Info(fmt.Sprintf("UserController.CreateUser - End Request userId: %#v", userId))

	ctx.JSON(http.StatusOK, userId)
}

// Login godoc
// @Summary Login user
// @Tags User Controller
// @Accept json
// @Produce json
// @Param requestBody body request.UserLoginRequest true "User object that needs to be created"
// @Success 200 {object} response.UserLoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /internal/user/login [post]
func (userController *UserController) Login(ctx *gin.Context) {
	var req request.UserLoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{Status: http.StatusBadRequest, Description: err.Error()})
		return
	}

	if err := userController.validator.Validate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Status:      http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	log.GetLogger().Info(fmt.Sprintf("UserController.Login - Start Request: %#v", req))

	user, err := userController.userService.Login(ctx.Request.Context(), req)

	if err != nil {
		ctx.Error(err)
		return
	}

	token, err := auth.GetAuth().GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{Status: http.StatusInternalServerError, Description: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response.UserLoginResponse{Token: token})
}

// UpdateProgress godoc
// @Summary Update user progress
// @Tags User Controller
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /internal/user/progress [post]
func (userController *UserController) UpdateProgress(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err := userController.userService.UpdateProgress(ctx.Request.Context(), userIDInt64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
