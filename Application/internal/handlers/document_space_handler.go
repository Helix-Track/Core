package handlers

import "github.com/gin-gonic/gin"

type document_space_handler struct{}

func Newdocument_space_handler() *document_space_handler { return &document_space_handler{} }

func (h *document_space_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *document_space_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *document_space_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *document_space_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *document_space_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
