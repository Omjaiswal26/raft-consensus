package api

import (
	"net/http"
	"raft-consensus/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {return true},
}

func (h *ClusterHandler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// v1: push snapshot every 200-500ms
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		snapshot := h.service.Snapshot()
		if err := conn.WriteJSON(gin.H{"type": "cluster", "nodes": snapshot}); err != nil {
			return
		}
	}
}