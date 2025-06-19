package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// NormalizeTrailingSlash is a Gin middleware that normalizes request paths
// by removing a trailing slash if present, to ensure consistent routing.
// This allows you to register routes without trailing slashes (e.g., "/todos")
// and requests to both "/todos" and "/todos/" will be handled by the same route
// without Gin's default 301 redirect.
func NormalizeTrailingSlash() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.URL.Path) > 1 && strings.HasSuffix(c.Request.URL.Path, "/") {
			// Remove trailing slash
			c.Request.URL.Path = c.Request.URL.Path[:len(c.Request.URL.Path)-1]
		}
		// Continue to the next handler (Gin's router)
		c.Next()
	}
} 
