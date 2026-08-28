package response

import (
	"net/http"
	"raft-consensus/internal/dto"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func BadResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusBadRequest, dto.Response{
		Success: false,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, dto.Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}
