package handlers

import "github.com/gin-gonic/gin"

type request_handler struct{}

func Newrequest_handler() *request_handler { return &request_handler{} }

func (h *request_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *request_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *request_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *request_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *request_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
