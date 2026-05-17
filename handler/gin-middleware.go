package handler

import (
	"fmt"
	"net/http"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gin-gonic/gin"
)

// For Gin based middlewares
func (h *Handler) GinVerifyStudent(ctx *gin.Context) {

	// Create a new session
	currentHandler := Handler{
		MongikClient: h.MongikClient,
		JwkSet:       h.JwkSet,
		Session:      &Session{},
		Config: Config{
			Mode: MIDDLEWARE,
		},
	}

	currentHandler.HandlerVerifyStudentIdToken(ctx)
	student := currentHandler.Session.Student

	if student != nil {
		ctx.Set(constants.SESSION, student)
		ctx.Next()
	} else {
		ctx.Abort()
	}
}

// To be Used only after GinVerifyStudent
func (h *Handler) GinGetRoleCheckHandlerForStudent(roles ...string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		value, exists := ctx.Get(constants.SESSION)
		student, ok := value.(*model.StudentPopulated)

		if !exists || !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": constants.ERROR_ROLE_CHECK_FAILED,
				"error":   "Student does not exist",
			})
			return
		}
		for _, role := range roles {
			if !util.CheckRoleExists(&student.GroupDetails, role) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"message": constants.ERROR_ROLE_CHECK_FAILED,
					"error":   fmt.Sprintf("Student Does not have Role '%s'", role),
				})
				return
			}
		}
	}
}

func (h *RoleCheckerHandler) GinVerifyRole(ctx *gin.Context) {
	entity, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(200, gin.H{
			"message": constants.ERROR_ROLE_CHECK_FAILED,
			"error":   "Entity does not exist",
		})
		return
	}
	var entityGroups *interfaces.Groups
	entityBytes, err := json.Marshal(entity)
	if err != nil {
		ctx.AbortWithStatusJSON(200, gin.H{
			"message": constants.ERROR_ROLE_CHECK_FAILED,
			"error":   err,
		})
		return
	}
	err = json.Unmarshal(entityBytes, &entityGroups)
	if err != nil {
		ctx.AbortWithStatusJSON(200, gin.H{
			"message": constants.ERROR_ROLE_CHECK_FAILED,
			"error":   err,
		})
		return
	}
	if !util.CheckRoleExists(&entityGroups.Groups, h.Role) {
		ctx.AbortWithStatusJSON(200, gin.H{
			"message": constants.ERROR_ROLE_CHECK_FAILED,
			"error":   "Role does not exist",
		})
		return
	}

	ctx.Next()
}
