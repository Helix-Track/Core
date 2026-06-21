package handlers

import "github.com/gin-gonic/gin"

type chat_handler struct{}

func Newchat_handler() *chat_handler { return &chat_handler{} }

func (h *chat_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *chat_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *chat_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *chat_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *chat_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
