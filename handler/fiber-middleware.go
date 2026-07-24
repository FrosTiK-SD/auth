package handler

import (
	"errors"
	"fmt"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// For Gin based middlewares
func (h *Handler) FiberVerifyStudent(ctx *fiber.Ctx) error {
	idToken := ctx.Get("token", "")
	noCache := false
	if ctx.Get("cache-control") == constants.NO_CACHE {
		noCache = true
	}

	email, _, err := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, noCache)

	if err != nil {
		return errors.New(*err)
	}
	student, err := controller.GetUserByEmail(h.MongikClient, email, &constants.ROLE_STUDENT, noCache)
	if err != nil {
		return errors.New(*err)
	}

	impersonateId := ctx.Get("x-impersonate-student-id", "")
	if impersonateId == "" {
		impersonateId = ctx.Get("X-Impersonate-Student-Id", "")
	}

	if impersonateId != "" {
		if util.CheckRoleExists(&student.GroupDetails, constants.ROLE_OPPORTUNITIES_WRITE) || util.CheckRoleExists(&student.GroupDetails, constants.ROLE_ADMIN) {
			targetObjId, parseErr := primitive.ObjectIDFromHex(impersonateId)
			if parseErr == nil {
				targetStudent, targetErr := controller.GetStudentById(h.MongikClient, targetObjId, noCache)
				if targetErr == nil && targetStudent != nil {
					h.LogActivityDirect(student.Id, "IMPERSONATION", fmt.Sprintf("Admin %s (%s) impersonating Student %s (%s)", student.FirstName, student.InstituteEmail, targetStudent.FirstName, targetStudent.InstituteEmail))
					student = targetStudent
				}
			}
		} else {
			return errors.New("Unauthorized impersonation attempt")
		}
	}

	ctx.Locals(constants.SESSION, student)
	ctx.Next()

	return nil
}

func (h *RoleCheckerHandler) FiberVerifyRole(ctx *fiber.Ctx) error {
	entity := ctx.Locals(constants.SESSION)
	var entityGroups *interfaces.Groups
	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	err = json.Unmarshal(entityBytes, &entityGroups)
	if err != nil {
		return err
	}
	if !util.CheckRoleExists(&entityGroups.Groups, h.Role) {
		return errors.New(constants.ERROR_ROLE_CHECK_FAILED)
	}

	ctx.Next()
	return nil
}
