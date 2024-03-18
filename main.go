package main

import (
	"fmt"
	"os"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/handler"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/FrosTiK-SD/mongik"
	mongikConstants "github.com/FrosTiK-SD/mongik/constants"
	mongikModels "github.com/FrosTiK-SD/mongik/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	r := gin.Default()

	mongikClient := mongik.NewClient(os.Getenv(constants.CONNECTION_STRING), &mongikModels.Config{
		Client: mongikConstants.REDIS,
		TTL:    constants.CACHING_DURATION,
		RedisConfig: &mongikModels.RedisConfig{
			URI:      os.Getenv(constants.REDIS_URI),
			Password: os.Getenv(constants.REDIS_PASSWORD),
			Username: os.Getenv(constants.REDIS_USERNAME),
		},
	})

	// Initialie default JWKs
	defaultJwkSet, jwkSetRetrieveError := controller.GetJWKs(mongikClient.CacheClient, true)
	if jwkSetRetrieveError != nil {
		fmt.Println("Error retrieving JWKs")
	}

	r.Use(cors.New(util.DefaultCors()))

	handler := &handler.Handler{
		MongikClient: mongikClient,
		JwkSet:       defaultJwkSet,
		Config: handler.Config{
			Mode: handler.HANDLER,
		},
		Session: &handler.Session{},
	}

	r.GET("/api/token/student/verify", handler.HandlerVerifyStudentIdToken)
	r.POST("/api/student/update", handler.GinVerifyStudent, handler.HandlerUpdateStudentDetails)
	r.POST("/api/student/register", handler.HandlerRegisterStudentDetails)
	r.GET("/api/token/invalidate_cache", handler.InvalidateCache)

	port := "" + os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
