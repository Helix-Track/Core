package handlers

import "github.com/gin-gonic/gin"

type user_handler struct{}

func Newuser_handler() *user_handler { return &user_handler{} }

func (h *user_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *user_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *user_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *user_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *user_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
