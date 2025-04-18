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
		FallbackToDefault: true,
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
		student.GET("", handler.GinVerifyStudent, handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_ADMIN), handler.GetAllStudents)
		student.GET("/id", handler.GinVerifyStudent, handler.GetStudentById)
		student.GET("/tpr/all", handler.GinVerifyStudent, handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_ADMIN), handler.GetAllTprs)
		student.GET("/tprLogin", handler.GinVerifyStudent, handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_TPR), handler.HandlerTprLogin)
		student.PUT("/update", handler.GinVerifyStudent, handler.HandlerUpdateStudentDetails)

		student.GET("/profile", handler.GinVerifyStudent, handler.HandlerGetStudentProfile)
		student.PUT("/profile", handler.GinVerifyStudent, handler.HandlerUpdateStudentProfile)
		student.POST("/register", handler.HandlerRegisterStudentDetails)
	}

	group := r.Group("/api/group", handler.GinVerifyStudent)
	{
		group.GET("", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_GROUP_READ), handler.GetAllGroups)
		group.POST("/batch", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_GROUP_CREATE), handler.BatchCreateGroup)
		group.PUT("/batch/edit", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_GROUP_EDIT), handler.BatchEditGroup)
		group.DELETE("/batch/delete", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_GROUP_DELETE), handler.BatchDeleteGroup)
		group.POST("/batch/assign", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_GROUP_ASSIGN), handler.BatchAssignGroup)
	}

	domain := r.Group("/api/domain", handler.GinVerifyStudent)
	{
		domain.GET("", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_ALL_READ), handler.GetAllDomains)
		domain.GET("/id", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_ALL_READ), handler.GetDomainById)
		domain.POST("/batch", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_CREATE), handler.BatchCreateDomain)
		domain.PUT("/id", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_EDIT), handler.EditDomainById)
		domain.DELETE("/id", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_DOMAIN_DELETE), handler.DeleteDomainById)
	}

	companies := r.Group("/api/company", handler.GinVerifyStudent)
	{
		companies.GET("/all", handler.GinGetRoleCheckHandlerForStudent(constants.ROLE_COMPANY_ALL_READ), handler.GetAllCompanies)
	}

	port := "" + os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
