package handlers

import "github.com/gin-gonic/gin"

type websocket_handler struct{}

func Newwebsocket_handler() *websocket_handler { return &websocket_handler{} }

func (h *websocket_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *websocket_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *websocket_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *websocket_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *websocket_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
