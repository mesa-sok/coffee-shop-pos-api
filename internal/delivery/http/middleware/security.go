package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// maxBodyBytes is the maximum allowed size of a request body (1 MB).
const maxBodyBytes = 1 << 20

// BodySizeLimit returns a middleware that caps incoming request bodies at 1 MB.
// Requests with bodies exceeding this limit are rejected with 413 Request Entity Too Large,
// preventing memory-exhaustion denial-of-service attacks.
func BodySizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	}
}

// SecurityHeaders returns a middleware that sets standard security response headers on
// every HTTP response to harden the API against common web vulnerabilities.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
