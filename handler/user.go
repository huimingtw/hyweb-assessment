package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingz/hyweb-assessment/model"
	"github.com/huimingz/hyweb-assessment/service"
)

// ChangePassword godoc
// @Summary      Change authenticated user's password
// @Tags         user
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      model.ChangePasswordRequest  true  "Change password payload"
// @Success      200   {object}  Response
// @Failure      400   {object}  Response
// @Failure      401   {object}  Response
// @Failure      404   {object}  Response
// @Failure      500   {object}  Response
// @Router       /api/v1/user/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	email := c.GetString("email")

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	err := h.userSvc.ChangePassword(c.Request.Context(), email, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWrongPassword):
			Fail(c, http.StatusBadRequest, "old password is incorrect", nil)
		case errors.Is(err, service.ErrUserNotFound):
			Fail(c, http.StatusNotFound, "user not found", nil)
		default:
			h.logger.ErrorContext(c.Request.Context(), "change password failed", "error", err, "email", email)
			Fail(c, http.StatusInternalServerError, "internal server error", nil)
		}
		return
	}

	OK(c, "password changed successfully", nil)
}
