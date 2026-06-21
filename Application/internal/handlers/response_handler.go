package handlers

import "github.com/gin-gonic/gin"

type response_handler struct{}

func Newresponse_handler() *response_handler { return &response_handler{} }

func (h *response_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *response_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *response_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *response_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *response_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
