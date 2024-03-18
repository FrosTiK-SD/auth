package handler

import (
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/interfaces"
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
