package handlers

import "github.com/gin-gonic/gin"

type user_role_handler struct{}

func Newuser_role_handler() *user_role_handler { return &user_role_handler{} }

func (h *user_role_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *user_role_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *user_role_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *user_role_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *user_role_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
