package handlers

import "github.com/gin-gonic/gin"

type document_template_handler struct{}

func Newdocument_template_handler() *document_template_handler { return &document_template_handler{} }

func (h *document_template_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *document_template_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *document_template_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *document_template_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *document_template_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
