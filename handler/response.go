package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Error     interface{} `json:"error"`
	Code      int         `json:"code"`
	Timestamp string      `json:"timestamp"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Error:     nil,
		Code:      http.StatusOK,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func Fail(c *gin.Context, httpStatus int, message string, detail interface{}) {
	c.JSON(httpStatus, Response{
		Success:   false,
		Message:   message,
		Data:      nil,
		Error:     detail,
		Code:      httpStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
