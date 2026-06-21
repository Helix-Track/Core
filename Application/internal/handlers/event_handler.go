package handlers

import "github.com/gin-gonic/gin"

type event_handler struct{}

func Newevent_handler() *event_handler { return &event_handler{} }

func (h *event_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *event_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *event_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *event_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *event_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
