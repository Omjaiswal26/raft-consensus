package api 

import (
	"net/http"
	"raft-consensus/internal/services"
	"github.com/gin-gonic/gin"
)

type ClusterHandler struct {
	service *services.ClusterService
}

func NewClusterHandler(service *services.ClusterService) *ClusterHandler {
	return &ClusterHandler{service: service}
}

func (h *ClusterHandler) GetCluster(c *gin.Context) {
	nodesSnapshot := h.service.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"nodes": nodesSnapshot,
	})
}