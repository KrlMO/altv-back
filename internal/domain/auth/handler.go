package auth

import (
	"arena/internal/domain/auth/rules"
	"arena/internal/domain/users"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login/local", m.LoginLocal)
		auth.POST("/register/local", m.RegisterLocalAndLogin)
		auth.POST("/refresh", m.Refresh)
		auth.POST("/logout", m.Logout)
		auth.POST("/verify-email-verification", m.VerifyEmailVerification)
		auth.POST("/resend-verification", m.ResendEmailVerification)
	}
}

// LoginLocal godoc
// @Summary Login a user
// @Description Logs in a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User info"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/login/local [post]
func (m *Module) LoginLocal(c *gin.Context) {
	ctx := c.Request.Context()
	req := LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	responseData, refreshToken, err := m.service.LocalLogin(
		ctx,
		req.Email,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		if errors.Is(err, rules.ErrUserNotFound) || errors.Is(err, rules.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, response.ResponseError{
				Error: "invalid credentials",
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToGenerateAccessToken) || errors.Is(err, rules.ErrFailedToGenerateRefreshToken) {
			c.JSON(http.StatusInternalServerError, response.ResponseError{
				Error: "failed to generate tokens",
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToStoreRefreshToken) {
			c.JSON(http.StatusInternalServerError, response.ResponseError{
				Error: "failed to store refresh token",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "internal server error",
		})
		return
	}

	setRefreshCookie(c, refreshToken)

	c.JSON(http.StatusOK, response.ResponseOk[LoginResponse]{
		Data: responseData,
	})
}

// Refresh godoc
// @Summary Refresh user token
// @Description Refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.ResponseOk[LoginResponse]
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (m *Module) Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "missing refresh token",
		})
		return
	}

	reponse, newRefreshToken, err := m.service.Refresh(
		ctx,
		cookie,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		if errors.Is(err, rules.ErrInvalidRefreshToken) || errors.Is(err, rules.ErrRefreshTokenReuseDetected) || errors.Is(err, rules.ErrRefreshTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToGenerateAccessToken) || errors.Is(err, rules.ErrFailedToGenerateRefreshToken) || errors.Is(err, rules.ErrFailedToStoreRefreshToken) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to refresh tokens",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	setRefreshCookie(c, newRefreshToken)

	c.JSON(http.StatusOK, response.ResponseOk[LoginResponse]{
		Data: reponse,
	})
}

// Logout godoc
// @Summary Logout a user
// @Description Logs out a user by revoking their refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.ResponseOk[any]
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (m *Module) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		_ = m.service.Logout(ctx, cookie)
	}

	clearRefreshCookie(c)

	c.JSON(http.StatusOK, response.ResponseOk[any]{
		Msg: "logged out successfully",
	})
}

// RegisterLocalAndLogin godoc
// @Summary Register and login a user
// @Description Registers a user with local auth and logs them in
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body users.RegisterUserRequest true "User info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/register/local [post]
func (m *Module) RegisterLocalAndLogin(c *gin.Context) {
	ctx := c.Request.Context()
	req := users.RegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	if err := m.service.usersService.Register(c, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	if err := m.service.InitiateEmailVerification(ctx, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to verify email",
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[any]{
		Msg: "check your email to verify your account",
	})
}

// VerifyEmailVerification godoc
// @Summary Verify email verification
// @Description Verifies if the email verification token is valid and returns the status
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Email info"
// @Success 200 {object} response.ResponseOk[VerifyEmailVerificationResponse]
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/verify-email [post]
func (m *Module) VerifyEmailVerification(c *gin.Context) {
	ctx := c.Request.Context()
	req := VerifyEmailRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	responseData, err := m.service.VerifyEmailVerification(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[VerifyEmailVerificationResponse]{
		Data: *responseData,
	})
}

// ResendEmailVerification godoc
// @Summary Resend email verification
// @Description Resends the email verification token to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Email info"
// @Success 200 {object} response.ResponseOk[any]
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/resend-verification [post]
func (m *Module) ResendEmailVerification(c *gin.Context) {
	ctx := c.Request.Context()
	req := VerifyEmailRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	err := m.service.InitiateEmailVerification(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to resend verification email",
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[any]{
		Msg: "verification email resent successfully",
	})
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verifies the email using the token and marks the user as verified
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} response.ResponseOk[any]
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/verify-email [get]
func (m *Module) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.Query("token")

	if token == "" || len(token) < 20 {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid token",
		})
		return
	}

	err := m.service.VerifyEmail(ctx, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid or expired token",
		})
		return
	}

	c.Redirect(http.StatusAccepted, m.service.frontendURL+"/email-verified")
}
