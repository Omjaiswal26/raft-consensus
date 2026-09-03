package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *ClusterHandler) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/api/cluster", h.GetCluster)
	r.GET("/ws", h.ServeWS)
	r.POST("/api/command", h.SubmitCommand)

	return r
}
