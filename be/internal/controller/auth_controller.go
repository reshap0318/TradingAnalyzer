package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/dtos"
)

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dtos.LoginRequest true "Login credentials"
// @Success 200 {object} dtos.LoginResponse
// @Failure 400 {object} dtos.LoginResponse
// @Router /api/auth/login [post]
func (c *Controller) Login(ctx *gin.Context) {
	var req dtos.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.LoginResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	response, err := c.srvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.LoginResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusUnauthorized, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and invalidate session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.LogoutResponse
// @Failure 400 {object} dtos.LogoutResponse
// @Router /api/auth/logout [post]
func (c *Controller) Logout(ctx *gin.Context) {
	username, exists := ctx.Get("username")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dtos.LogoutResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	response, err := c.srvc.Logout(ctx, username.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.LogoutResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusUnauthorized, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Me returns current user info
// @Summary Get current user
// @Description Get current authenticated user info
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/me [get]
func (c *Controller) Me(ctx *gin.Context) {
	username, _ := ctx.Get("username")
	name, _ := ctx.Get("name")

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"username": username,
		"name":     name,
	})
}
