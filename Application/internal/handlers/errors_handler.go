package handlers

import "github.com/gin-gonic/gin"

type errors_handler struct{}

func Newerrors_handler() *errors_handler { return &errors_handler{} }

func (h *errors_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *errors_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *errors_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *errors_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *errors_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
