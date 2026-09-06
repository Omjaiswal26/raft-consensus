package api

import (
	"net/http"
	"raft-consensus/internal/response"
	"raft-consensus/internal/services"
	"strconv"
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

func (h *ClusterHandler) CrashNode(c *gin.Context) {
	nodeIDStr := c.Param("id")

	nodeID64, err := strconv.ParseUint(nodeIDStr, 10, 0)
	if err != nil {
		response.ErrorResponse(c, 400, "Invalid node id: "+err.Error())
		return
	}

	nodeID := uint(nodeID64)

	if err := h.service.CrashNode(nodeID); err != nil {
		response.ErrorResponse(c, 500, "Failed to crash node: "+err.Error())
		return
	}

	response.SuccessResponse(c, "Node crashed successfully", nil)
}

func (h *ClusterHandler) RecoverNode(c *gin.Context) {
	nodeIDStr := c.Param("id")

	nodeID64, err := strconv.ParseUint(nodeIDStr, 10, 0)
	if err != nil {
		response.ErrorResponse(c, 400, "Invalid node id: "+err.Error())
		return
	}

	nodeID := uint(nodeID64)

	if err := h.service.RecoverNode(nodeID); err != nil {
		response.ErrorResponse(c, 500, "Failed to recover node: "+err.Error())
		return
	}

	response.SuccessResponse(c, "Node recovered successfully", nil)
}
