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

	token := r.Group("/api/token")
	{
		token.GET("/verify", handler.HandlerVerifyRecruiterIdToken)
		token.GET("/student/verify", handler.HandlerVerifyStudentIdToken)
		token.GET("/invalidate_cache", handler.InvalidateCache)
	}

	student := r.Group("/api/student")
	{
		student.GET("/", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_ADMIN), handler.GetAllStudents)
		student.GET("/id", handler.GinVerifyStudent, handler.GetStudentById)
		student.GET("/tpr/all", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_ADMIN), handler.GetAllTprs)
		student.GET("/tprLogin", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_TPR), handler.HandlerTprLogin)
		student.PUT("/update", handler.GinVerifyStudent, handler.HandlerUpdateStudentDetails)
		student.POST("/register", handler.HandlerRegisterStudentDetails)
	}

	group := r.Group("/api/group")
	{
		group.GET("/", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_GROUP_READ), handler.GetAllGroups)
		group.POST("/batch", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_GROUP_CREATE), handler.BatchCreateGroup)
		group.PUT("/batch/edit", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_GROUP_EDIT), handler.BatchEditGroup)
		group.DELETE("/batch/delete", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_GROUP_DELETE), handler.BatchDeleteGroup)
		group.POST("/batch/assign", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_GROUP_ASSIGN), handler.BatchAssignGroup)
	}

	domain := r.Group("/api/domain")
	{
		domain.GET("/", handler.GinVerifyStudent, handler.GetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_ALL_READ), handler.GetAllDomains)
		domain.GET("/id")
		domain.DELETE("/id")
	}

	port := "" + os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
