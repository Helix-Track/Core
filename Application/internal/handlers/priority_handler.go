package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type PriorityHandler struct{}

func NewPriorityHandler() *PriorityHandler {
	return &PriorityHandler{}
}

func (h *PriorityHandler) Create(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "created"}) }
func (h *PriorityHandler) Get(c *gin.Context)    { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (h *PriorityHandler) List(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (h *PriorityHandler) Update(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"status": "updated"}) }
func (h *PriorityHandler) Delete(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"status": "deleted"}) }
