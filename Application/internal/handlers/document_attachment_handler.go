package handlers

import "github.com/gin-gonic/gin"

type document_attachment_handler struct{}

func Newdocument_attachment_handler() *document_attachment_handler { return &document_attachment_handler{} }

func (h *document_attachment_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *document_attachment_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *document_attachment_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *document_attachment_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *document_attachment_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
