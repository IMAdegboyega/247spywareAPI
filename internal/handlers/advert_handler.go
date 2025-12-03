package handlers

import (
	"net/http"
	"strconv"

	"blog-backend/internal/models"
	"blog-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AdvertHandler struct {
	advertService *services.AdvertService
}

func NewAdvertHandler(advertService *services.AdvertService) *AdvertHandler {
	return &AdvertHandler{advertService: advertService}
}

func (h *AdvertHandler) Create(c *gin.Context) {
	var req models.CreateAdvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	advert, err := h.advertService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, advert)
}

func (h *AdvertHandler) GetAll(c *gin.Context) {
	adverts, err := h.advertService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, adverts)
}

func (h *AdvertHandler) GetActive(c *gin.Context) {
	position := c.Query("position")

	var adverts []models.Advert
	var err error

	if position != "" {
		adverts, err = h.advertService.GetByPosition(position)
	} else {
		adverts, err = h.advertService.GetActive()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, adverts)
}

func (h *AdvertHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid advert ID"})
		return
	}

	var req models.UpdateAdvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	advert, err := h.advertService.Update(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, advert)
}

func (h *AdvertHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid advert ID"})
		return
	}

	if err := h.advertService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "advert deleted successfully"})
}

func (h *AdvertHandler) RecordClick(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid advert ID"})
		return
	}

	if err := h.advertService.RecordClick(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "click recorded"})
}