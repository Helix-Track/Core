package handlers

import "github.com/gin-gonic/gin"

type document_version_handler struct{}

func Newdocument_version_handler() *document_version_handler { return &document_version_handler{} }

func (h *document_version_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *document_version_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *document_version_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *document_version_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *document_version_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
