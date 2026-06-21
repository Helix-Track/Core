package handlers

import "github.com/gin-gonic/gin"

type document_analytics_handler struct{}

func Newdocument_analytics_handler() *document_analytics_handler { return &document_analytics_handler{} }

func (h *document_analytics_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *document_analytics_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *document_analytics_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *document_analytics_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *document_analytics_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
