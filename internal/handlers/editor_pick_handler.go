package handlers

import (
	"net/http"
	"strconv"

	"blog-backend/internal/models"
	"blog-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type EditorPickHandler struct {
	editorPickService *services.EditorPickService
}

func NewEditorPickHandler(editorPickService *services.EditorPickService) *EditorPickHandler {
	return &EditorPickHandler{editorPickService: editorPickService}
}

func (h *EditorPickHandler) Add(c *gin.Context) {
	var req models.AddEditorPickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pick, err := h.editorPickService.Add(req.PostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pick)
}

func (h *EditorPickHandler) GetAll(c *gin.Context) {
	picks, err := h.editorPickService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, picks)
}

func (h *EditorPickHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid editor pick ID"})
		return
	}

	if err := h.editorPickService.Remove(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "editor pick removed successfully"})
}

func (h *EditorPickHandler) Reorder(c *gin.Context) {
	var req models.ReorderEditorPicksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.editorPickService.Reorder(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "editor picks reordered successfully"})
}