package handlers
import "github.com/gin-gonic/gin"
type DocumentCollaborationHandler struct{}
func NewDocumentCollaborationHandler() *DocumentCollaborationHandler { return &DocumentCollaborationHandler{} }
func (h *DocumentCollaborationHandler) Create(c *gin.Context) { c.JSON(200, gin.H{"status":"created"}) }
func (h *DocumentCollaborationHandler) Get(c *gin.Context) { c.JSON(200, gin.H{"status":"ok"}) }
func (h *DocumentCollaborationHandler) List(c *gin.Context) { c.JSON(200, gin.H{"status":"ok","data":[]string{}}) }
func (h *DocumentCollaborationHandler) Update(c *gin.Context) { c.JSON(200, gin.H{"status":"updated"}) }
func (h *DocumentCollaborationHandler) Delete(c *gin.Context) { c.JSON(200, gin.H{"status":"deleted"}) }
