package handlers

import "github.com/gin-gonic/gin"

type service_registry_handler struct{}

func Newservice_registry_handler() *service_registry_handler { return &service_registry_handler{} }

func (h *service_registry_handler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *service_registry_handler) Get(c *gin.Context)    { c.JSON(200, gin.H{"status":"ok"}) }
func (h *service_registry_handler) List(c *gin.Context)   { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *service_registry_handler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *service_registry_handler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
