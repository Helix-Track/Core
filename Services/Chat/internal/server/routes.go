package server

import (
	"github.com/gin-gonic/gin"
)

// RouteInfo contains route information
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	AuthRequired bool
}

// GetRoutes returns all registered routes
func (s *Server) GetRoutes() []RouteInfo {
	routes := []RouteInfo{
		{
			Method:       "GET",
			Path:         "/health",
			Handler:      "healthHandler",
			AuthRequired: false,
		},
		{
			Method:       "GET",
			Path:         "/version",
			Handler:      "versionHandler",
			AuthRequired: false,
		},
		{
			Method:       "POST",
			Path:         "/api/do",
			Handler:      "doHandler",
			AuthRequired: true,
		},
		{
			Method:       "GET",
			Path:         "/ws",
			Handler:      "websocketHandler",
			AuthRequired: true,
		},
	}

	return routes
}

// PrintRoutes logs all registered routes
func (s *Server) PrintRoutes() {
	routes := s.GetRoutes()
	for _, route := range routes {
		authStatus := "public"
		if route.AuthRequired {
			authStatus = "auth required"
		}
		gin.DefaultWriter.Write([]byte(
			route.Method + " " + route.Path + " -> " + route.Handler + " (" + authStatus + ")\n",
		))
	}
}
