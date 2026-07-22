package util

import (
	"time"

	"github.com/gin-contrib/cors"
)

func DefaultCors() cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Accept",
		"Authorization",
		"Cache-Control",
		"token",
		"id",
		"X-Hijack-Email",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.MaxAge = 12 * time.Hour

	return config
}
