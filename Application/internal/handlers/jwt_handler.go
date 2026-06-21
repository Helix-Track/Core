package handlers

import "github.com/gin-gonic/gin"

type jwt_handler struct{}

func Newjwt_handler() *jwt_handler { return &jwt_handler{} }

func (h *jwt_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *jwt_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *jwt_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *jwt_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *jwt_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
