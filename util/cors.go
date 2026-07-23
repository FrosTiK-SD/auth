package util

import (
	"github.com/gin-contrib/cors"
)

func DefaultCors() cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Content-Type,access-control-allow-origin, access-control-allow-headers")
	config.AllowHeaders = append(config.AllowHeaders, "token", "id")

	return config
}
