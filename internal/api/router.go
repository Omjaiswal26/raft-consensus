package api

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(h *ClusterHandler) *gin.Engine {
	r := gin.Default()
	r.GET("/api/cluster", h.GetCluster)
	r.GET("/ws", h.ServeWS)
	return r
}