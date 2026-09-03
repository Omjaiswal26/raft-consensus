package api

import (
	"net/http"
	"raft-consensus/internal/response"
	"raft-consensus/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ClusterHandler struct {
	service *services.ClusterService
	hub     *Hub
}

func NewClusterHandler(service *services.ClusterService, hub *Hub) *ClusterHandler {
	return &ClusterHandler{service: service, hub: hub}
}

func (h *ClusterHandler) GetCluster(c *gin.Context) {
	nodesSnapshot := h.service.Snapshot()
	response.SuccessResponse(c, "Cluster fetched successfully", gin.H{"nodes": nodesSnapshot})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *ClusterHandler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := h.hub.Register(conn)
	defer h.hub.Unregister(client)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-client.done:
			return
		case <-ticker.C:
			snapshot := h.service.Snapshot()
			if !h.hub.Send(client, gin.H{"type": "cluster", "nodes": snapshot}) {
				return
			}
		}
	}
}

func (h *ClusterHandler) SubmitCommand(c *gin.Context) {
	var body struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadResponse(c, err.Error(), nil)
		return
	}

	if err := h.service.SubmitCommand(body.Command); err != nil {
		response.ErrorResponse(c, http.StatusConflict, err.Error())
		return
	}

	response.SuccessResponse(c, "Command submitted successfully", nil)
}
