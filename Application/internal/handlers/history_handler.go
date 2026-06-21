package handlers

import "github.com/gin-gonic/gin"

type history_handler struct{}

func Newhistory_handler() *history_handler { return &history_handler{} }

func (h *history_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *history_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *history_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *history_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *history_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
