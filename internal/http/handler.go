package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/nurpe/snowops-auth/internal/http/middleware"
	"github.com/nurpe/snowops-auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
	log         zerolog.Logger
}

func NewHandler(authService *service.AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		authService: authService,
		log:         log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")

	auth.POST("/login", h.login)
	auth.POST("/send-code", h.sendCode)
	auth.POST("/verify-code", h.verifyCode)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)
	auth.GET("/me", authMiddleware, h.me)
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		req.Login,
		req.Password,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) sendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	masked, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"masked_phone": masked,
	})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.VerifyCode(
		c.Request.Context(),
		req.Phone,
		req.Code,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.authService.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		metaFromContext(c),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) me(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse("missing token claims"))
		return
	}

	info, err := h.authService.GetMe(c.Request.Context(), claims.UserID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(info))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrCodeInvalid),
		errors.Is(err, service.ErrCodeExpired):
		c.JSON(http.StatusUnauthorized, errorResponse(err.Error()))
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	case errors.Is(err, service.ErrPermissionDenied),
		errors.Is(err, service.ErrHierarchyViolation):
		c.JSON(http.StatusForbidden, errorResponse(err.Error()))
	case errors.Is(err, service.ErrConflict):
		c.JSON(http.StatusConflict, errorResponse(err.Error()))
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func metaFromContext(c *gin.Context) service.AuthMeta {
	return service.AuthMeta{
		UserAgent: c.GetHeader("User-Agent"),
		ClientIP:  c.ClientIP(),
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}
