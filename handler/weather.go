package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetWeather godoc
// @Summary      Get today's New Taipei City weather
// @Tags         weather
// @Produce      json
// @Success      200  {object}  Response
// @Failure      500  {object}  Response
// @Router       /api/v1/weather [get]
func (h *Handler) GetWeather(c *gin.Context) {
	h.getTodayWeather(c)
}

// GetWeatherMe godoc
// @Summary      Get today's New Taipei City weather (protected)
// @Tags         weather
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  Response
// @Failure      401  {object}  Response
// @Failure      500  {object}  Response
// @Router       /api/v1/weather/me [get]
func (h *Handler) GetWeatherMe(c *gin.Context) {
	h.getTodayWeather(c)
}

func (h *Handler) getTodayWeather(c *gin.Context) {
	result, err := h.weatherSvc.GetTodayWeather(c.Request.Context())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get weather failed", "error", err)
		Fail(c, http.StatusInternalServerError, "internal server error", nil)
		return
	}
	OK(c, "success", result)
}
