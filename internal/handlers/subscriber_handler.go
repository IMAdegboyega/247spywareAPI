package handlers

import (
	"net/http"
	"strconv"

	"blog-backend/internal/models"
	"blog-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SubscriberHandler struct {
	subscriberService *services.SubscriberService
}

func NewSubscriberHandler(subscriberService *services.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{subscriberService: subscriberService}
}

func (h *SubscriberHandler) Subscribe(c *gin.Context) {
	var req models.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriber, err := h.subscriberService.Subscribe(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "successfully subscribed to newsletter",
		"email":   subscriber.Email,
	})
}

func (h *SubscriberHandler) Unsubscribe(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsubscribe token required"})
		return
	}

	if err := h.subscriberService.Unsubscribe(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully unsubscribed from newsletter"})
}

func (h *SubscriberHandler) GetAll(c *gin.Context) {
	subscribers, err := h.subscriberService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscribers)
}

func (h *SubscriberHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscriber ID"})
		return
	}

	if err := h.subscriberService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscriber deleted successfully"})
}